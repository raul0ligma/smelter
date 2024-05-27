package config

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/rahul0tripathi/smelter/vm"
)

type Config struct {
	ChainConfig *params.ChainConfig
	Difficulty  *big.Int
	Origin      common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        uint64
	GasLimit    uint64
	GasPrice    *big.Int
	Value       *big.Int
	Debug       bool
	EVMConfig   vm.Config
	BaseFee     *big.Int
	BlobBaseFee *big.Int
	BlobHashes  []common.Hash
	BlobFeeCap  *big.Int
	Random      *common.Hash

	State     *state.StateDB
	GetHashFn func(n uint64) common.Hash
}

func (c *Config) TxCtx(origin common.Address) vm.TxContext {
	return vm.TxContext{
		Origin:     origin,
		GasPrice:   c.GasPrice,
		BlobHashes: c.BlobHashes,
		BlobFeeCap: c.BlobFeeCap,
	}
}

func (c *Config) BlockContext(blockNumber *big.Int, baseFee *big.Int, time uint64) vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: vm.CanTransfer,
		Transfer:    vm.Transfer,
		GetHash: func(u uint64) common.Hash {
			return common.HexToHash(fmt.Sprintf("%d", u))
		},
		Coinbase:    c.Coinbase,
		BlockNumber: blockNumber,
		Time:        time,
		Difficulty:  c.Difficulty,
		GasLimit:    c.GasLimit,
		BaseFee:     baseFee,
		BlobBaseFee: c.BlobBaseFee,
	}
}

func (c *Config) ExecutionConfig(tracer *tracing.Hooks) (*params.ChainConfig, vm.Config) {
	c.EVMConfig.Tracer = tracer
	return c.ChainConfig, c.EVMConfig
}

func setDefaults(cfg *Config) *Config {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			ChainID:                       big.NewInt(1),
			HomesteadBlock:                new(big.Int),
			DAOForkBlock:                  new(big.Int),
			DAOForkSupport:                false,
			EIP150Block:                   new(big.Int),
			EIP155Block:                   new(big.Int),
			EIP158Block:                   new(big.Int),
			ByzantiumBlock:                new(big.Int),
			ConstantinopleBlock:           new(big.Int),
			PetersburgBlock:               new(big.Int),
			IstanbulBlock:                 new(big.Int),
			MuirGlacierBlock:              new(big.Int),
			BerlinBlock:                   new(big.Int),
			LondonBlock:                   new(big.Int),
			TerminalTotalDifficultyPassed: true,
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxUint64
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
	if cfg.BaseFee == nil {
		cfg.BaseFee = big.NewInt(params.InitialBaseFee)
	}
	if cfg.BlobBaseFee == nil {
		cfg.BlobBaseFee = big.NewInt(params.BlobTxMinBlobGasprice)
	}

	return cfg
}

func NewConfigWithDefaults() *Config {
	return setDefaults(&Config{})
}
