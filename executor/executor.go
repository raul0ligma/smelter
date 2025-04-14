package executor

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/rahul0tripathi/smelter/config"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/fork"
	"github.com/rahul0tripathi/smelter/producer"
	"github.com/rahul0tripathi/smelter/statedb"
	"github.com/rahul0tripathi/smelter/vm"
)

type SerialExecutor struct {
	mu            sync.RWMutex
	db            *fork.DB
	cfg           *config.Config
	provider      entity.ChainStateReader
	txn           *entity.TransactionStorage
	blocks        *entity.BlockStorage
	prevBlockHash common.Hash
	prevBlockNum  uint64
}

func NewExecutor(
	ctx context.Context,
	cfg *config.Config,
	db *fork.DB,
	provider entity.ChainStateReader,
	opts ...Option,
) (*SerialExecutor, error) {
	e := &SerialExecutor{
		db:       db,
		cfg:      cfg,
		provider: provider,
		txn:      entity.NewTransactionStorage(),
		blocks:   entity.NewBlockStorage(),
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.prevBlockHash == common.HexToHash("") {
		block, err := provider.BlockByNumber(ctx, cfg.ForkConfig.ForkBlock)
		if err != nil {
			return nil, fmt.Errorf("fetch block err %w", err)
		}

		e.prevBlockNum = cfg.ForkConfig.ForkBlock.Uint64()
		e.prevBlockHash = block.Hash()
	}
	return e, nil
}

func (e *SerialExecutor) CallAndPersist(
	ctx context.Context,
	tx ethereum.CallMsg,
	tracer entity.TraceProvider,
	overrides entity.StateOverrides,
) (txHash *common.Hash, ret []byte, leftOverGas uint64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	executionDB := statedb.NewDB(ctx, e.db)
	chainCfg, evmCfg := e.cfg.ExecutionConfig(tracer.Hooks())
	if err = executionDB.ApplyOverrides(overrides); err != nil {
		return nil, nil, 0, err
	}

	env := vm.NewEVM(e.cfg.BlockContext(new(big.Int).Add(e.cfg.ForkConfig.ForkBlock, new(big.Int).SetUint64(1)),
		new(big.Int),
		uint64(time.Now().Unix())),
		executionDB,
		chainCfg,
		evmCfg)
	value, _ := uint256.FromBig(tx.Value)
	ret, leftOverGas, err = env.Call(
		tx.From,
		*tx.To,
		tx.Data,
		tx.Gas,
		value,
	)
	if err != nil {
		return
	}

	txHash = e.roll(ctx, tx, leftOverGas, executionDB, tracer)
	return
}

func (e *SerialExecutor) roll(
	ctx context.Context,
	msg ethereum.CallMsg,
	left uint64,
	executionDB *statedb.StateDB,
	traceProvider entity.TraceProvider,
) *common.Hash {
	e.db.ApplyStorage(executionDB.Dirty().GetAccountStorage())
	e.db.ApplyState(executionDB.Dirty().GetAccountState())

	nonce, err := e.db.GetNonce(ctx, msg.From)
	if err != nil {
		fmt.Println(err)
		nonce = 0
	}
	if err = e.db.SetNonce(ctx, msg.From, nonce+1); err != nil {
		fmt.Println(err)
	}

	tx := producer.NewTransactionContext(nonce+1, msg)

	hash, block, err := producer.MineBlockWithSingleTransaction(
		tx,
		left,
		new(big.Int).SetUint64(e.prevBlockNum),
		e.prevBlockHash,
		executionDB.Dirty(),
		e.db,
		e.txn, e.blocks)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	e.prevBlockHash = hash
	e.prevBlockNum = block.Uint64()
	e.txn.AddTrace(tx.Hash(), traceProvider.OtterTrace())

	txHash := tx.Hash()
	return &txHash
}

func (e *SerialExecutor) Call(
	ctx context.Context,
	tx ethereum.CallMsg,
	tracer entity.TraceProvider,
	overrides entity.StateOverrides,
) (ret []byte, leftOverGas uint64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	executionDB := statedb.NewDB(ctx, e.db)
	if err = executionDB.ApplyOverrides(overrides); err != nil {
		return nil, 0, err
	}

	chainCfg, evmCfg := e.cfg.ExecutionConfig(tracer.Hooks())
	env := vm.NewEVM(e.cfg.BlockContext(new(big.Int).Add(e.cfg.ForkConfig.ForkBlock, new(big.Int).SetUint64(1)),
		new(big.Int),
		uint64(time.Now().Unix())),
		executionDB,
		chainCfg,
		evmCfg)
	value, _ := uint256.FromBig(tx.Value)
	ret, leftOverGas, err = env.Call(
		tx.From,
		*tx.To,
		tx.Data,
		tx.Gas,
		value,
	)

	return
}

func (e *SerialExecutor) CallWithDB(
	ctx context.Context,
	tx ethereum.CallMsg,
	tracer entity.TraceProvider,
	db *fork.DB,
	overrides entity.StateOverrides,
) (ret []byte, leftOverGas uint64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	executionDB := statedb.NewDB(ctx, db)
	if err = executionDB.ApplyOverrides(overrides); err != nil {
		return nil, 0, err
	}

	chainCfg, evmCfg := e.cfg.ExecutionConfig(tracer.Hooks())
	env := vm.NewEVM(e.cfg.BlockContext(new(big.Int).Add(e.cfg.ForkConfig.ForkBlock, new(big.Int).SetUint64(1)),
		new(big.Int),
		uint64(time.Now().Unix())),
		executionDB,
		chainCfg,
		evmCfg)
	value, _ := uint256.FromBig(tx.Value)
	ret, leftOverGas, err = env.Call(
		tx.From,
		*tx.To,
		tx.Data,
		tx.Gas,
		value,
	)

	return
}

func (e *SerialExecutor) TxnStorage() *entity.TransactionStorage {
	return e.txn
}

func (e *SerialExecutor) BlockStorage() *entity.BlockStorage {
	return e.blocks
}

func (e *SerialExecutor) Latest() (common.Hash, uint64) {
	return e.prevBlockHash, e.prevBlockNum
}
