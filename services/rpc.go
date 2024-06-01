package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rahul0tripathi/smelter/entity"
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

//type Client interface {
//	BalanceAt(ctx context.Context, account common.Address, blockNumber string) (string, error)
//	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
//	BlockByNumber(ctx context.Context, number string) (*types.Block, error)
//	BlockNumber(ctx context.Context,) (uint64, error)
//	ChainID(ctx context.Context,) (*big.Int, error)
//	CodeAt(ctx context.Context, account common.Address, blockNumber string) (string, error)
//	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
//	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
//	HeaderByNumber(ctx context.Context, number string) (*types.Header, error)
//	NetworkID(ctx context.Context,) (string, error)
//	PendingBalanceAt(ctx context.Context, account common.Address) (string, error)
//	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
//	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
//	PendingTransactionCount(ctx context.Context,) (uint, error)
//	SuggestGasPrice(ctx context.Context,) (string, error)
//	SuggestGasTipCap(ctx context.Context,) (string, error)
//	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
//	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
//	TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error)
//	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
//	SyncProgress(ctx context.Context,) (*ethereum.SyncProgress, error)
//	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber string) (string, error)
//	PendingCallContract(ctx context.Context, msg ethereum.CallMsg) (string, error)
//	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
//	SendTransaction(ctx context.Context, tx *types.Transaction) error
//}

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

func (r *Rpc) GetBalance(ctx context.Context, account common.Address, blockNumber string) (*string, error) {
	switch {
	case blockNumber == "", blockNumber == "latest":
		return getBalanceFromForkDB(ctx, r.forkDB, account)

	case len(blockNumber) > 2 && blockNumber[:2] == "0x":
		block, err := parseBlockNumber(blockNumber)
		if err != nil {
			return nil, err
		}

		_, latest := r.executor.Latest()
		switch {
		case block.Uint64() > latest:
			return nil, fmt.Errorf("invalid block height, received %d, current %d", block.Uint64(), latest)

		case block.Uint64() == latest:
			return getBalanceFromForkDB(ctx, r.forkDB, account)

		case block.Uint64() >= r.cfg.ForkBlock.Uint64():
			return getBalanceFromBlockStorage(r.executor, account, block.Uint64())

		default:
			return getBalanceFromReader(ctx, r.reader, account, block)
		}

	default:
		return nil, fmt.Errorf("failed to parse block number %s", blockNumber)
	}
}
