package entity

import "github.com/ethereum/go-ethereum/core/tracing"

type TraceProvider interface {
	Hooks() *tracing.Hooks
	OtterTrace() TransactionTraces
}
