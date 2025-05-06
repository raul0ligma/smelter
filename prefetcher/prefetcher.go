package prefetcher

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/raul0ligma/smelter/entity"
	"github.com/raul0ligma/smelter/provider"
)

type forkDB interface {
	Config() entity.ForkConfig
	LoadSlots(ctx context.Context, slots entity.Slots)
	SetCode(ctx context.Context, addr common.Address, code []byte)
	CreateStateWithValues(addr common.Address, nonce uint64, bal *big.Int, code []byte)
}

type TxHandlerFunc func(ctx context.Context, tx ethereum.CallMsg, rpc entity.BatchedRpc, config entity.ForkConfig) ([]entity.BatchReq, error)

type ResponseHandlerFunc func(ctx context.Context, tx ethereum.CallMsg, rpc entity.BatchedRpc, db forkDB, requests []entity.BatchReq, responses []json.RawMessage) error

type Prefetcher struct {
	rpc                entity.BatchedRpc
	db                 forkDB
	signatureHandlers  map[string]TxHandlerFunc
	responseHandlers   map[string]ResponseHandlerFunc
	defaultRespHandler ResponseHandlerFunc
}

func NewPrefetcher(rpc entity.BatchedRpc, db forkDB) *Prefetcher {
	p := &Prefetcher{
		rpc:               rpc,
		db:                db,
		signatureHandlers: make(map[string]TxHandlerFunc),
		responseHandlers:  make(map[string]ResponseHandlerFunc),
	}

	p.defaultRespHandler = defaultResponseHandler

	p.registerDefaultHandlers()

	return p
}

func (p *Prefetcher) RegisterSignatureHandler(signature string, reqHandler TxHandlerFunc, respHandler ResponseHandlerFunc) {
	p.signatureHandlers[signature] = reqHandler
	if respHandler != nil {
		p.responseHandlers[signature] = respHandler
	}
}

func (p *Prefetcher) registerDefaultHandlers() {
	p.RegisterSignatureHandler("0x6a761202", handleGnosisSafeExecTransaction, handleGnosisSafeResponses)
}

func (p *Prefetcher) AnalyzeTxAndPrefetch(ctx context.Context, tx ethereum.CallMsg) error {
	if !p.rpc.SupportsBatching() || tx.To == nil || len(tx.Data) < 4 {
		return nil
	}

	signature := "0x" + common.Bytes2Hex(tx.Data[:4])
	reqHandler, exists := p.signatureHandlers[signature]
	if !exists {
		return nil
	}

	handlerReqs, err := reqHandler(ctx, tx, p.rpc, p.db.Config())
	if err != nil || len(handlerReqs) == 0 {
		return err
	}

	var batchReqs []entity.BatchReq
	batchReqs = append(batchReqs,
		entity.BatchReq{Method: provider.MethodCodeAt, Params: []any{tx.To.Hex(), p.db.Config().ForkBlock}},
		entity.BatchReq{Method: provider.MethodBalanceAt, Params: []any{tx.To.Hex(), p.db.Config().ForkBlock}},
		entity.BatchReq{Method: provider.MethodNonceAt, Params: []any{tx.To.Hex(), p.db.Config().ForkBlock}},
	)
	batchReqs = append(batchReqs, handlerReqs...)

	responses, err := p.rpc.Batch(ctx, batchReqs)
	if err != nil {
		return err
	}

	if len(responses) >= 3 {
		createStateFromResponses(p.db, *tx.To, responses[0], responses[1], responses[2])
	}

	handlerResps := responses[3:]
	handlerReqsOnly := handlerReqs

	respHandler, exists := p.responseHandlers[signature]
	if !exists {
		respHandler = p.defaultRespHandler
	}

	return respHandler(ctx, tx, p.rpc, p.db, handlerReqsOnly, handlerResps)
}

func createStateFromResponses(db forkDB, addr common.Address, codeResp, balanceResp, nonceResp json.RawMessage) {
	var code hexutil.Bytes
	var balance hexutil.Big
	var nonce hexutil.Uint64

	bal := big.NewInt(0)
	var nonceVal uint64 = 0
	var codeBytes []byte

	if err := json.Unmarshal(codeResp, &code); err == nil {
		codeBytes = code
	}

	if err := json.Unmarshal(balanceResp, &balance); err == nil {
		bal = (*big.Int)(&balance)
	}

	if err := json.Unmarshal(nonceResp, &nonce); err == nil {
		nonceVal = uint64(nonce)
	}

	db.CreateStateWithValues(addr, nonceVal, bal, codeBytes)
}

func defaultResponseHandler(ctx context.Context, tx ethereum.CallMsg, rpc entity.BatchedRpc, db forkDB, requests []entity.BatchReq, responses []json.RawMessage) error {
	var slots entity.Slots

	for i, resp := range responses {
		if i >= len(requests) {
			break
		}

		req := requests[i]
		if req.Method != provider.MethodGetStorageAt || len(req.Params) < 2 {
			continue
		}

		addrStr, ok := req.Params[0].(string)
		if !ok {
			continue
		}

		keyStr, ok := req.Params[1].(string)
		if !ok {
			continue
		}

		addr := common.HexToAddress(addrStr)
		key := common.HexToHash(keyStr)

		var hexValue string
		if err := json.Unmarshal(resp, &hexValue); err != nil {
			continue
		}
		value := common.HexToHash(hexValue)

		slots = append(slots, entity.Slot{
			Key:   key,
			Addr:  addr,
			Value: value[:],
		})
	}

	if len(slots) > 0 {
		db.LoadSlots(ctx, slots)
	}

	return nil
}

func handleGnosisSafeExecTransaction(ctx context.Context, tx ethereum.CallMsg, rpc entity.BatchedRpc, forkConfig entity.ForkConfig) ([]entity.BatchReq, error) {
	if tx.To == nil {
		return nil, nil
	}

	keys := []common.Hash{
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"), // implementation singleton
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000005"), // nonce
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"), // threshold
		common.HexToHash("0x4a204f620c8c5ccdca3fd54d003badd85ba500436a431f0cbda4f558c93c34c8"), // guard
	}

	var requests []entity.BatchReq
	for _, key := range keys {
		requests = append(requests, entity.BatchReq{
			Method: provider.MethodGetStorageAt,
			Params: []any{tx.To.Hex(), key.Hex(), forkConfig.ForkBlock},
		})
	}

	return requests, nil
}

func handleGnosisSafeResponses(ctx context.Context, tx ethereum.CallMsg, rpc entity.BatchedRpc, db forkDB, requests []entity.BatchReq, responses []json.RawMessage) error {
	var slots entity.Slots
	var implAddr common.Address

	for i, resp := range responses {
		if i >= len(requests) {
			break
		}

		req := requests[i]
		if req.Method != provider.MethodGetStorageAt || len(req.Params) < 2 {
			continue
		}

		addrStr, ok := req.Params[0].(string)
		if !ok {
			continue
		}

		keyStr, ok := req.Params[1].(string)
		if !ok {
			continue
		}

		addr := common.HexToAddress(addrStr)
		key := common.HexToHash(keyStr)

		var hexValue string
		if err := json.Unmarshal(resp, &hexValue); err != nil {
			continue
		}
		value := common.HexToHash(hexValue)

		slots = append(slots, entity.Slot{
			Key:   key,
			Addr:  addr,
			Value: value[:],
		})

		// we would want to load the implementation address as well
		if key == common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000") {
			implAddr = common.BytesToAddress(value.Bytes())
		}
	}

	if len(slots) > 0 {
		db.LoadSlots(ctx, slots)
	}

	if implAddr != common.HexToAddress("") && implAddr != *tx.To {
		stateReqs := []entity.BatchReq{
			{Method: provider.MethodCodeAt, Params: []any{implAddr.Hex(), db.Config().ForkBlock}},
			{Method: provider.MethodBalanceAt, Params: []any{implAddr.Hex(), db.Config().ForkBlock}},
			{Method: provider.MethodNonceAt, Params: []any{implAddr.Hex(), db.Config().ForkBlock}},
		}

		stateResps, err := rpc.Batch(ctx, stateReqs)
		if err != nil || len(stateResps) < 3 {
			return err
		}

		createStateFromResponses(db, implAddr, stateResps[0], stateResps[1], stateResps[2])
	}

	return nil
}
