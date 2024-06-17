package services

import (
	"context"

	"github.com/rahul0tripathi/smelter/entity"
)

type ErigonRpc struct {
	backend otterscanBackend
}

func NewErigonRpc(backend otterscanBackend) *ErigonRpc {
	return &ErigonRpc{
		backend: backend,
	}
}

func (o *ErigonRpc) GetHeaderByNumber(ctx context.Context, blockNum uint64) (*entity.BlockData, error) {
	b, err := o.backend.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, err
	}

	return &entity.SerializeBlockDetail(b).Block, nil
}
