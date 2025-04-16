package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ForkConfig struct {
	ChainID   uint64   `json:"chainId"`
	ForkBlock *big.Int `json:"forkBlock"`
}

type Slot struct {
	Addr  common.Address
	Key   common.Hash
	Value []byte
}

type Slots []Slot
