package statedb

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/raul0ligma/smelter/entity"
)

type forkDB interface {
	State(ctx context.Context, addr common.Address) (*entity.AccountState, *entity.AccountStorage, error)
	CreateState(ctx context.Context, addr common.Address) error
	GetState(ctx context.Context, addr common.Address, hash common.Hash) (common.Hash, error)
}
