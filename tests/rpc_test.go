package tests

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"
	"github.com/go-resty/resty/v2"
	"github.com/raul0ligma/smelter/app"
	"github.com/raul0ligma/smelter/types"
	"github.com/stretchr/testify/require"
)

const (
	_headerCaller = "X-Caller"
)

type jsonRpcMessage struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

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
		if err = app.Run(ctx, rpcURL, block, chainID, time.Minute*5, time.Minute*10, started); err != nil {
			errChan <- err
			return
		}
	}(started, errChan)

	select {
	case err = <-errChan:
		return err
	case <-started:
		<-time.After(time.Second * 5)
		return nil
	case <-time.After(time.Second * 10):
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

	forkedRPC := "http://0.0.0.0:6969/v1/rpc/something"
	r, err := rpc.DialContext(ctx, forkedRPC)
	require.NoError(t, err, "failed to connect to forked rpc")
	client := ethclient.NewClient(r)

	wethAddr := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	weth, err := NewErc20Binding(wethAddr, client)
	require.NoError(t, err, "failed to create WETH binding")

	beforeBal, err := weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("before weth balance", beforeBal.String())

	ethBal, err := client.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("before eth balance", ethBal.String())

	r.SetHeader(_headerCaller, types.Address0x69.Hex())
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

func Test_ReadPrevBlock(t *testing.T) {
	forkBlock := uint64(20011602)
	chainID := new(big.Int).SetInt64(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := start(ctx, chainID, forkBlock); err != nil {
		t.Error(err)
		return
	}

	forkedRPC := "http://0.0.0.0:6969/v1/rpc/something"
	r, err := rpc.DialContext(ctx, forkedRPC)
	require.NoError(t, err, "failed to connect to forked rpc")
	client := ethclient.NewClient(r)

	wethAddr := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	weth, err := NewErc20Binding(wethAddr, client)
	require.NoError(t, err, "failed to create WETH binding")

	r.SetHeader(_headerCaller, types.Address0x69.Hex())
	depositCall, _ := hexutil.Decode("0xd0e30db0")
	tx := types2.NewTx(&types2.LegacyTx{
		Nonce: 1,
		Gas:   30000,
		To:    &wethAddr,
		Value: new(big.Int).SetInt64(1000000000000),
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

	prevBlock := new(big.Int).SetUint64(forkBlock)
	beforeBal, err := weth.BalanceOf(&bind.CallOpts{
		BlockNumber: prevBlock,
	}, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("before weth balance", beforeBal.String())
	require.Equal(t, "0", beforeBal.String(), "wrong weth bal prev block")

	ethBal, err := client.BalanceAt(ctx, types.Address0x69, prevBlock)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("before eth balance", ethBal.String())
	require.Equal(t, "1000000000000", ethBal.String(), "wrong eth bal prev block")

	beforeBal, err = weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("after weth balance", beforeBal.String())
	require.Equal(t, "1000000000000", beforeBal.String(), "wrong weth bal post block")

	ethBal, err = client.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("after eth balance", ethBal.String())
	require.Equal(t, "0", ethBal.String(), "wrong eth bal post block")

	_, err = resty.New().R().SetBody(&jsonRpcMessage{
		Jsonrpc: "2.0",
		Id:      1,
		Method:  "eth_setBalance",
		Params:  []interface{}{types.Address0x69.Hex(), "5000000"},
	}).Post(forkedRPC)

	require.NoError(t, err, "failed to topup balance")

	transferCall, _ := hexutil.Decode("0xa9059cbb00000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000064")
	r.SetHeader(_headerCaller, types.Address0x69.Hex())
	transferTx := types2.NewTx(&types2.LegacyTx{
		Nonce: 2,
		Gas:   5e6,
		To:    &wethAddr,
		Value: new(big.Int).SetInt64(0),
		Data:  transferCall,
	})

	t.Log("sending transaction", transferTx.Hash())
	err = client.SendTransaction(ctx, transferTx)
	require.NoError(t, err, "failed to call SendTransaction")

	input, _, err = client.TransactionByHash(ctx, transferTx.Hash())
	require.NoError(t, err, "failed to get TransactionByHash")
	logJSON(t, input, "transaction")

	receipt, err = client.TransactionReceipt(ctx, transferTx.Hash())
	require.NoError(t, err, "failed to get TransactionReceipt")
	logJSON(t, receipt, "receipt")

	receiver := common.HexToAddress("0x0000000000000000000000000000000000000007")
	prevBlock = prevBlock.Add(prevBlock, new(big.Int).SetInt64(1))

	beforeBal, err = weth.BalanceOf(&bind.CallOpts{
		BlockNumber: prevBlock,
	}, receiver)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("before weth balance", beforeBal.String())
	require.Equal(t, "0", beforeBal.String(), "wrong weth bal pre block")

	beforeBal, err = weth.BalanceOf(nil, receiver)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("after weth balance", beforeBal.String())
	require.Equal(t, "100", beforeBal.String(), "wrong weth bal post block")

	beforeBal, err = weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("after weth balance", beforeBal.String())
	require.Equal(t, "999999999900", beforeBal.String(), "wrong weth bal post block")
}

func Test_SmelterRpc(t *testing.T) {
	forkBlock := uint64(20011602)
	chainID := new(big.Int).SetInt64(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := start(ctx, chainID, forkBlock); err != nil {
		t.Error(err)
		return
	}

	forkedRPC := "http://0.0.0.0:6969/v1/rpc/something"
	r, err := rpc.DialContext(ctx, forkedRPC)
	require.NoError(t, err, "failed to connect to forked rpc")
	client := ethclient.NewClient(r)

	wethAddr := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	weth, err := NewErc20Binding(wethAddr, client)
	require.NoError(t, err, "failed to create WETH binding")

	beforeBal, err := weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceOf")
	t.Log("before weth balance", beforeBal.String())

	ethBal, err := client.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call BalanceAt")
	t.Log("before eth balance", ethBal.String())

	_, err = resty.New().R().SetBody(&jsonRpcMessage{
		Jsonrpc: "2.0",
		Id:      1,
		Method:  "smelter_impersonateAccount",
		Params:  []interface{}{types.Address0x69.Hex()},
	}).Post(forkedRPC)
	require.NoError(t, err, "failed to call smelter_impersonateAccount")

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
