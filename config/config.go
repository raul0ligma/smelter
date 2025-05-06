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
	"github.com/holiman/uint256"
	"github.com/raul0ligma/smelter/entity"
	"github.com/raul0ligma/smelter/vm"
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

	State      *state.StateDB
	GetHashFn  func(n uint64) common.Hash
	ForkConfig *entity.ForkConfig
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
		CanTransfer: func(db vm.StateDB, addr common.Address, amount *uint256.Int) bool {
			return db.GetBalance(addr).Cmp(amount) >= 0
		},
		Transfer: func(db vm.StateDB, sender, recipient common.Address, amount *uint256.Int) {
			db.SubBalance(sender, amount, tracing.BalanceChangeTransfer)
			db.AddBalance(recipient, amount, tracing.BalanceChangeTransfer)
		},
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
		Random:      &common.Hash{},
	}
}

func (c *Config) ExecutionConfig(tracer *tracing.Hooks) (
	*params.ChainConfig,
	vm.Config,
) {
	c.EVMConfig.Tracer = tracer
	return c.ChainConfig, c.EVMConfig
}

func newUint64(val uint64) *uint64 { return &val }

func setDefaults(cfg *Config) *Config {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			HomesteadBlock:          new(big.Int).SetInt64(0),
			DAOForkBlock:            new(big.Int).SetInt64(0),
			DAOForkSupport:          true,
			EIP150Block:             new(big.Int).SetInt64(0),
			EIP155Block:             new(big.Int).SetInt64(0),
			EIP158Block:             new(big.Int).SetInt64(0),
			ByzantiumBlock:          new(big.Int).SetInt64(0),
			ConstantinopleBlock:     new(big.Int).SetInt64(0),
			PetersburgBlock:         new(big.Int).SetInt64(0),
			IstanbulBlock:           new(big.Int).SetInt64(0),
			MuirGlacierBlock:        new(big.Int).SetInt64(0),
			BerlinBlock:             new(big.Int).SetInt64(0),
			LondonBlock:             new(big.Int).SetInt64(0),
			TerminalTotalDifficulty: params.MainnetChainConfig.TerminalTotalDifficulty,
			ShanghaiTime:            newUint64(0),
			CancunTime:              newUint64(0),
			// enables eip 7702
			PragueTime: newUint64(0),
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = cfg.ChainConfig.TerminalTotalDifficulty
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxBig256.Uint64()
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
