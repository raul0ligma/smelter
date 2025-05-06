package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/raul0ligma/smelter/config"
	"github.com/raul0ligma/smelter/entity"
	executorPkg "github.com/raul0ligma/smelter/executor"
	"github.com/raul0ligma/smelter/fork"
	"github.com/raul0ligma/smelter/pkg/server"
)

type ExecutionCtx struct {
	Impersonator common.Address
	Overrides    entity.StateOverrides
	CreatedAt    time.Time
	Executor     executor
	Db           forkDB
}

type ExecutionCtxStorage struct {
	cfg             entity.ForkConfig
	reader          entity.ChainStateAndTransactionReader
	mu              sync.RWMutex
	storage         map[string]*ExecutionCtx
	executionCtxTTL time.Duration
}

func NewExecutionStorage(
	cfg entity.ForkConfig,
	reader entity.ChainStateAndTransactionReader,
	executionCtxTTL time.Duration,
) *ExecutionCtxStorage {
	return &ExecutionCtxStorage{
		cfg:             cfg,
		reader:          reader,
		executionCtxTTL: executionCtxTTL,
		storage:         make(map[string]*ExecutionCtx),
	}
}

func (e *ExecutionCtxStorage) cleanup() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for k, v := range e.storage {
		if time.Now().After(v.CreatedAt.Add(e.executionCtxTTL)) {
			delete(e.storage, k)
		}
	}
}

func (e *ExecutionCtxStorage) Watcher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.cleanup()
		case <-ctx.Done():
			fmt.Println("killing state watcher")
			return
		}
	}
}

func (e *ExecutionCtxStorage) create(ctx context.Context, key string) (*ExecutionCtx, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	db := fork.NewDB(e.reader, e.cfg, entity.NewAccountsStorage(), entity.NewAccountsState())
	cfg := config.NewConfigWithDefaults()
	cfg.ForkConfig = &e.cfg

	exec, err := executorPkg.NewExecutor(ctx, cfg, db, e.reader)
	if err != nil {
		return nil, fmt.Errorf("new executor error: %w", err)
	}

	execCtx := &ExecutionCtx{
		CreatedAt: time.Now(),
		Executor:  exec,
		Db:        db,
	}

	e.storage[key] = execCtx
	return execCtx, nil
}

func (e *ExecutionCtxStorage) getOrCreate(ctx context.Context, key string) (*ExecutionCtx, error) {
	e.mu.RLock()
	execCtx, ok := e.storage[key]
	e.mu.RUnlock()
	if !ok {
		return e.create(ctx, key)
	}

	return execCtx, nil
}

func (e *ExecutionCtxStorage) Get(key string) (*ExecutionCtx, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	execCtx, ok := e.storage[key]
	if !ok {
		return nil, errors.New("not found")
	}

	return execCtx, nil
}

func (e *ExecutionCtxStorage) GetOrCreate(ctx context.Context) (*ExecutionCtx, error) {
	caller, ok := ctx.Value(server.Key{}).(string)
	if !ok {
		caller = "default"
	}

	return e.getOrCreate(ctx, caller)
}
