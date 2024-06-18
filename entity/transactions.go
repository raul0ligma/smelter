package entity

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	types2 "github.com/rahul0tripathi/smelter/types"
	"github.com/rahul0tripathi/smelter/utils"
)

type SerializedTransaction struct {
	From             common.Address `json:"from"`
	BlockHash        common.Hash    `json:"blockHash"`
	BlockNumber      string         `json:"blockNumber"`
	ChainId          string         `json:"chainId"`
	Confirmations    uint64         `json:"confirmations"`
	Creates          common.Address `json:"creates"`
	Data             string         `json:"data"`
	Gas              string         `json:"gas"`
	GasLimit         uint64         `json:"gasLimit"`
	GasPrice         string         `json:"gasPrice"`
	Hash             common.Hash    `json:"hash"`
	Nonce            string         `json:"nonce"`
	R                string         `json:"r"`
	S                string         `json:"s"`
	V                string         `json:"v"`
	TransactionIndex uint           `json:"transactionIndex"`
	Type             string         `json:"type"`
	Value            string         `json:"value"`
	Input            string         `json:"input"`
}

type SerializedReceipt struct {
	From common.Address `json:"from"`
	types.Receipt
}

func (sr *SerializedReceipt) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	receiptData, err := json.Marshal(sr.Receipt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(receiptData, &data); err != nil {
		return nil, err
	}

	data["from"] = sr.From
	data["type"] = 0
	data["timestamp"] = time.Now().Unix()

	return json.Marshal(data)
}

func SerializeReceipt(r *types.Receipt) *SerializedReceipt {
	return &SerializedReceipt{
		From:    types2.Address0x1,
		Receipt: *r,
	}
}

func SerializeTransaction(tx *types.Transaction, receipt *types.Receipt) *SerializedTransaction {
	if tx == nil || receipt == nil {
		return nil
	}

	return &SerializedTransaction{
		From:             types2.Address0x1,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      utils.Big2Hex(receipt.BlockNumber),
		ChainId:          hexutil.Encode(tx.ChainId().Bytes()),
		Confirmations:    1,
		Creates:          common.HexToAddress(""),
		Data:             hexutil.Encode(tx.Data()),
		Input:            hexutil.Encode(tx.Data()),
		GasLimit:         tx.Gas(),
		GasPrice:         utils.Big2Hex(tx.GasPrice()),
		Hash:             tx.Hash(),
		Nonce:            hexutil.EncodeUint64(tx.Nonce()),
		R:                utils.Big2Hex(nil),
		S:                utils.Big2Hex(nil),
		V:                utils.Big2Hex(nil),
		Gas:              hexutil.EncodeUint64(tx.Gas()),
		TransactionIndex: receipt.TransactionIndex,
		Type:             hexutil.EncodeUint64(uint64(tx.Type())),
		Value:            utils.Big2Hex(tx.Value()),
	}
}

type TransactionStorage struct {
	mu       sync.RWMutex
	txs      map[common.Hash]*types.Transaction
	receipts map[common.Hash]*types.Receipt
	traces   map[common.Hash]TransactionTraces
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

func (ts *TransactionStorage) AddTrace(hash common.Hash, trace TransactionTraces) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.traces[hash] = trace
}

func (ts *TransactionStorage) GetTrace(hash common.Hash) TransactionTraces {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.traces[hash]
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

func (ts *TransactionStorage) All() []*types.Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	txns := make([]*types.Transaction, 0)
	for _, tx := range ts.txs {
		txns = append(txns, tx)
	}

	return txns
}

type LogStorage []*types.Log
