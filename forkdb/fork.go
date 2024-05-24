package forkdb

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

type State struct {
	Balance *big.Int
	Code    []byte
	Nonce   uint64
	Storage map[common.Hash]common.Hash
}

type ForkDB struct {
	client     *ethclient.Client
	ctx        context.Context
	dirtyState map[common.Address]State
}

func (f *ForkDB) Dump() {
	v, _ := json.MarshalIndent(f.dirtyState, "", "    ")
	fmt.Println(string(v))
}

func NewForkDB(ctx context.Context, cli string) (*ForkDB, error) {
	client, err := ethclient.Dial(cli)
	if err != nil {
		return nil, err
	}

	return &ForkDB{
		client:     client,
		ctx:        ctx,
		dirtyState: make(map[common.Address]State),
	}, nil
}

func (db *ForkDB) getOrCreateState(address common.Address) State {
	if state, ok := db.dirtyState[address]; ok {
		return state
	}
	state := State{
		Balance: nil,
		Code:    nil,
		Nonce:   0,
		Storage: make(map[common.Hash]common.Hash),
	}
	db.dirtyState[address] = state

	return state
}

func (db *ForkDB) setState(address common.Address, state State) {
	db.dirtyState[address] = state
}

func (db *ForkDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	//TODO implement me
	panic("implement me")
}

func (db *ForkDB) SetTransientState(addr common.Address, key, value common.Hash) {
	//TODO implement me
	panic("implement me")
}

// Account related methods
func (db *ForkDB) CreateAccount(addr common.Address) {
	fmt.Println("unimplemented CreateAccount()")
}

func (db *ForkDB) CreateContract(addr common.Address) {
	fmt.Println("unimplemented CreateContract()")
}

// Balance related methods
func (db *ForkDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	state := db.getOrCreateState(addr)
	if state.Balance == nil {
		state.Balance = db.GetBalance(addr).ToBig()
	}

	state.Balance = new(big.Int).Sub(state.Balance, amount.ToBig())
	db.setState(addr, state)
}

func (db *ForkDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	state := db.getOrCreateState(addr)
	if state.Balance == nil {
		state.Balance = db.GetBalance(addr).ToBig()
	}

	state.Balance = new(big.Int).Add(state.Balance, amount.ToBig())
	db.setState(addr, state)
}

func (db *ForkDB) GetBalance(addr common.Address) *uint256.Int {
	state := db.getOrCreateState(addr)
	if state.Balance != nil {
		return uint256.MustFromBig(state.Balance)
	}

	bal, err := db.client.BalanceAt(db.ctx, addr, nil)
	if err != nil {
		panic(err)
	}
	state.Balance = bal
	db.setState(addr, state)
	return uint256.MustFromBig(bal) // Placeholder return
}

// Nonce related methods
func (db *ForkDB) GetNonce(addr common.Address) uint64 {
	fmt.Println("StateDB: GetNonce: ", addr)
	state := db.getOrCreateState(addr)
	if state.Nonce != 0 {
		return state.Nonce
	}

	nonce, err := db.client.NonceAt(db.ctx, addr, nil)
	if err != nil {
		panic(err)
	}
	state.Nonce = nonce
	db.setState(addr, state)
	return nonce // Placeholder return
}

func (db *ForkDB) SetNonce(addr common.Address, nonce uint64) {
	fmt.Println("unimplemented SetNonce()")
}

// Code related methods
func (db *ForkDB) GetCodeHash(addr common.Address) common.Hash {
	return crypto.Keccak256Hash(db.GetCode(addr)) // Placeholder return
}

func (db *ForkDB) GetCode(addr common.Address) []byte {
	fmt.Println("StateDB: GetCode: ", addr)
	state := db.getOrCreateState(addr)
	if len(state.Code) != 0 {
		return state.Code
	}

	code, err := db.client.CodeAt(db.ctx, addr, nil)
	if err != nil {
		panic(err)
	}
	state.Code = code
	db.setState(addr, state)
	return code // Placeholder return
}

func (db *ForkDB) SetCode(addr common.Address, code []byte) {
	fmt.Println("StateDB: SetCode: ", addr, code)
	state := db.getOrCreateState(addr)
	state.Code = code
	db.setState(addr, state)
}

func (db *ForkDB) GetCodeSize(addr common.Address) int {
	return len(db.GetCode(addr))
}

// Refund related methods
func (db *ForkDB) AddRefund(gas uint64) {
	fmt.Println("unimplemented AddRefund()")
}

func (db *ForkDB) SubRefund(gas uint64) {
	fmt.Println("unimplemented SubRefund()")
}

func (db *ForkDB) GetRefund() uint64 {
	fmt.Println("unimplemented GetRefund()")
	return 0 // Placeholder return
}

// State related methods
func (db *ForkDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	// Implement the method logic
	return common.Hash{} // Placeholder return
}

func (db *ForkDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	fmt.Println("StateDB: GetState: ", addr)
	state := db.getOrCreateState(addr)
	if val, ok := state.Storage[hash]; ok {
		return val
	}

	str, err := db.client.StorageAt(db.ctx, addr, hash, nil)
	if err != nil {
		panic(err)
	}

	state.Storage[hash] = common.Hash(str)
	db.setState(addr, state)
	return state.Storage[hash] // Placeholder return
}

func (db *ForkDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	fmt.Println("StateDB: SetState: ", addr)
	state := db.getOrCreateState(addr)
	state.Storage[key] = value
	db.setState(addr, state)
}

func (db *ForkDB) GetStorageRoot(addr common.Address) common.Hash {
	fmt.Println("unimplemented AddressInAccessList()")
	return common.Hash{} // Placeholder return
}

// Self-destruct related methods
func (db *ForkDB) SelfDestruct(addr common.Address) {
	fmt.Println("unimplemented AddressInAccessList()")
}

func (db *ForkDB) HasSelfDestructed(addr common.Address) bool {
	fmt.Println("unimplemented AddressInAccessList()")
	return false // Placeholder return
}

func (db *ForkDB) Selfdestruct6780(addr common.Address) {
	fmt.Println("unimplemented AddressInAccessList()")
}

// Existence and emptiness checks
func (db *ForkDB) Exist(addr common.Address) bool {
	db.getOrCreateState(addr)
	return true // Placeholder return
}

func (db *ForkDB) Empty(addr common.Address) bool {
	fmt.Println("unimplemented AddressInAccessList()")
	return true // Placeholder return
}

// Access list methods
func (db *ForkDB) AddressInAccessList(addr common.Address) bool {
	fmt.Println("unimplemented AddressInAccessList()")
	return false // Placeholder return
}

func (db *ForkDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	fmt.Println("unimplemented AddAddressToAccessList()")
	return false, false // Placeholder return
}

func (db *ForkDB) AddAddressToAccessList(addr common.Address) {
	fmt.Println("unimplemented AddAddressToAccessList()")
}

func (db *ForkDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	fmt.Println("unimplemented AddSlotToAccessList()")
}

// Snapshot methods
func (db *ForkDB) Prepare(
	rules params.Rules,
	sender, coinbase common.Address,
	dest *common.Address,
	precompiles []common.Address,
	txAccesses types.AccessList,
) {
	// Implement the method logic
}

func (db *ForkDB) RevertToSnapshot(id int) {
	fmt.Println("RevertToSnapshot(id int)")
}

func (db *ForkDB) Snapshot() int {
	fmt.Println("unimplemented Snapshot()")
	return 0 // Placeholder return
}

// Log and preimage methods
func (db *ForkDB) AddLog(log *types.Log) {
	fmt.Println("unimplemented AddLog()")
}

func (db *ForkDB) AddPreimage(hash common.Hash, data []byte) {
	fmt.Println("unimplemented AddPreimage()")
}

// Point cache method
func (db *ForkDB) PointCache() *utils.PointCache {
	fmt.Println("unimplemented PointCache()")
	return nil // Placeholder return
}
