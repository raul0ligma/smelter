package entity

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum"
)

type ChainStateAndTransactionReader interface {
	ethereum.BlockNumberReader
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.ChainIDReader
	BatchedRpc
}

type BatchReq struct {
	Method string
	Params []any
}

type BatchedRpc interface {
	SupportsBatching() bool
	BatchWithUnmarshal(ctx context.Context, requests []BatchReq, outputs []any) error
	Batch(ctx context.Context, requests []BatchReq) ([]json.RawMessage, error)
}

type BatchedStateReader interface {
	ethereum.ChainStateReader
	BatchedRpc
}
