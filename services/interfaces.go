package services

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/fork"
)

type executor interface {
	CallAndPersist(
		ctx context.Context,
		tx ethereum.CallMsg,
		hooks *tracing.Hooks,
		overrides entity.StateOverrides,
	) (hash *common.Hash, ret []byte, leftOverGas uint64, err error)
	Call(
		ctx context.Context,
		tx ethereum.CallMsg,
		hooks *tracing.Hooks,
		overrides entity.StateOverrides,
	) (ret []byte, leftOverGas uint64, err error)
	CallWithDB(
		ctx context.Context,
		tx ethereum.CallMsg,
		hooks *tracing.Hooks,
		db *fork.DB,
		overrides entity.StateOverrides,
	) (ret []byte, leftOverGas uint64, err error)
	TxnStorage() *entity.TransactionStorage
	BlockStorage() *entity.BlockStorage
	Latest() (common.Hash, uint64)
}

type forkDB interface {
	CreateState(ctx context.Context, addr common.Address) error
	State(ctx context.Context, addr common.Address) (*entity.AccountState, *entity.AccountStorage, error)
	GetBalance(ctx context.Context, addr common.Address) (*big.Int, error)
	SetBalance(ctx context.Context, addr common.Address, amount *big.Int) error
	GetNonce(ctx context.Context, addr common.Address) (uint64, error)
	SetNonce(ctx context.Context, addr common.Address, nonce uint64) error
	GetCodeHash(ctx context.Context, addr common.Address) (common.Hash, error)
	GetCode(ctx context.Context, addr common.Address) ([]byte, error)
	GetCodeSize(ctx context.Context, addr common.Address) (int, error)
	GetState(ctx context.Context, addr common.Address, hash common.Hash) (common.Hash, error)
	ApplyState(s *entity.AccountsState)
	ApplyStorage(s *entity.AccountsStorage)
}

type executionCtx interface {
	GetOrCreate(ctx context.Context) (*ExecutionCtx, error)
}
