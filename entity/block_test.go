package entity

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	prevBlockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	number := big.NewInt(1)

	// Create a sample transaction
	tx := types.NewTransaction(0, common.HexToAddress("0x0"), big.NewInt(0), 21000, big.NewInt(1), nil)
	transactions := types.Transactions{tx}

	// Create sample logs
	log1 := &types.Log{
		Address: common.HexToAddress("0x0"),
		Topics:  []common.Hash{common.HexToHash("0x1")},
		Data:    []byte("log1"),
	}
	log2 := &types.Log{
		Address: common.HexToAddress("0x0"),
		Topics:  []common.Hash{common.HexToHash("0x2")},
		Data:    []byte("log2"),
	}

	// Create a sample receipt with the transaction hash and logs
	receipt := &types.Receipt{
		Status:            1,
		CumulativeGasUsed: 21000,
		TxHash:            tx.Hash(),
		Logs:              []*types.Log{log1, log2},
	}
	receipts := types.Receipts{receipt}

	block := NewBlock(prevBlockHash, number, transactions, receipts)

	assert.Equal(t, prevBlockHash, block.ParentHash(), "ParentHash should match")
	assert.Equal(t, number, block.Number(), "Block number should match")
	assert.Equal(t, uint64(90000000), block.GasLimit(), "GasLimit should be 90000000")
	assert.Len(t, block.Transactions(), 1, "Transactions length should be 1")
	assert.Equal(t, tx.Hash(), common.HexToHash("0x7b8da361b3612a2e3e416ffbf702964254d7c922f5e27bb1e6ea9aeb2303636e"), "TxHash should match transaction hash")
	assert.Equal(t, block.ReceiptHash(), common.HexToHash("0x6adaed440f50dfc22347f76edba73c671aba5394e9967febce9bf2b8a6434034"), "ReceiptHash should match transaction hash")
	assert.Equal(t, block.Hash(), common.HexToHash("0x875a4eec469db94329b494c76a8989a0371bf987e91454daff615652a938913b"), "BlockHash should match transaction hash")

}

func TestBlockStorage_AddAndGetBlock(t *testing.T) {
	storage := NewBlockStorage()

	// Create a new block
	prevHash := common.HexToHash("0x0")
	blockNumber := big.NewInt(1)
	transactions := types.Transactions{}
	receipts := types.Receipts{}
	block := NewBlock(prevHash, blockNumber, transactions, receipts)

	// Add the block to storage
	storage.AddBlock(block)

	// Retrieve the block by number
	retrievedBlock := storage.GetBlockByNumber(block.NumberU64())
	if retrievedBlock == nil {
		t.Fatalf("Expected block, got nil")
	}
	if retrievedBlock.Hash() != block.Hash() {
		t.Fatalf("Expected block hash %s, got %s", block.Hash().Hex(), retrievedBlock.Hash().Hex())
	}

	// Retrieve the block by hash
	retrievedBlockByHash := storage.GetBlockByHash(block.Hash())
	if retrievedBlockByHash == nil {
		t.Fatalf("Expected block, got nil")
	}
	if retrievedBlockByHash.Hash() != block.Hash() {
		t.Fatalf("Expected block hash %s, got %s", block.Hash().Hex(), retrievedBlockByHash.Hash().Hex())
	}
}

func TestBlockStorage_Exists(t *testing.T) {
	storage := NewBlockStorage()

	// Create a new block
	prevHash := common.HexToHash("0x0")
	blockNumber := big.NewInt(1)
	transactions := types.Transactions{}
	receipts := types.Receipts{}
	block := NewBlock(prevHash, blockNumber, transactions, receipts)

	// Add the block to storage
	storage.AddBlock(block)

	// Check if the block exists
	if !storage.Exists(block.NumberU64()) {
		t.Fatalf("Expected block to exist")
	}

	// Check if a non-existent block exists
	if storage.Exists(2) {
		t.Fatalf("Expected block to not exist")
	}
}

func TestBlockStorage_AddBlock_Concurrency(t *testing.T) {
	storage := NewBlockStorage()

	// Create a new block
	prevHash := common.HexToHash("0x0")
	blockNumber := big.NewInt(1)
	transactions := types.Transactions{}
	receipts := types.Receipts{}
	block := NewBlock(prevHash, blockNumber, transactions, receipts)

	// Add the block to storage concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			storage.AddBlock(block)
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check if the block exists
	if !storage.Exists(block.NumberU64()) {
		t.Fatalf("Expected block to exist")
	}
}
