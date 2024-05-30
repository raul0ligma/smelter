package services

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/rahul0tripathi/smelter/entity"
)

type executor interface {
	CallAndPersist(
		ctx context.Context,
		tx ethereum.CallMsg,
		hooks *tracing.Hooks,
		overrides entity.StateOverrides,
	) (ret []byte, leftOverGas uint64, err error)
	Call(
		ctx context.Context,
		tx ethereum.CallMsg,
		hooks *tracing.Hooks,
		overrides entity.StateOverrides,
	) (ret []byte, leftOverGas uint64, err error)
}
