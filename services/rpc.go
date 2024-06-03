package services

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/tracer"
	types2 "github.com/rahul0tripathi/smelter/types"
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

func (r *Rpc) GetBalance(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	switch {
	case blockNumber == "", blockNumber == "latest":
		return getBalanceFromForkDB(ctx, r.forkDB, account)

	case len(blockNumber) > 2 && blockNumber[:2] == "0x":
		block, err := parseBlockNumber(blockNumber)
		if err != nil {
			return "0x", err
		}

		_, latest := r.executor.Latest()
		switch {
		case block.Uint64() > latest:
			return "0x", fmt.Errorf("invalid block height, received %d, current %d", block.Uint64(), latest)

		case block.Uint64() == latest:
			return getBalanceFromForkDB(ctx, r.forkDB, account)

		case block.Uint64() >= r.cfg.ForkBlock.Uint64():
			bal, err := getBalanceFromBlockStorage(r.executor, account, block.Uint64())
			switch {
			case err != nil:
				return "0x", err
			case bal == "0x":
				return getBalanceFromReader(ctx, r.reader, account, block)
			default:
				return bal, nil
			}

		default:
			return getBalanceFromReader(ctx, r.reader, account, block)
		}

	default:
		return "0x", fmt.Errorf("failed to parse block number %s", blockNumber)
	}
}

func (r *Rpc) GetBlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	storage := r.executor.BlockStorage().GetBlockByHash(hash)
	if storage == nil {
		return r.reader.BlockByHash(ctx, hash)
	}

	return storage.Block, nil
}

func (r *Rpc) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {

	storage := r.executor.BlockStorage().GetBlockByNumber(number)
	if storage == nil {
		return r.reader.BlockByNumber(ctx, new(big.Int).SetUint64(number))
	}

	return storage.Block, nil
}

func (r *Rpc) GetStorageAt(
	ctx context.Context,
	account common.Address,
	slot common.Hash,
	blockNumber string,
) (string, error) {
	switch {
	case blockNumber == "", blockNumber == "latest":
		hash, err := r.forkDB.GetState(ctx, account, slot)
		if err != nil {
			return "0x", err
		}

		return hash.Hex(), nil

	case len(blockNumber) > 2 && blockNumber[:2] == "0x":
		block, err := parseBlockNumber(blockNumber)
		if err != nil {
			return "0x", err
		}

		_, latest := r.executor.Latest()
		switch {
		case block.Uint64() > latest:
			return "0x", fmt.Errorf("invalid block height, received %d, current %d", block.Uint64(), latest)

		case block.Uint64() == latest:
			hash, err := r.forkDB.GetState(ctx, account, slot)
			if err != nil {
				return "0x", err
			}

			return hash.Hex(), nil

		case block.Uint64() >= r.cfg.ForkBlock.Uint64():
			state, err := getStateFromBlockStorage(r.executor, account, slot, block.Uint64())
			if err != nil {
				return "0x", err
			}

			if state == common.HexToHash("") {
				return getStateFromReader(ctx, r.reader, account, slot, block)
			}

			return state.Hex(), nil

		default:
			return getStateFromReader(ctx, r.reader, account, slot, block)
		}

	default:
		return "0x", fmt.Errorf("failed to parse block number %s", blockNumber)
	}

}

func (r *Rpc) GetCode(ctx context.Context, account common.Address, blockNumber string) (string, error) {
	switch {
	case blockNumber == "", blockNumber == "latest":
		code, err := r.forkDB.GetCode(ctx, account)
		if err != nil {
			return "0x", err
		}

		return hexutil.Encode(code), nil

	case len(blockNumber) > 2 && blockNumber[:2] == "0x":
		block, err := parseBlockNumber(blockNumber)
		if err != nil {
			return "0x", err
		}

		_, latest := r.executor.Latest()
		switch {
		case block.Uint64() > latest:
			return "0x", fmt.Errorf("invalid block height, received %d, current %d", block.Uint64(), latest)

		case block.Uint64() == latest:
			code, err := r.forkDB.GetCode(ctx, account)
			if err != nil {
				return "0x", err
			}

			return hexutil.Encode(code), nil

		case block.Uint64() >= r.cfg.ForkBlock.Uint64():
			return getCodeFromBlockStorage(r.executor, account, block.Uint64())

		default:
			code, err := r.reader.CodeAt(ctx, account, block)
			if err != nil {
				return "0x", err
			}

			return hexutil.Encode(code), nil
		}

	default:
		return "0x", fmt.Errorf("failed to parse block number %s", blockNumber)
	}
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

type jsonCallMsg struct {
	From      common.Address // the sender of the 'transaction'
	To        common.Address // the destination contract (nil for contract creation)
	Gas       uint64         // if 0, the call executes with near-infinite gas
	GasPrice  string         // wei <-> gas exchange ratio
	GasFeeCap string         // EIP-1559 fee cap per gas.
	GasTipCap string         // EIP-1559 tip per gas.
	Value     string         // amount of wei sent along with the call
	Input     string         // input data, usually an ABI-encoded contract method invocation
}

func (r *Rpc) Call(
	ctx context.Context,
	msg *jsonCallMsg,
	// TODO: add block number handling with state
	blockNumber string,
) (string, error) {
	t := tracer.NewTracer(false)
	call := ethereum.CallMsg{}
	call.Value = new(big.Int).SetUint64(0)
	if msg.Value != "" {
		call.Value, _ = new(big.Int).SetString(msg.Value, 10)
	}

	call.Gas = 30e6
	call.From = msg.From
	call.To = &msg.To
	callData, _ := hexutil.Decode(msg.Input)

	call.Data = callData

	ret, _, err := r.executor.Call(ctx, call, t.Hooks(), entity.StateOverrides{})
	if err != nil {
		return "0x", err
	}

	return hexutil.Encode(ret), nil
}

func (r *Rpc) SendRawTransaction(
	ctx context.Context,
	// RLP encoded transaction
	encoded string,
) (string, error) {
	t := tracer.NewTracer(false)
	decoded, err := hexutil.Decode(encoded)
	if err != nil {
		return "0x", err
	}

	tx := types.NewTx(&types.LegacyTx{})
	if err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(decoded), uint64(len(decoded)))); err != nil {
		return "0x", err
	}

	from := types2.Address0x69
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
