package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rahul0tripathi/smelter/entity"
)

func getBalanceFromForkDB(ctx context.Context, forkDB forkDB, account common.Address) (*string, error) {
	bal, err := forkDB.GetBalance(ctx, account)
	if err != nil {
		return nil, err
	}
	balStr := bal.String()
	return &balStr, nil
}

func parseBlockNumber(blockNumber string) (*big.Int, error) {
	numBytes, err := hexutil.Decode(blockNumber)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(numBytes), nil
}

func getBalanceFromBlockStorage(executor executor, account common.Address, blockNum uint64) (*string, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return nil, fmt.Errorf("block %d not found in block storage", blockNum)
	}
	balAt := b.State.GetBalance(account)
	balStr := "0x0"
	if balAt != nil {
		balStr = balAt.String()
	}
	return &balStr, nil
}

func getBalanceFromReader(
	ctx context.Context,
	reader entity.ChainStateReader,
	account common.Address,
	block *big.Int,
) (*string, error) {
	at, err := reader.BalanceAt(ctx, account, block)
	if err != nil {
		return nil, err
	}
	balStr := at.String()
	return &balStr, nil
}
