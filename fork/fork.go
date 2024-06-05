package fork

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rahul0tripathi/smelter/entity"
)

type DB struct {
	stateReader    ethereum.ChainStateReader
	config         entity.ForkConfig
	accountStorage *entity.AccountsStorage
	accountState   *entity.AccountsState
}

func NewDB(
	stateReader ethereum.ChainStateReader,
	config entity.ForkConfig,
	accountStorage *entity.AccountsStorage,
	accountState *entity.AccountsState,
) *DB {
	return &DB{
		stateReader:    stateReader,
		config:         config,
		accountStorage: accountStorage,
		accountState:   accountState,
	}
}

func (db *DB) CreateState(ctx context.Context, addr common.Address) error {
	if db.accountState.Exists(addr) {
		return nil
	}

	code, err := db.stateReader.CodeAt(ctx, addr, db.config.ForkBlock)
	if err != nil {
		return err
	}

	bal, err := db.stateReader.BalanceAt(ctx, addr, db.config.ForkBlock)
	if err != nil {
		return err
	}

	nonce, err := db.stateReader.NonceAt(ctx, addr, db.config.ForkBlock)
	if err != nil {
		return err
	}

	db.accountState.NewAccount(addr, nonce, bal)
	db.accountStorage.NewAccount(addr, code)
	return nil
}

func (db *DB) State(ctx context.Context, addr common.Address) (*entity.AccountState, *entity.AccountStorage, error) {
	if err := db.CreateState(ctx, addr); err != nil {
		return nil, nil, err
	}

	return db.accountState.State(addr), db.accountStorage.State(addr), nil
}

func (db *DB) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	if err := db.CreateState(ctx, addr); err != nil {
		return nil, err
	}

	return db.accountState.GetBalance(addr), nil
}

func (db *DB) SetBalance(ctx context.Context, addr common.Address, amount *big.Int) error {
	if err := db.CreateState(ctx, addr); err != nil {
		return err
	}

	db.accountState.SetBalance(addr, amount)
	return nil
}

func (db *DB) GetNonce(ctx context.Context, addr common.Address) (uint64, error) {
	if err := db.CreateState(ctx, addr); err != nil {
		return 0, err
	}

	return db.accountState.GetNonce(addr), nil
}

func (db *DB) SetNonce(ctx context.Context, addr common.Address, nonce uint64) error {
	if err := db.CreateState(ctx, addr); err != nil {
		return err
	}
	db.accountState.SetNonce(addr, nonce)
	return nil
}

func (db *DB) GetCodeHash(ctx context.Context, addr common.Address) (common.Hash, error) {
	code, err := db.GetCode(ctx, addr)
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(code), nil
}

func (db *DB) GetCode(ctx context.Context, addr common.Address) ([]byte, error) {
	if err := db.CreateState(ctx, addr); err != nil {
		return nil, err
	}

	return db.accountStorage.GetCode(addr), nil // Placeholder return
}

func (db *DB) GetCodeSize(ctx context.Context, addr common.Address) (int, error) {
	code, err := db.GetCode(ctx, addr)
	if err != nil {
		return 0, err
	}

	return len(code), nil
}

func (db *DB) GetState(ctx context.Context, addr common.Address, hash common.Hash) (common.Hash, error) {
	if err := db.CreateState(ctx, addr); err != nil {
		return common.Hash{}, err
	}
	emptyHash := common.Hash{}
	val := db.accountStorage.ReadStorage(addr, hash)
	if val != emptyHash {
		return val, nil
	}

	raw, err := db.stateReader.StorageAt(ctx, addr, hash, db.config.ForkBlock)
	if err != nil {
		return common.Hash{}, err
	}

	db.accountStorage.SetStorage(addr, hash, common.BytesToHash(raw))
	return common.BytesToHash(raw), nil
}

func (db *DB) ApplyState(s *entity.AccountsState) {
	db.accountState.Apply(s)
}

func (db *DB) ApplyStorage(s *entity.AccountsStorage) {
	db.accountStorage.Apply(s)
}

func (db *DB) Copy() (*entity.AccountsStorage, *entity.AccountsState) {
	return entity.NewAccountsStorageWitStorage(db.accountStorage.Clone()), entity.NewAccountsStateWithStorage(db.accountState.Clone())
}
