package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/pkg/server"
	"github.com/rahul0tripathi/smelter/tracer"
)

type Rpc struct {
	executor executor
	forkDB   forkDB

	reader entity.ChainStateReader
	cfg    *entity.ForkConfig
}

func NewRpcService(
	executor executor,
	forkDB forkDB,
	cfg *entity.ForkConfig,
	reader entity.ChainStateReader,
) *Rpc {
	return &Rpc{
		executor: executor,
		forkDB:   forkDB,
		cfg:      cfg,
		reader:   reader,
	}
}

func (r *Rpc) ChainId(_ context.Context) string {
	return hexutil.Encode(new(big.Int).SetUint64(r.cfg.ChainID).Bytes())
}

func (r *Rpc) BlockNumber(_ context.Context) (string, error) {
	_, blockNum := r.executor.Latest()
	if blockNum == 0 {
		return hexutil.Encode(r.cfg.ForkBlock.Bytes()), nil
	}

	return hexutil.Encode(new(big.Int).SetUint64(blockNum).Bytes()), nil
}

func (r *Rpc) GetBlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	storage := r.executor.BlockStorage().GetBlockByHash(hash)
	if storage == nil {
		return r.reader.BlockByHash(ctx, hash)
	}

	return storage.Block, nil
}

func (r *Rpc) GetStorageAt(
	ctx context.Context,
	account common.Address,
	slot common.Hash,
	blockNumber string,
) (string, error) {
	_, latest := r.executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return "0x", err
	}

	if block.Uint64() == latest {
		hash, err := r.forkDB.GetState(ctx, account, slot)
		if err != nil {
			return "0x", err
		}
		return hash.Hex(), nil
	}

	if block.Uint64() >= r.cfg.ForkBlock.Uint64() {
		state, err := getStateFromBlockStorage(r.executor, account, slot, block.Uint64())
		if err != nil {
			return "0x", err
		}
		if state == common.HexToHash("") {
			return getStateFromReader(ctx, r.reader, account, slot, block)
		}
		return state.Hex(), nil
	}

	return getStateFromReader(ctx, r.reader, account, slot, block)
}

func (r *Rpc) GetHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	block, err := r.GetBlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	return block.Header(), nil
}

func (r *Rpc) GetHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error) {
	block, err := r.GetBlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return block.Header(), nil
}

func (r *Rpc) Call(
	ctx context.Context,
	msg jsonCallMsg,
	blockNumber string,
) (string, error) {
	t := tracer.NewTracer(false)
	call, err := createEthCallMsg(msg)
	if err != nil {
		return "0x", err
	}

	// TODO: handle blockNumber
	ret, _, err := r.executor.Call(ctx, call, t.Hooks(), entity.StateOverrides{})
	if err != nil {
		return "0x", err
	}

	return hexutil.Encode(ret), nil
}

func (r *Rpc) SendRawTransaction(
	ctx context.Context,
	encoded string,
) (string, error) {
	t := tracer.NewTracer(false)
	decoded, err := decodeHexString(encoded)
	if err != nil {
		return "0x", err
	}

	tx := types.NewTx(&types.LegacyTx{})
	if err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(decoded), uint64(len(decoded)))); err != nil {
		return "0x", err
	}

	from, ok := ctx.Value(server.Caller{}).(common.Address)
	if !ok {
		return "0x", errors.New("failed to parse caller")
	}

	msg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	txHash, _, _, err := r.executor.CallAndPersist(ctx, msg, t.Hooks(), entity.StateOverrides{})
	if err != nil {
		return "0x", err
	}

	fmt.Println(t.Fmt())

	return txHash.Hex(), nil
}

func (r *Rpc) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt := r.executor.TxnStorage().GetReceipt(txHash)
	if receipt == nil {
		return r.reader.TransactionReceipt(ctx, txHash)
	}

	return receipt, nil
}

func (r *Rpc) GetTransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	var err error
	txn := r.executor.TxnStorage().GetTransaction(txHash)
	if txn == nil {
		txn, _, err = r.reader.TransactionByHash(ctx, txHash)
		return txn, err
	}

	return txn, nil
}

func (r *Rpc) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return 0, nil
}

func (r *Rpc) GasPrice(ctx context.Context) (string, error) {
	return new(big.Int).SetInt64(0).String(), nil
}

func (r *Rpc) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	return getBlockFromStorageOrReader(r.executor, r.reader, number)
}

func (r *Rpc) GetBalance(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	_, latest := r.executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return hexPrefix, err
	}

	if block.Uint64() == r.cfg.ForkBlock.Uint64() {
		return getBalanceFromForkDB(ctx, r.forkDB, account)
	}

	bal, err := getBalanceFromBlockStorage(r.executor, account, block.Uint64())
	if err != nil || bal == hexPrefix {
		return getBalanceFromReader(ctx, r.reader, account, block)
	}
	return bal, nil
}

func (r *Rpc) GetCode(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	_, latest := r.executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return "0x", err
	}

	if block.Uint64() == latest {
		code, err := r.forkDB.GetCode(ctx, account)
		if err != nil {
			return "0x", err
		}
		return hexutil.Encode(code), nil
	}

	if block.Uint64() >= r.cfg.ForkBlock.Uint64() {
		return getCodeFromBlockStorage(r.executor, account, block.Uint64())
	}

	return getCodeFromReader(ctx, r.reader, account, block)
}
