package services

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Rpc struct {
	executor executor
}

func NewRpcService(executor executor) *Rpc {
	return &Rpc{executor: executor}
}

func (r *Rpc) ChainId(ctx context.Context) string {
	return hexutil.Encode(new(big.Int).SetInt64(69).Bytes())
}
