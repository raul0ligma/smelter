package entity

import "math/big"

type ForkConfig struct {
	ChainID   uint64   `json:"chainId"`
	ForkBlock *big.Int `json:"forkBlock"`
}
