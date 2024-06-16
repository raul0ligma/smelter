package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/rahul0tripathi/smelter/types"
	"github.com/stretchr/testify/require"
)

func Test_ExecutionStorage(t *testing.T) {
	forkBlock := uint64(20011602)
	chainID := new(big.Int).SetInt64(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := start(ctx, chainID, forkBlock); err != nil {
		t.Error(err)
		return
	}

	_, err := resty.New().R().SetBody(&jsonRpcMessage{
		Jsonrpc: "2.0",
		Id:      1,
		Method:  "eth_setBalance",
		Params:  []interface{}{types.Address0x69.Hex(), "0x4C4B40"},
	}).Post("http://0.0.0.0:6969/v1/rpc/execCtxA")
	require.NoError(t, err, "failed to connect to forked rpc")

	cli, err := ethclient.Dial("http://0.0.0.0:6969/v1/rpc/execCtxA")
	require.NoError(t, err, "failed to connect to forked rpc")

	bal, err := cli.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call balance of")

	require.Equal(t, "5000000", bal.String(), "incorrect bal")
	cliB, err := ethclient.Dial("http://0.0.0.0:6969/v1/rpc/execCtxB")
	require.NoError(t, err, "failed to connect to forked rpc")

	balB, err := cliB.BalanceAt(ctx, types.Address0x69, nil)
	require.NoError(t, err, "failed to call balance of")

	require.Equal(t, "1000000000000", balB.String(), "failed to call getBalCtxB")
}
