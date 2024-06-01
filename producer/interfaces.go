package producer

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rahul0tripathi/smelter/entity"
)

type transactionStorage interface {
	AddTransaction(tx *types.Transaction)
	AddReceipt(receipt *types.Receipt)
}

type postExecutionStateFetcher interface {
	Logs() entity.LogStorage
}

type blockStorage interface {
	AddBlock(block *entity.BlockState)
}

type forkDB interface {
	Copy() (*entity.AccountsStorage, *entity.AccountsState)
}
