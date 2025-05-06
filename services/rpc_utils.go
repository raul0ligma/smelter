package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/raul0ligma/smelter/entity"
	"github.com/raul0ligma/smelter/fork"
)

const (
	hexPrefix   = "0x"
	latestBlock = "latest"
)

func parseAndValidateBlockNumber(blockNumber string, latest uint64) (*big.Int, error) {
	if blockNumber == "" || blockNumber == latestBlock {
		return new(big.Int).SetUint64(latest), nil
	}

	if len(blockNumber) > 2 && blockNumber[:2] == hexPrefix {
		block, err := parseBigInt(blockNumber)
		if err != nil {
			return nil, err
		}

		if block.Uint64() > latest {
			return nil, fmt.Errorf("invalid block height, received %d, current %d", block.Uint64(), latest)
		}
		return block, nil
	}

	return nil, fmt.Errorf("failed to parse block number %s", blockNumber)
}

func getBlockFromStorageOrReader(
	exec executor,
	reader entity.ChainStateAndTransactionReader,
	number uint64,
) (*entity.SerializedBlock, error) {
	storage := exec.BlockStorage().GetBlockByNumber(number)
	if storage != nil {
		return entity.SerializeBlock(storage.Block), nil
	}
	block, err := reader.BlockByNumber(context.Background(), new(big.Int).SetUint64(number))
	if err != nil {
		return nil, err
	}

	return entity.SerializeBlock(block), nil

}

func getBalanceFromForkDB(ctx context.Context, forkDB forkDB, account common.Address) (string, error) {
	bal, err := forkDB.GetBalance(ctx, account)
	if err != nil {
		return "0x", err
	}

	if bal.Uint64() == 0 {
		return "0x0", nil
	}

	return hexutil.Encode(bal.Bytes()), nil
}

func parseBigInt(blockNumber string) (*big.Int, error) {
	if len(blockNumber) > 2 && blockNumber[:2] != "0x" {
		num, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			return nil, err
		}

		return new(big.Int).SetUint64(num), nil
	}

	if len(blockNumber)%2 != 0 {
		blockNumber = blockNumber[:2] + "0" + blockNumber[2:]
	}

	numBytes, err := hexutil.Decode(blockNumber)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(numBytes), nil
}

func getBalanceFromBlockStorage(
	ctx context.Context,
	executor executor,
	chainID uint64,
	reader readerAndCaller,
	account common.Address,
	blockNum uint64,
) (string, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return "0x", fmt.Errorf("block %d not found in block storage", blockNum)
	}

	db := fork.NewDB(reader, entity.ForkConfig{
		ChainID:   chainID,
		ForkBlock: new(big.Int).SetUint64(blockNum),
	}, b.Accounts, b.State)

	balAt, err := db.GetBalance(ctx, account)
	if err != nil {
		return "0x0", err
	}

	return hexutil.Encode(balAt.Bytes()), nil
}

func getBalanceFromReader(
	ctx context.Context,
	reader entity.ChainStateAndTransactionReader,
	account common.Address,
	block *big.Int,
) (string, error) {
	at, err := reader.BalanceAt(ctx, account, block)
	if err != nil {
		return "0x", err
	}

	if at.Uint64() == 0 {
		return "0x0", nil
	}

	return hexutil.Encode(at.Bytes()), nil
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
	ctx context.Context,
	executor executor,
	chainID uint64,
	reader readerAndCaller,
	account common.Address,
	slot common.Hash,
	blockNum uint64,
) (common.Hash, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return common.Hash{}, fmt.Errorf("block %d not found in block storage", blockNum)
	}

	db := fork.NewDB(reader, entity.ForkConfig{
		ChainID:   chainID,
		ForkBlock: new(big.Int).SetUint64(blockNum),
	}, b.Accounts, b.State)

	storage, err := db.GetState(ctx, account, slot)
	if err != nil {
		return [32]byte{}, err
	}

	return storage, nil
}

func getStateFromReader(
	ctx context.Context,
	reader entity.ChainStateAndTransactionReader,
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

type jsonCallMsg struct {
	From      common.Address
	To        common.Address
	Gas       uint64
	GasPrice  string
	GasFeeCap string
	GasTipCap string
	Value     string
	Input     string
	Data      string
}

func (msg *jsonCallMsg) ToEthCallMsg() (call ethereum.CallMsg, err error) {
	call.Value = new(big.Int).SetUint64(0)
	if msg.Value != "" {
		var set bool
		call.Value, set = new(big.Int).SetString(msg.Value, 10)
		if !set {
			return call, errors.New("invalid value")
		}

	}

	call.Gas = 30e6
	call.From = msg.From
	call.To = &msg.To
	callData, err := hexutil.Decode(msg.Input)
	if err != nil {
		return call, err
	}

	call.Data = callData
	return
}

func getCodeFromReader(
	ctx context.Context,
	reader entity.ChainStateAndTransactionReader,
	account common.Address,
	block *big.Int,
) (string, error) {
	code, err := reader.CodeAt(ctx, account, block)
	if err != nil {
		return "0x", err
	}
	return hexutil.Encode(code), nil
}

func decodeHexString(encoded string) ([]byte, error) {
	return hexutil.Decode(encoded)
}

func createEthCallMsg(msg jsonCallMsg) (ethereum.CallMsg, error) {
	call := ethereum.CallMsg{
		From:  msg.From,
		To:    &msg.To,
		Gas:   30e6,
		Value: new(big.Int).SetInt64(0),
	}

	if msg.Value != "" {
		var set bool
		call.Value, set = new(big.Int).SetString(msg.Value, 10)
		if !set {
			return call, errors.New("invalid value")
		}
	}

	input := msg.Input
	if msg.Data != "" {
		input = msg.Data
	}

	callData, err := decodeHexString(input)
	if err != nil {
		return call, err
	}

	call.Data = callData
	return call, nil
}

func getBlockStorage(
	executor executor,
	blockNum uint64,
) (*entity.BlockState, error) {
	b := executor.BlockStorage().GetBlockByNumber(blockNum)
	if b == nil {
		return nil, fmt.Errorf("block %d not found in block storage", blockNum)
	}

	return b, nil
}

func callOnReader(
	ctx context.Context,
	caller ethereum.ContractCaller,
	msg ethereum.CallMsg,
	block *big.Int,
) (string, error) {
	msg.Gas = 5e6
	ret, err := caller.CallContract(ctx, msg, block)
	if err != nil {
		return "0x", err
	}

	return hexutil.Encode(ret), nil
}
