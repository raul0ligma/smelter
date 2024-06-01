package producer

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rahul0tripathi/smelter/entity"
)

func NewTransactionContext(nonce uint64, msg ethereum.CallMsg) *types.Transaction {
	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: msg.GasPrice,
		Gas:      msg.Gas,
		To:       msg.To,
		Value:    msg.Value,
		Data:     msg.Data,
	})
}

func MineBlockWithSignleTransaction(
	tx *types.Transaction,
	left uint64,
	prevBlockNumber *big.Int,
	prevBlockHash common.Hash,
	db postExecutionStateFetcher,
	fork forkDB,
	txStore transactionStorage,
	blockStore blockStorage,
) (common.Hash, *big.Int, error) {
	blockNumber := new(big.Int).Add(prevBlockNumber, new(big.Int).SetUint64(1))
	receipt := &types.Receipt{
		Type:              tx.Type(),
		Status:            1,
		CumulativeGasUsed: tx.Gas() - left,
		// TODO: create logs bloom
		Bloom:             types.Bloom{},
		Logs:              db.Logs(),
		TxHash:            tx.Hash(),
		ContractAddress:   *tx.To(),
		GasUsed:           tx.Gas() - left,
		EffectiveGasPrice: tx.GasPrice(),
		BlockNumber:       blockNumber,
		TransactionIndex:  0,
	}

	block := entity.NewBlock(prevBlockHash, blockNumber, types.Transactions{tx}, types.Receipts{receipt})
	receipt.BlockHash = block.Hash()
	txStore.AddTransaction(tx)
	txStore.AddReceipt(receipt)

	accounts, state := fork.Copy()
	blockStore.AddBlock(&entity.BlockState{
		Accounts: accounts,
		State:    state,
		Block:    block,
	})

	return block.Hash(), blockNumber, nil
}
