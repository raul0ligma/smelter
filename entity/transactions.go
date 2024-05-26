package entity

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactionStorage struct {
	mu       sync.RWMutex
	txs      map[common.Hash]*types.Transaction
	receipts map[common.Hash]*types.Receipt
}

func NewTransactionStorage() *TransactionStorage {
	return &TransactionStorage{
		txs:      make(map[common.Hash]*types.Transaction),
		receipts: make(map[common.Hash]*types.Receipt),
	}
}

// AddTransaction adds a transaction to the storage.
func (ts *TransactionStorage) AddTransaction(tx *types.Transaction) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.txs[tx.Hash()] = tx
}

// GetTransaction retrieves a transaction by its hash.
func (ts *TransactionStorage) GetTransaction(hash common.Hash) *types.Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.txs[hash]
}

// AddReceipt adds a receipt to the storage.
func (ts *TransactionStorage) AddReceipt(receipt *types.Receipt) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.receipts[receipt.TxHash] = receipt
}

// GetReceipt retrieves a receipt by its transaction hash.
func (ts *TransactionStorage) GetReceipt(hash common.Hash) *types.Receipt {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.receipts[hash]
}

func (ts *TransactionStorage) Apply(s *TransactionStorage) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for hash, v := range s.txs {
		ts.txs[hash] = v
	}

	for hash, v := range s.receipts {
		ts.receipts[hash] = v
	}
}
