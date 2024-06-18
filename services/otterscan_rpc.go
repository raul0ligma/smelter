package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/smelter/entity"
)

type OtterscanRPC struct {
	execStorage executionCtx
	backend     otterscanBackend
}

func NewOtterscanRpc(backend otterscanBackend, execStorage executionCtx) *OtterscanRPC {
	return &OtterscanRPC{backend: backend, execStorage: execStorage}
}

func (o *OtterscanRPC) GetApiLevel(_ context.Context) (int, error) {
	return 8, nil
}

func (o *OtterscanRPC) HasCode(ctx context.Context, address common.Address, block interface{}) (bool, error) {
	blockTag, ok := block.(string)
	if !ok {
		blockNum, ok := block.(uint64)
		if !ok {
			return false, errors.New("invalid block value")
		}
		blockTag = fmt.Sprintf("%d", blockNum)
	}

	code, err := o.backend.GetCode(ctx, address, blockTag)
	if err != nil {
		return false, err
	}

	return len(code) > 0, nil
}

func (o *OtterscanRPC) GetContractCreator(_ context.Context, _ common.Address) (common.Address, error) {
	return common.HexToAddress(""), nil
}

func (o *OtterscanRPC) SearchTransactionsBefore(
	ctx context.Context,
	address common.Address,
	from int,
	to int,
) (*entity.TransactionSearchResponse, error) {
	storage, err := o.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	txStorage := storage.Executor.TxnStorage()
	resp := &entity.TransactionSearchResponse{
		Txs:      make([]*entity.SerializedTransaction, 0),
		Receipts: make([]*entity.SerializedReceipt, 0),
	}

	for _, tx := range txStorage.All() {
		receipt := txStorage.GetReceipt(tx.Hash())
		resp.Txs = append(resp.Txs, entity.SerializeTransaction(tx, receipt))
		resp.Receipts = append(resp.Receipts, entity.SerializeReceipt(receipt))
	}

	return resp, nil
}

func (o *OtterscanRPC) GetBlockDetails(ctx context.Context, block uint64) (*entity.BlockDetailResponse, error) {
	b, err := o.backend.GetBlockByNumber(ctx, strconv.FormatUint(block, 10), false)
	if err != nil {
		return nil, err
	}

	return entity.SerializeBlockDetail(b.Raw), nil
}

func (o *OtterscanRPC) GetTransactionError(_ context.Context, _ common.Hash) (string, error) {
	return "0x", nil
}

func (o *OtterscanRPC) GetBlockTransactions(
	ctx context.Context,
	block uint64,
	from int,
	to int,
) (*entity.BlockTransactionsResponse, error) {
	b, err := o.backend.GetBlockByNumber(ctx, strconv.FormatUint(block, 10), false)
	if err != nil {
		return nil, err
	}

	resp := &entity.BlockTransactionsResponse{
		FullBlock: entity.FullBlock{
			BlockData:    entity.SerializeBlockDetailFromSeralizedBlock(b).Block,
			Transactions: make([]*entity.SerializedTransaction, 0),
		},
		Receipts: make([]*entity.SerializedReceipt, 0),
	}

	txns := b.Raw.Transactions()
	for i := 0; i < 2 && i < len(b.Raw.Transactions()); i++ {
		receipt, err := o.backend.GetTransactionReceipt(ctx, txns[i].Hash())
		if err != nil {
			return nil, err
		}

		resp.FullBlock.Transactions = append(resp.FullBlock.Transactions, entity.SerializeTransaction(txns[i], receipt))
		resp.Receipts = append(resp.Receipts, entity.SerializeReceipt(receipt))
	}

	return resp, nil
}

func (o *OtterscanRPC) TraceTransaction(ctx context.Context, hash common.Hash) (entity.TransactionTraces, error) {
	exec, err := o.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	return exec.Executor.TxnStorage().GetTrace(hash), nil
}
