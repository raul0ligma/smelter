package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/smelter/app"
	"github.com/rahul0tripathi/smelter/types"
	"github.com/stretchr/testify/require"
)

type jsonRPCRequest struct {
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func Test_Rpc(t *testing.T) {
	forkBlock := 20011602
	chainID := new(big.Int).SetInt64(1)
	ctx, cancel := context.WithCancel(context.Background())
	rpcURL, err := GetRPClient(ctx, chainID.Uint64())
	require.NoError(t, err, "failed to get rpc client")

	go func() {
		err := app.Run(ctx, rpcURL, uint64(forkBlock), chainID)
		if err != nil {
			t.Error(err, "failed to spin up server")
			cancel()
		}
		fmt.Println("server running")
	}()
	defer cancel()
	defer func() {
		fmt.Println("callig")
		p, _ := os.FindProcess(os.Getpid())
		err := p.Signal(syscall.SIGINT)
		if err != nil {
			fmt.Println(err)
		}
	}()

	<-time.After(time.Second * 5)
	require.NoError(t, err, "failed to start app")

	forkedRPC := "http://0.0.0.0:6969/v1/rpc"

	client, err := ethclient.DialContext(ctx, forkedRPC)
	require.NoError(t, err, "failed to connect to forked rpc")
	wethAddr := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	weth, err := NewErc20Binding(wethAddr, client)
	require.NoError(t, err, "failed to create weth binding")

	beforeBal, err := weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceIOf")
	t.Log("before balance", beforeBal.String())
	deposit, _ := hexutil.Decode("0xd0e30db0")
	amt, _ := new(big.Int).SetString("1000000000000", 10)

	tx := types2.NewTx(&types2.LegacyTx{
		Nonce: 1,
		Gas:   30000,
		To:    &wethAddr,
		Value: amt,
		Data:  deposit,
	})

	fmt.Println(tx.Hash(), "sending")

	err = client.SendTransaction(ctx, tx)
	require.NoError(t, err, "failed to call deposit")

	txInput, _, err := client.TransactionByHash(ctx, tx.Hash())
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println(txInput)

	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		t.Fatal(err)
		return
	}

	v, _ := json.Marshal(receipt)

	fmt.Println(string(v))

	beforeBal, err = weth.BalanceOf(nil, types.Address0x69)
	require.NoError(t, err, "failed to call balanceIOf")
	t.Log("after balance", beforeBal.String())

}
