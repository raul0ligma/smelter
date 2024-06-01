package entity

import (
	"hash"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

type legacyHasher struct {
	hasher hash.Hash
}

// NewHasher returns a new testHasher instance.
func newHasher() *legacyHasher {
	return &legacyHasher{hasher: sha3.NewLegacyKeccak256()}
}

// Reset resets the hash state.
func (h *legacyHasher) Reset() {
	h.hasher.Reset()
}

// Update updates the hash state with the given key and value.
func (h *legacyHasher) Update(key, val []byte) error {
	h.hasher.Write(key)
	h.hasher.Write(val)
	return nil
}

// Hash returns the hash value.
func (h *legacyHasher) Hash() common.Hash {
	return common.BytesToHash(h.hasher.Sum(nil))
}

func NewBlock(
	prevBlockHash common.Hash,
	number *big.Int,
	transactions types.Transactions,
	receipts types.Receipts,
) *types.Block {
	header := &types.Header{
		ParentHash: prevBlockHash,
		Number:     number,
		GasLimit:   90000000,
	}

	b := types.NewBlock(header, &types.Body{
		Transactions: transactions,
	}, receipts, newHasher())

	return b
}

type BlockState struct {
	Accounts *AccountsStorage
	State    *AccountsState
	Block    *types.Block
}

type BlockStorage struct {
	storage  map[common.Hash]*BlockState
	num2Hash map[uint64]common.Hash
	latest   uint64
	mu       sync.Mutex
}

func NewBlockStorage() *BlockStorage {
	return &BlockStorage{
		storage:  make(map[common.Hash]*BlockState),
		num2Hash: make(map[uint64]common.Hash),
		latest:   0,
	}
}

func (b *BlockStorage) Exists(block uint64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.num2Hash[block]
	return ok
}

func (b *BlockStorage) GetBlockByNumber(block uint64) *BlockState {
	b.mu.Lock()
	defer b.mu.Unlock()
	num, ok := b.num2Hash[block]
	if !ok {
		return nil
	}

	return b.storage[num]
}

func (b *BlockStorage) GetBlockByHash(hash common.Hash) *BlockState {
	b.mu.Lock()
	defer b.mu.Unlock()
	state, ok := b.storage[hash]
	if !ok {
		return nil
	}

	return state
}

func (b *BlockStorage) AddBlock(state *BlockState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, exists := b.num2Hash[state.Block.NumberU64()]
	if exists {
		return
	}

	b.storage[state.Block.Hash()] = state
	b.num2Hash[state.Block.NumberU64()] = state.Block.Hash()

	if b.latest < state.Block.NumberU64() {
		b.latest = state.Block.NumberU64()
	}
}

func (b *BlockStorage) Latest() uint64 {
	return b.latest
}

func (b *BlockStorage) Apply(s *BlockStorage) {
	for _, v := range s.storage {
		b.AddBlock(v)
	}
}
