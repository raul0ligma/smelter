package entity

import (
	"hash"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	internal "github.com/raul0ligma/smelter/types"
	"github.com/raul0ligma/smelter/utils"
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
	GasUsed uint64,
) *types.Block {
	header := &types.Header{
		ParentHash: prevBlockHash,
		Number:     number,
		GasLimit:   90000000,
		Time:       uint64(time.Now().Unix()),
		GasUsed:    GasUsed,
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

type SerializedBlock struct {
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parentHash"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	Miner            string        `json:"miner"`
	StateRoot        string        `json:"stateRoot"`
	TransactionsRoot string        `json:"transactionsRoot"`
	ReceiptsRoot     string        `json:"receiptsRoot"`
	LogsBloom        string        `json:"logsBloom"`
	Difficulty       string        `json:"difficulty"`
	Number           string        `json:"number"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Timestamp        string        `json:"timestamp"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	ExtraData        string        `json:"extraData"`
	MixHash          string        `json:"mixHash"`
	Nonce            string        `json:"nonce"`
	BaseFeePerGas    string        `json:"baseFeePerGas"`
	WithdrawalsRoot  string        `json:"withdrawalsRoot"`
	Uncles           []string      `json:"uncles"`
	Transactions     []string      `json:"transactions"`
	Size             string        `json:"size"`
	Withdrawals      []interface{} `json:"withdrawals"`
	Raw              *types.Block  `json:"-"`
}

func SerializeBlock(block *types.Block) *SerializedBlock {
	if block == nil {
		return nil
	}
	txs := make([]string, 0)
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash().Hex())
	}

	return &SerializedBlock{
		Hash:             block.Hash().Hex(),
		ParentHash:       block.ParentHash().Hex(),
		Sha3Uncles:       block.UncleHash().Hex(),
		Miner:            internal.Address0xSmelter.Hex(),
		StateRoot:        block.Root().Hex(),
		TransactionsRoot: common.HexToHash("").Hex(),
		ReceiptsRoot:     block.ReceiptHash().Hex(),
		LogsBloom:        hexutil.Encode(block.Bloom().Bytes()),
		Difficulty:       utils.Big2Hex(block.Difficulty()),
		Number:           utils.Big2Hex(block.Number()),
		GasLimit:         utils.ToEvenLength(hexutil.EncodeUint64(block.GasLimit())),
		GasUsed:          utils.ToEvenLength(hexutil.EncodeUint64(block.GasUsed())),
		Timestamp:        utils.ToEvenLength(hexutil.EncodeUint64(block.Time())),
		TotalDifficulty:  utils.Big2Hex(block.Difficulty()),
		ExtraData:        "0x",
		MixHash:          common.HexToHash("").Hex(),
		Nonce:            hexutil.EncodeUint64(block.Nonce()),
		BaseFeePerGas:    utils.Big2Hex(block.BaseFee()),
		WithdrawalsRoot:  common.HexToHash("").Hex(),
		Uncles:           []string{},
		Transactions:     txs,
		Size:             utils.ToEvenLength(hexutil.EncodeUint64(block.Size())),
		Withdrawals:      make([]interface{}, 0),
		Raw:              block,
	}
}
