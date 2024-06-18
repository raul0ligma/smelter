package services

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/smelter/entity"
)

type SmelterRpc struct {
	execStorage executionCtx
}

func NewSmelterRpc(exec executionCtx) *SmelterRpc {
	return &SmelterRpc{execStorage: exec}
}

func (s *SmelterRpc) ImpersonateAccount(ctx context.Context, address common.Address) error {
	execCtx, err := s.execStorage.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	execCtx.Impersonator = address
	return nil
}

func (s *SmelterRpc) StopImpersonatingAccount(ctx context.Context) error {
	execCtx, err := s.execStorage.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	execCtx.Impersonator = common.HexToAddress("")
	return nil
}

func (s *SmelterRpc) GetState(ctx context.Context) (json.RawMessage, error) {
	execCtx, err := s.execStorage.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	v, err := json.Marshal(execCtx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *SmelterRpc) SetStateOverrides(ctx context.Context, overrides entity.StateOverrides) error {
	execCtx, err := s.execStorage.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	execCtx.Overrides = overrides
	return nil
}
