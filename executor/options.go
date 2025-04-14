package executor

import "github.com/ethereum/go-ethereum/common"

type Option func(*SerialExecutor)

func WithPreviousState(blockHash common.Hash, prevBlockNum uint64) Option {
	return func(se *SerialExecutor) {
		se.prevBlockHash = blockHash
		se.prevBlockNum = prevBlockNum
	}
}
