package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rahul0tripathi/smelter/entity"
)

func getBalanceFromForkDB(ctx context.Context, forkDB forkDB, account common.Address) (string, error) {
	bal, err := forkDB.GetBalance(ctx, account)
	if err != nil {
		return "0x", err
	}

	return bal.String(), nil
}

func parseBlockNumber(blockNumber string) (*big.Int, error) {
	numBytes, err := hexutil.Decode(blockNumber)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(numBytes), nil
}

func getBalanceFromBlockStorage(executor executor, account common.Address, blockNum uint64) (string, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return "0x", fmt.Errorf("block %d not found in block storage", blockNum)
	}

	if balAt := b.State.GetBalance(account); balAt != nil {
		return balAt.String(), nil
	}

	return "0x", nil
}

func getBalanceFromReader(
	ctx context.Context,
	reader entity.ChainStateReader,
	account common.Address,
	block *big.Int,
) (string, error) {
	at, err := reader.BalanceAt(ctx, account, block)
	if err != nil {
		return "0x", err
	}

	return at.String(), nil
}

func getCodeFromBlockStorage(executor executor, account common.Address, blockNum uint64) (string, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return "0x", fmt.Errorf("block %d not found in block storage", blockNum)
	}
	code := b.Accounts.GetCode(account)
	codeStr := "0x"
	if code != nil {
		codeStr = hexutil.Encode(code)
	}
	return codeStr, nil
}

func getStateFromBlockStorage(
	executor executor,
	account common.Address,
	slot common.Hash,
	blockNum uint64,
) (common.Hash, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return common.Hash{}, fmt.Errorf("block %d not found in block storage", blockNum)
	}
	storage := b.Accounts.ReadStorage(account, slot)

	return storage, nil
}

func getStateFromReader(
	ctx context.Context,
	reader entity.ChainStateReader,
	account common.Address,
	slot common.Hash,
	block *big.Int,
) (string, error) {
	at, err := reader.StorageAt(ctx, account, slot, block)
	if err != nil {
		return "0x", err
	}

	return hexutil.Encode(at), nil
}
