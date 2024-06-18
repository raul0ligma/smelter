package tests

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/rahul0tripathi/smelter/config"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/executor"
	"github.com/rahul0tripathi/smelter/fork"
	"github.com/rahul0tripathi/smelter/tracer"
	"github.com/rahul0tripathi/smelter/types"
	"github.com/stretchr/testify/require"
)

func TestExecuteE2E(t *testing.T) {
	ctx := context.Background()
	reader := mockProvider{}
	accountsState := entity.NewAccountsState()
	accountsStorage := entity.NewAccountsStorage()
	forkCfg := entity.ForkConfig{
		ChainID:   69,
		ForkBlock: new(big.Int).SetUint64(1),
	}
	db := fork.NewDB(&reader, forkCfg, accountsStorage, accountsState)
	cfg := config.NewConfigWithDefaults()
	cfg.ForkConfig = &forkCfg
	stateTracer := tracer.NewTracer(true)
	target := types.Address0x69
	exec, err := executor.NewExecutor(ctx, cfg, db, &reader)
	if err != nil {
		panic(err)
	}

	sender := common.HexToAddress("0x0000000000000000000000000000000000000006")

	// Deposit transaction
	deposit, _ := hexutil.Decode("0xd0e30db0")
	_, ret, _, err := exec.CallAndPersist(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  deposit,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(6969),
	}, stateTracer, map[common.Address]entity.StateOverride{
		sender: {Balance: abi.MaxUint256},
	})
	require.NoError(t, err, "failed to deposit")
	t.Log("trace", stateTracer.Fmt())

	// Check balance of sender before transfer
	balanceOfx06, _ := hexutil.Decode("0x70a082310000000000000000000000000000000000000000000000000000000000000006")
	ret, _, err = exec.Call(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  balanceOfx06,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer, nil)
	require.NoError(t, err, "failed to read 0x6 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(6969), "invalid 0x6 balance received pre transfer")

	// Transfer transaction
	stateTracer = tracer.NewTracer(true)
	transferCall, _ := hexutil.Decode("0xa9059cbb00000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000001b37")
	_, ret, _, err = exec.CallAndPersist(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  transferCall,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer, map[common.Address]entity.StateOverride{
		sender: {Balance: abi.MaxUint256},
	})
	require.NoError(t, err, "failed to transfer weth")
	t.Log("trace", stateTracer.Fmt())

	// Check balance of recipient after transfer
	balanceOfx07, _ := hexutil.Decode("0x70a082310000000000000000000000000000000000000000000000000000000000000007")
	ret, _, err = exec.Call(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  balanceOfx07,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer, nil)
	require.NoError(t, err, "failed to read 0x7 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(6967), "invalid 0x7 balance received post transfer")

	// Check balance of sender after transfer
	ret, _, err = exec.Call(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  balanceOfx06,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer, nil)
	require.NoError(t, err, "failed to read 0x6 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(2), "invalid 0x6 balance received post transfer")
}

func TestBlockProduction(t *testing.T) {
	ctx := context.Background()
	reader := mockProvider{}
	accountsState := entity.NewAccountsState()
	accountsStorage := entity.NewAccountsStorage()
	forkCfg := entity.ForkConfig{
		ChainID:   69,
		ForkBlock: new(big.Int).SetUint64(1),
	}
	db := fork.NewDB(&reader, forkCfg, accountsStorage, accountsState)
	cfg := config.NewConfigWithDefaults()
	cfg.ForkConfig = &forkCfg
	stateTracer := tracer.NewTracer(true)
	target := types.Address0x69
	exec, err := executor.NewExecutor(ctx, cfg, db, &reader)
	if err != nil {
		panic(err)
	}

	sender := common.HexToAddress("0x0000000000000000000000000000000000000006")

	// Deposit transaction
	deposit, _ := hexutil.Decode("0xd0e30db0")
	msg := ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  deposit,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(6969),
	}
	hash, _, _, err := exec.CallAndPersist(ctx, msg, stateTracer, map[common.Address]entity.StateOverride{
		sender: {Balance: abi.MaxUint256},
	})
	require.NoError(t, err, "failed to deposit")
	t.Log("trace", stateTracer.Fmt())

	require.NotNil(t, hash, "missing tx hash")

	t.Log("transaction Hash", hash.Hex())
	require.Equal(t, "0xfa652df356f74e065519c07cd5473ba8d26383b7f000f6f06563906b6d3d83e0", hash.Hex(), "mismatch txn hash")

	txn := exec.TxnStorage().GetTransaction(*hash)
	require.Equal(t, txn.Type(), uint8(0), "invalid txn type")
	require.Equal(t, *txn.To(), target, "invalid txn target")
	require.Equal(t, txn.Value().String(), "6969", "invalid txn value")
	require.Equal(t, txn.Data(), deposit, "invalid txn data")

	receipt := exec.TxnStorage().GetReceipt(*hash)

	require.Equal(t, len(receipt.Logs), 1, "invalid logs emitted")
	require.Equal(t, receipt.Logs[0].Address, target, "invalid emitter")
	require.Equal(t, receipt.Logs[0].Topics[0].Hex(), "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c", "invalid log topic [0]")
	require.Equal(t, receipt.Logs[0].Topics[1].Hex(), "0x0000000000000000000000000000000000000000000000000000000000000006", "invalid log topic [1]")
	require.Equal(t, hexutil.Encode(receipt.Logs[0].Data), "0x0000000000000000000000000000000000000000000000000000000000001b39", "invalid log data")

	txnBytes, err := txn.MarshalJSON()
	require.NoError(t, err, "MarshalJSON")
	t.Log("transaction", string(txnBytes))

	receiptBytes, err := receipt.MarshalJSON()
	require.NoError(t, err, "MarshalJSON")
	t.Log("receipt", string(receiptBytes))

	blockHash, blockNum := exec.Latest()
	require.NotEqualf(t, blockHash.Hex(), "0x0000000000000000000000000000000000000000000000000000000000000000", "empty block hash")
	require.Equal(t, blockNum, uint64(2), "invalid block number")

	state := exec.BlockStorage().GetBlockByHash(blockHash)

	require.True(t, reflect.DeepEqual(state.Block.Transactions(), types2.Transactions{txn}), "invalid txn found uin block")
	require.Equal(t, state.Block.Number().Uint64(), uint64(2), "invalid block number")
}
