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

type BlockStorage struct {
	storage  map[common.Hash]*types.Block
	num2Hash map[uint64]common.Hash
	latest   uint64
	mu       sync.Mutex
}

func NewBlockStorage() *BlockStorage {
	return &BlockStorage{
		storage:  make(map[common.Hash]*types.Block),
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

func (b *BlockStorage) GetBlockByNumber(block uint64) *types.Block {
	b.mu.Lock()
	defer b.mu.Unlock()
	num, ok := b.num2Hash[block]
	if !ok {
		return nil
	}

	return b.storage[num]
}

func (b *BlockStorage) GetBlockByHash(hash common.Hash) *types.Block {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.storage[hash]
}

func (b *BlockStorage) AddBlock(block *types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, exists := b.num2Hash[block.NumberU64()]
	if exists {
		return
	}

	b.storage[block.Hash()] = block
	b.num2Hash[block.NumberU64()] = block.Hash()

	if b.latest < block.NumberU64() {
		b.latest = block.NumberU64()
	}
}

func (b *BlockStorage) Apply(s *BlockStorage) {
	for _, v := range s.storage {
		b.AddBlock(v)
	}
}
