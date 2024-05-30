package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	exec := executor.NewExecutor(cfg, db, &reader)

	sender := common.HexToAddress("0x0000000000000000000000000000000000000006")

	// Deposit transaction
	deposit, _ := hexutil.Decode("0xd0e30db0")
	ret, _, err := exec.CallAndPersist(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  deposit,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(6969),
	}, stateTracer.Hooks(), map[common.Address]entity.StateOverride{
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
	}, stateTracer.Hooks(), nil)
	require.NoError(t, err, "failed to read 0x6 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(6969), "invalid 0x6 balance received pre transfer")

	// Transfer transaction
	stateTracer = tracer.NewTracer(true)
	transferCall, _ := hexutil.Decode("0xa9059cbb00000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000001b37")
	ret, _, err = exec.CallAndPersist(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  transferCall,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer.Hooks(), map[common.Address]entity.StateOverride{
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
	}, stateTracer.Hooks(), nil)
	require.NoError(t, err, "failed to read 0x7 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(6967), "invalid 0x7 balance received post transfer")

	// Check balance of sender after transfer
	ret, _, err = exec.Call(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &target,
		Data:  balanceOfx06,
		Gas:   30000000,
		Value: new(big.Int).SetInt64(0),
	}, stateTracer.Hooks(), nil)
	require.NoError(t, err, "failed to read 0x6 balance")
	require.Equal(t, new(big.Int).SetBytes(ret).Int64(), int64(2), "invalid 0x6 balance received post transfer")
}
