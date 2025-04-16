package entity

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum"
)

type ChainStateAndTransactionReader interface {
	ethereum.BlockNumberReader
	ethereum.ChainReader
	BatchedStateReader
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.ChainIDReader
}

type BatchReq struct {
	Method any
	Params []any
}

type BatchedStateReader interface {
	ethereum.ChainStateReader
	SupportsBatching() bool
	BatchWithUnmarshal(ctx context.Context, requests []BatchReq, outputs []any) error
	Batch(ctx context.Context, requests []BatchReq) ([]json.RawMessage, error)
}
