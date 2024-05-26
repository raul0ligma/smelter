package entity

import "github.com/ethereum/go-ethereum"

type ChainStateReader interface {
	ethereum.BlockNumberReader
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.ChainIDReader
}
