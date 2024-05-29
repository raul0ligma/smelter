package executor

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/rahul0tripathi/smelter/config"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/fork"
	"github.com/rahul0tripathi/smelter/statedb"
	"github.com/rahul0tripathi/smelter/vm"
)

type SerialExecutor struct {
	mu       sync.RWMutex
	db       *fork.DB
	cfg      *config.Config
	provider entity.ChainStateReader
}

func NewExecutor(cfg *config.Config, db *fork.DB, provider entity.ChainStateReader) *SerialExecutor {
	return &SerialExecutor{
		db:       db,
		cfg:      cfg,
		provider: provider,
	}
}

func (e *SerialExecutor) Exec(
	ctx context.Context,
	tx ethereum.CallMsg,
	hooks *tracing.Hooks,
	overrides entity.StateOverrides,
) (ret []byte, leftOverGas uint64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// TODO: use block from state
	executionDB := statedb.NewDB(ctx, e.db)
	chainCfg, evmCfg := e.cfg.ExecutionConfig(hooks)
	if err = executionDB.ApplyOverrides(overrides); err != nil {
		return nil, 0, err
	}

	env := vm.NewEVM(e.cfg.BlockContext(new(big.Int).Add(e.cfg.ForkConfig.ForkBlock, new(big.Int).SetUint64(1)),
		new(big.Int),
		uint64(time.Now().Unix())),
		e.cfg.TxCtx(tx.From),
		executionDB,
		chainCfg,
		evmCfg)
	value, _ := uint256.FromBig(tx.Value)
	ret, leftOverGas, err = env.Call(
		vm.AccountRef(tx.From),
		*tx.To,
		tx.Data,
		tx.Gas,
		value,
	)

	if err != nil {
		return
	}

	e.db.ApplyStorage(executionDB.Dirty().GetAccountStorage())
	e.db.ApplyState(executionDB.Dirty().GetAccountState())
	return
}

func (e *SerialExecutor) Simulate(
	ctx context.Context,
	tx ethereum.CallMsg,
	hooks *tracing.Hooks,
	overrides entity.StateOverrides,
) (ret []byte, leftOverGas uint64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// TODO: use block from state
	executionDB := statedb.NewDB(ctx, e.db)
	if err = executionDB.ApplyOverrides(overrides); err != nil {
		return nil, 0, err
	}

	chainCfg, evmCfg := e.cfg.ExecutionConfig(hooks)
	env := vm.NewEVM(e.cfg.BlockContext(new(big.Int).Add(e.cfg.ForkConfig.ForkBlock, new(big.Int).SetUint64(1)),
		new(big.Int),
		uint64(time.Now().Unix())),
		e.cfg.TxCtx(tx.From),
		executionDB,
		chainCfg,
		evmCfg)
	value, _ := uint256.FromBig(tx.Value)
	ret, leftOverGas, err = env.Call(
		vm.AccountRef(tx.From),
		*tx.To,
		tx.Data,
		tx.Gas,
		value,
	)

	return
}
