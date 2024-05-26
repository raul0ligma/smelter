package entity

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestAccountsStorage(t *testing.T) {
	addr := common.HexToAddress("0x0")
	key := common.HexToHash("0x1")
	value := common.HexToHash("0x2")
	code := []byte{0x60, 0x60, 0x60, 0x40}

	storage := NewAccountsStorage()
	storage.NewAccount(addr, code)

	storage.SetStorage(addr, key, value)
	assert.Equal(t, value, storage.ReadStorage(addr, key))

	storage.SetCode(addr, code)
	assert.Equal(t, code, storage.GetCode(addr))
}

func TestAccountsState(t *testing.T) {
	addr := common.HexToAddress("0x0")
	balance := big.NewInt(1000)
	nonce := uint64(1)

	state := NewAccountsState()
	state.NewAccount(addr, nonce, balance)

	state.SetBalance(addr, big.NewInt(2000))
	assert.Equal(t, big.NewInt(2000), state.GetBalance(addr))

	state.SetNonce(addr, nonce+1)
	assert.Equal(t, nonce+1, state.GetNonce(addr))
}
