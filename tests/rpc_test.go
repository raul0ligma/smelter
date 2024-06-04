package tests

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/rahul0tripathi/smelter/app"
	"github.com/rahul0tripathi/smelter/types"
	"github.com/stretchr/testify/require"
)

func logJSON(t *testing.T, val interface{}, params ...interface{}) {
	v, _ := json.Marshal(val)
	t.Log(string(v), params)
}

func start(ctx context.Context, chainID *big.Int, block uint64) error {
	rpcURL, err := GetRPClient(ctx, chainID.Uint64())
	if err != nil {
		return err
	}
	started := make(chan struct{}, 1)
	errChan := make(chan error, 1)
	go func(startChan chan<- struct{}, errChan chan<- error) {
		if err = app.Run(ctx, rpcURL, block, chainID, started); err != nil {
			errChan <- err
			return
		}
	}(started, errChan)

	select {
	case err = <-errChan:
		return err
	case <-started:
		return nil
	case <-time.After(time.Second * 5):
		return errors.New("failed to start server, timed out")
	}
}

func Test_Rpc(t *testing.T) {
	forkBlock := uint64(20011602)
	chainID := new(big.Int).SetInt64(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := start(ctx, chainID, forkBlock); err != nil {
		t.Error(err)
		return
	}

	forkedRPC := "http://0.0.0.0:6969/v1/rpc"

	client, err := ethclient.DialContext(ctx, forkedRPC)
	require.NoError(t, err, "failed to connect to forked rpc")

	wethAddr := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	weth, err := NewErc20Binding(wethAddr, client)
	require.NoError(t, err, "failed to create WETH binding")

	beforeBal, err := weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("before weth balance", beforeBal.String())

	ethBal, err := client.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("before eth balance", ethBal.String())

	depositCall, _ := hexutil.Decode("0xd0e30db0")
	tx := types2.NewTx(&types2.LegacyTx{
		Nonce: 1,
		Gas:   30000,
		To:    &wethAddr,
		Value: ethBal,
		Data:  depositCall,
	})

	t.Log("sending transaction", tx.Hash())
	err = client.SendTransaction(ctx, tx)
	require.NoError(t, err, "failed to call SendTransaction")

	input, _, err := client.TransactionByHash(ctx, tx.Hash())
	require.NoError(t, err, "failed to get TransactionByHash")
	logJSON(t, input, "transaction")

	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err, "failed to get TransactionReceipt")
	logJSON(t, receipt, "receipt")

	beforeBal, err = weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("after weth balance", beforeBal.String())

	ethBal, err = client.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("after eth balance", ethBal.String())
}
