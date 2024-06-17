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
	"github.com/rahul0tripathi/smelter/fork"
	"github.com/rahul0tripathi/smelter/pkg/server"
	"github.com/rahul0tripathi/smelter/tracer"
)

type readerAndCaller interface {
	entity.ChainStateReader
	ethereum.ContractCaller
}

type EthRpc struct {
	execStorage     executionCtx
	readerAndCaller readerAndCaller
	cfg             entity.ForkConfig
}

func NewRpcService(
	storage executionCtx,
	cfg entity.ForkConfig,
	readerAndCaller readerAndCaller,
) *EthRpc {
	return &EthRpc{
		execStorage:     storage,
		cfg:             cfg,
		readerAndCaller: readerAndCaller,
	}
}

func (r *EthRpc) ChainId(_ context.Context) string {
	return hexutil.Encode(new(big.Int).SetUint64(r.cfg.ChainID).Bytes())
}

func (r *EthRpc) BlockNumber(ctx context.Context) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	_, blockNum := execCtx.Executor.Latest()
	if blockNum == 0 {
		return hexutil.Encode(r.cfg.ForkBlock.Bytes()), nil
	}

	return hexutil.Encode(new(big.Int).SetUint64(blockNum).Bytes()), nil
}

func (r *EthRpc) GetBlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	storage := execCtx.Executor.BlockStorage().GetBlockByHash(hash)
	if storage == nil {
		return r.readerAndCaller.BlockByHash(ctx, hash)
	}

	return storage.Block, nil
}

func (r *EthRpc) GetStorageAt(
	ctx context.Context,
	account common.Address,
	slot common.Hash,
	blockNumber string,
) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	_, latest := execCtx.Executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return "0x", err
	}

	if block.Uint64() == latest {
		hash, err := execCtx.Db.GetState(ctx, account, slot)
		if err != nil {
			return "0x", err
		}
		return hash.Hex(), nil
	}

	if block.Uint64() > r.cfg.ForkBlock.Uint64() {
		state, err := getStateFromBlockStorage(ctx, execCtx.Executor, r.cfg.ChainID, r.readerAndCaller, account, slot, block.Uint64())
		if err != nil {
			return "0x", err
		}
		if state == common.HexToHash("") {
			return getStateFromReader(ctx, r.readerAndCaller, account, slot, block)
		}
		return state.Hex(), nil
	}

	return getStateFromReader(ctx, r.readerAndCaller, account, slot, block)
}

func (r *EthRpc) GetHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	block, err := r.GetBlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	return block.Header(), nil
}

func (r *EthRpc) GetHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error) {
	block, err := r.GetBlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return block.Header(), nil
}

func (r *EthRpc) Call(
	ctx context.Context,
	msg jsonCallMsg,
	blockNumber string,
) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	_, latest := execCtx.Executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return "0x", err
	}

	call, err := createEthCallMsg(msg)
	t := tracer.NewTracer(false)
	if block.Uint64() == latest {
		ret, _, err := execCtx.Executor.Call(ctx, call, t.Hooks(), entity.StateOverrides{})
		if err != nil {
			return "0x", err
		}

		return hexutil.Encode(ret), nil
	}

	if block.Uint64() > r.cfg.ForkBlock.Uint64() {
		storage, err := getBlockStorage(execCtx.Executor, block.Uint64())
		if err != nil {
			return "0x", err
		}

		db := fork.NewDB(r.readerAndCaller, r.cfg, storage.Accounts, storage.State)
		ret, _, err := execCtx.Executor.CallWithDB(ctx, call, t.Hooks(), db, entity.StateOverrides{})
		if err != nil {
			return "0x", err
		}

		return hexutil.Encode(ret), nil
	}

	return callOnReader(ctx, r.readerAndCaller, call, block)
}

func (r *EthRpc) SendRawTransaction(
	ctx context.Context,
	encoded string,
) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	t := tracer.NewTracer(false)
	decoded, err := decodeHexString(encoded)
	if err != nil {
		return "0x", err
	}

	tx := types.NewTx(&types.LegacyTx{})
	if err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(decoded), uint64(len(decoded)))); err != nil {
		return "0x", err
	}

	caller := execCtx.Impersonator
	if caller == common.HexToAddress("") {
		from, ok := ctx.Value(server.Caller{}).(common.Address)
		if !ok {
			return "0x", errors.New("failed to parse caller")
		}
		caller = from
	}

	msg := ethereum.CallMsg{
		From:     caller,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	txHash, _, _, err := execCtx.Executor.CallAndPersist(ctx, msg, t.Hooks(), entity.StateOverrides{})
	if err != nil {
		return "0x", err
	}

	fmt.Println(t.Fmt())

	return txHash.Hex(), nil
}

func (r *EthRpc) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	receipt := execCtx.Executor.TxnStorage().GetReceipt(txHash)
	if receipt == nil {
		return r.readerAndCaller.TransactionReceipt(ctx, txHash)
	}

	return receipt, nil
}

func (r *EthRpc) GetTransactionByHash(ctx context.Context, txHash common.Hash) (*entity.SerializedTransaction, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	txn := execCtx.Executor.TxnStorage().GetTransaction(txHash)
	if txn == nil {
		txn, _, err = r.readerAndCaller.TransactionByHash(ctx, txHash)
		if err != nil {
			return nil, err
		}
	}

	receipt, err := r.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	return entity.SerializeTransaction(txn, receipt), nil
}

func (r *EthRpc) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return 0, nil
}

func (r *EthRpc) GasPrice(ctx context.Context) (string, error) {
	return new(big.Int).SetInt64(0).String(), nil
}

func (r *EthRpc) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	return getBlockFromStorageOrReader(execCtx.Executor, r.readerAndCaller, number)
}

func (r *EthRpc) GetBalance(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	_, latest := execCtx.Executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return hexPrefix, err
	}

	if block.Uint64() == latest {
		return getBalanceFromForkDB(ctx, execCtx.Db, account)
	}

	if block.Uint64() > r.cfg.ForkBlock.Uint64() {
		return getBalanceFromBlockStorage(ctx, execCtx.Executor, r.cfg.ChainID, r.readerAndCaller, account, block.Uint64())
	}

	return getBalanceFromReader(ctx, r.readerAndCaller, account, block)
}

func (r *EthRpc) GetCode(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return "", err
	}

	_, latest := execCtx.Executor.Latest()
	block, err := parseAndValidateBlockNumber(blockNumber, latest)
	if err != nil {
		return "0x", err
	}

	if block.Uint64() == latest {
		code, err := execCtx.Db.GetCode(ctx, account)
		if err != nil {
			return "0x", err
		}
		return hexutil.Encode(code), nil
	}

	if block.Uint64() > r.cfg.ForkBlock.Uint64() {
		return getCodeFromBlockStorage(execCtx.Executor, account, block.Uint64())
	}

	return getCodeFromReader(ctx, r.readerAndCaller, account, block)
}

func (r *EthRpc) SetBalance(ctx context.Context, account common.Address, balance string) error {
	execCtx, err := r.execStorage.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	amount, err := parseBigInt(balance)
	if err != nil {
		return err
	}

	return execCtx.Db.SetBalance(ctx, account, amount)
}
