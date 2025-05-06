package statedb

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/raul0ligma/smelter/entity"
)

var (
	emptyHash   = common.Hash{}
	emptyKeccak = common.Hash{
		0xc5, 0xd2, 0x46, 0x01, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x03, 0xc0,
		0xe5, 0x00, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x04, 0x5d, 0x85, 0xa4, 0x70,
	}
)

type snapshot struct {
	storage entity.AccountsStorageCache
	state   entity.AccountStateStorage
}

type StateDB struct {
	ctx        context.Context
	db         forkDB
	dirty      *entity.DirtyState
	errorStack []error
	snapshots  map[uint64]snapshot
	counter    uint64
}

func NewDB(ctx context.Context, db forkDB) *StateDB {
	return &StateDB{
		ctx:        ctx,
		db:         db,
		dirty:      entity.NewDirtyState(),
		errorStack: make([]error, 0),
		snapshots:  map[uint64]snapshot{},
	}
}

func (s *StateDB) load(addr common.Address) error {
	if s.dirty.GetAccountState().Exists(addr) {
		return nil
	}

	loaded, storage, err := s.db.State(s.ctx, addr)
	if err != nil {
		return err
	}

	s.dirty.GetAccountState().NewAccount(loaded.Address, loaded.Nonce, loaded.Balance)
	s.dirty.GetAccountStorage().NewAccountWithStorage(addr, storage.Code, storage.Slots)
	return nil
}

func (s *StateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	s.errorStack = append(s.errorStack, errors.New("unimplemented GetTransientState()"))
	return [32]byte{}
}

func (s *StateDB) SetTransientState(addr common.Address, key, value common.Hash) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented SetTransientState()"))
}

func (s *StateDB) CreateAccount(addr common.Address) {
	if err := s.db.CreateState(s.ctx, addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("CreateAccount: %w", err))
	}

}

func (s *StateDB) CreateContract(addr common.Address) {
	s.CreateAccount(addr)
}

func (s *StateDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("SubBalance: %w", err))
		return *uint256.NewInt(0)
	}

	s.dirty.GetAccountState().SetBalance(
		addr, new(big.Int).Sub(s.dirty.GetAccountState().GetBalance(addr), amount.ToBig()),
	)
	return *uint256.NewInt(0)
}

func (s *StateDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("AddBalance: %w", err))
		return *uint256.NewInt(0)
	}

	s.dirty.GetAccountState().SetBalance(
		addr, new(big.Int).Add(s.dirty.GetAccountState().GetBalance(addr), amount.ToBig()),
	)
	return *uint256.NewInt(0)
}

func (s *StateDB) Finalise(bool) {}

func (s *StateDB) GetBalance(addr common.Address) *uint256.Int {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("GetBalance: %w", err))
		return nil
	}

	return uint256.MustFromBig(s.dirty.GetAccountState().GetBalance(addr))
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("GetNonce: %w", err))
		return 0
	}

	return s.dirty.GetAccountState().GetNonce(addr)
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("SetNonce: %w", err))
		return
	}

	s.dirty.GetAccountState().SetNonce(addr, nonce)
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	return crypto.Keccak256Hash(s.GetCode(addr))
}

func (s *StateDB) GetCode(addr common.Address) []byte {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("GetCode: %w", err))
		return nil
	}

	return s.dirty.GetAccountStorage().GetCode(addr)
}

func (s *StateDB) SetCode(addr common.Address, code []byte) []byte {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("SetCode: %w", err))
		return nil
	}

	s.dirty.GetAccountStorage().SetCode(addr, code)
	return code
}

func (s *StateDB) GetCodeSize(addr common.Address) int {
	return len(s.GetCode(addr))
}

func (s *StateDB) AddRefund(gas uint64) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented AddRefund()"))
}

func (s *StateDB) SubRefund(gas uint64) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented SubRefund()"))
}

func (s *StateDB) GetRefund() uint64 {
	s.errorStack = append(s.errorStack, errors.New("unimplemented GetRefund()"))
	return 0
}

func (s *StateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	s.errorStack = append(s.errorStack, errors.New("unimplemented GetCommittedState()"))
	return common.Hash{}
}

func (s *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("GetState: %w", err))
		return common.Hash{}
	}

	storage := s.dirty.GetAccountStorage().ReadStorage(addr, hash)
	if storage != emptyHash {
		return storage
	}

	slot, err := s.db.GetState(s.ctx, addr, hash)
	if err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("GetStateSlot: %w", err))
		return common.Hash{}
	}

	s.dirty.GetAccountStorage().SetStorage(addr, hash, slot)
	return slot
}

func (s *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("SetState: %w", err))
		return value
	}

	s.dirty.GetAccountStorage().SetStorage(addr, key, value)
	return value
}

func (s *StateDB) GetStorageRoot(addr common.Address) common.Hash {
	s.errorStack = append(s.errorStack, errors.New("unimplemented GetStorageRoot()"))
	return common.Hash{}
}

func (s *StateDB) SelfDestruct(addr common.Address) uint256.Int {
	s.errorStack = append(s.errorStack, errors.New("unimplemented SelfDestruct()"))
	return uint256.Int{}
}

func (s *StateDB) HasSelfDestructed(addr common.Address) bool {
	s.errorStack = append(s.errorStack, errors.New("unimplemented HasSelfDestructed()"))
	return false
}

func (s *StateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented GetStorageRoot()"))
	return uint256.Int{}, false
}

func (s *StateDB) Exist(addr common.Address) bool {
	return !s.Empty(addr)
}

func (s *StateDB) Empty(addr common.Address) bool {
	if err := s.load(addr); err != nil {
		s.errorStack = append(s.errorStack, fmt.Errorf("empty: %w", err))
		return true
	}
	return (s.dirty.GetAccountState().GetBalance(addr) == nil || s.dirty.GetAccountState().GetBalance(addr).Uint64() == 0) &&
		s.dirty.GetAccountState().GetNonce(addr) == 0 &&
		(s.GetCodeHash(addr) == emptyKeccak || s.GetCodeHash(addr) == emptyHash)
}

func (s *StateDB) AddressInAccessList(addr common.Address) bool {
	s.errorStack = append(s.errorStack, errors.New("unimplemented AddressInAccessList()"))
	return false // Placeholder return
}

func (s *StateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented SlotInAccessList()"))
	return false, false // Placeholder return
}

func (s *StateDB) AddAddressToAccessList(addr common.Address) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented AddAddressToAccessList()"))
}

func (s *StateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented AddSlotToAccessList()"))
}

// Snapshot methods
func (s *StateDB) Prepare(
	rules params.Rules,
	sender, coinbase common.Address,
	dest *common.Address,
	precompiles []common.Address,
	txAccesses types.AccessList,
) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented Prepare()"))
}

func (s *StateDB) RevertToSnapshot(id int) {
	cached, ok := s.snapshots[uint64(id)]
	if !ok {
		s.errorStack = append(s.errorStack, fmt.Errorf("failed to revert to snapshot %d", id))
		return
	}

	s.dirty.GetAccountState().Set(cached.state)
	s.dirty.GetAccountStorage().Set(cached.storage)
}

func (s *StateDB) Snapshot() int {
	atomic.AddUint64(&s.counter, 1)
	s.snapshots[s.counter] = snapshot{
		storage: s.dirty.GetAccountStorage().Clone(),
		state:   s.dirty.GetAccountState().Clone(),
	}

	return int(s.counter) // Placeholder return
}

func (s *StateDB) AddLog(log *types.Log) {
	s.dirty.AddLog(log)
}

func (s *StateDB) AddPreimage(hash common.Hash, data []byte) {
	s.errorStack = append(s.errorStack, errors.New("unimplemented AddPreimage()"))
}

func (s *StateDB) AccessEvents() *state.AccessEvents {
	return nil
}

func (s *StateDB) PointCache() *utils.PointCache {
	s.errorStack = append(s.errorStack, errors.New("unimplemented PointCache()"))
	return nil // Placeholder return
}

func (s *StateDB) Dirty() *entity.DirtyState {
	return s.dirty
}

func (s *StateDB) Witness() *stateless.Witness {
	return nil
}

func (s *StateDB) ApplyOverrides(overrides entity.StateOverrides) error {
	for addr, override := range overrides {
		if err := s.load(addr); err != nil {
			return err
		}

		if override.Balance != nil {
			s.dirty.GetAccountState().SetBalance(addr, override.Balance)
		}

		if len(override.Code) != 0 {
			s.dirty.GetAccountStorage().SetCode(addr, override.Code)
		}

		for k, v := range override.Storage {
			s.dirty.GetAccountStorage().SetStorage(addr, k, v)
		}
	}

	return nil
}
