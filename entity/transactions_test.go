package entity

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestAddAndGetTransaction(t *testing.T) {
	storage := &TransactionStorage{
		txs:      make(map[common.Hash]*types.Transaction),
		receipts: make(map[common.Hash]*types.Receipt),
	}

	tx := types.NewTransaction(1, common.HexToAddress("0x0"), nil, 21000, nil, nil)
	storage.AddTransaction(tx)

	retrievedTx := storage.GetTransaction(tx.Hash())
	if retrievedTx == nil {
		t.Fatalf("Expected transaction, got nil")
	}
	if retrievedTx.Hash() != tx.Hash() {
		t.Fatalf("Expected transaction hash %v, got %v", tx.Hash(), retrievedTx.Hash())
	}
}

func TestAddAndGetReceipt(t *testing.T) {
	storage := &TransactionStorage{
		txs:      make(map[common.Hash]*types.Transaction),
		receipts: make(map[common.Hash]*types.Receipt),
	}

	receipt := &types.Receipt{TxHash: common.HexToHash("0x1")}
	storage.AddReceipt(receipt)

	retrievedReceipt := storage.GetReceipt(receipt.TxHash)
	if retrievedReceipt == nil {
		t.Fatalf("Expected receipt, got nil")
	}
	if retrievedReceipt.TxHash != receipt.TxHash {
		t.Fatalf("Expected receipt hash %v, got %v", receipt.TxHash, retrievedReceipt.TxHash)
	}
}
