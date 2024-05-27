package entity

import (
	"maps"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ByteCode []byte

func (b ByteCode) MarshalJSON() ([]byte, error) {
	return []byte(hexutil.Encode(b)), nil
}

type AccountStorage struct {
	Code        []byte                      `json:"code"`
	Initialized bool                        `json:"initialized"`
	Slots       map[common.Hash]common.Hash `json:"slots"`
}

type AccountsStorage struct {
	mu   sync.RWMutex
	data map[common.Address]*AccountStorage
}

func NewAccountsStorage() *AccountsStorage {
	return &AccountsStorage{
		data: make(map[common.Address]*AccountStorage),
	}
}

func (a *AccountsStorage) State(addr common.Address) *AccountStorage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	src, ok := a.data[addr]
	if !ok || !src.Initialized {
		return nil
	}

	dst := make([]byte, len(src.Code))
	copy(dst, src.Code)

	dstSlots := make(map[common.Hash]common.Hash)
	maps.Copy(dstSlots, src.Slots)

	return &AccountStorage{
		Code:        dst,
		Slots:       dstSlots,
		Initialized: true,
	}
}

func (a *AccountsStorage) ReadStorage(addr common.Address, key common.Hash) common.Hash {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return common.Hash{}
	}

	return s.Slots[key]
}

func (a *AccountsStorage) SetStorage(addr common.Address, key common.Hash, value common.Hash) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return
	}

	s.Slots[key] = value
}

func (a *AccountsStorage) SetCode(addr common.Address, code []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return
	}

	s.Code = code
}

func (a *AccountsStorage) GetCode(addr common.Address) []byte {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return nil
	}

	return s.Code
}

func (a *AccountsStorage) NewAccount(addr common.Address, code []byte) *AccountStorage {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if ok && s.Initialized {
		return s
	}

	storage := &AccountStorage{
		Code:        code,
		Initialized: true,
		Slots:       map[common.Hash]common.Hash{},
	}

	a.data[addr] = storage
	return storage
}

func (a *AccountsStorage) NewAccountWithStorage(
	addr common.Address,
	code []byte,
	slots map[common.Hash]common.Hash,
) *AccountStorage {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if ok && s.Initialized {
		return s
	}

	storage := &AccountStorage{
		Code:        code,
		Initialized: true,
		Slots:       slots,
	}

	a.data[addr] = storage
	return storage
}

func (a *AccountsStorage) Apply(s *AccountsStorage) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for addr, storage := range s.data {
		existing, ok := a.data[addr]
		if !ok {
			a.data[addr] = storage
			continue
		}

		existing.Code = storage.Code
		for k, v := range storage.Slots {
			existing.Slots[k] = v
		}
	}
}

type AccountState struct {
	Address     common.Address
	Balance     *big.Int
	Nonce       uint64
	Initialized bool
	// we ignore code codeHash storageRoot as we directly read them from AccountsStorage
}

type AccountsState struct {
	mu   sync.RWMutex
	data map[common.Address]*AccountState
}

func NewAccountsState() *AccountsState {
	return &AccountsState{
		data: make(map[common.Address]*AccountState),
	}
}

func (a *AccountsState) Exists(addr common.Address) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return false
	}

	return true
}

func (a *AccountsState) State(addr common.Address) *AccountState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	src, ok := a.data[addr]
	if !ok || !src.Initialized {
		return nil
	}

	return &AccountState{
		Address:     src.Address,
		Balance:     new(big.Int).Set(src.Balance),
		Nonce:       src.Nonce,
		Initialized: true,
	}
}

func (a *AccountsState) SetBalance(addr common.Address, bal *big.Int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return
	}

	s.Balance = bal
}

func (a *AccountsState) GetBalance(addr common.Address) *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return nil
	}

	return s.Balance
}

func (a *AccountsState) GetNonce(addr common.Address) uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return 0
	}

	return s.Nonce
}

func (a *AccountsState) SetNonce(addr common.Address, nonce uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if !ok || !s.Initialized {
		return
	}

	s.Nonce = nonce
}

func (a *AccountsState) NewAccount(addr common.Address, nonce uint64, bal *big.Int) *AccountState {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.data[addr]
	if ok && s.Initialized {
		return s
	}

	state := &AccountState{
		Address:     addr,
		Balance:     bal,
		Nonce:       nonce,
		Initialized: true,
	}

	a.data[addr] = state
	return state
}

func (a *AccountsState) Apply(s *AccountsState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for addr, storage := range s.data {
		existing, ok := a.data[addr]
		if !ok {
			a.data[addr] = storage
			continue
		}

		existing.Balance = storage.Balance
		existing.Nonce = storage.Nonce
	}
}
