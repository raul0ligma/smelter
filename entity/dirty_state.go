package entity

import "github.com/ethereum/go-ethereum/core/types"

type DirtyState struct {
	accountStorage *AccountsStorage
	accountState   *AccountsState
	logs           LogStorage
}

func NewDirtyState() *DirtyState {
	return &DirtyState{
		accountStorage: NewAccountsStorage(),
		accountState:   NewAccountsState(),
		logs:           make(LogStorage, 0),
	}
}

func (ds *DirtyState) GetAccountStorage() *AccountsStorage {
	return ds.accountStorage
}

func (ds *DirtyState) GetAccountState() *AccountsState {
	return ds.accountState
}

func (ds *DirtyState) AddLog(log *types.Log) {
	ds.logs = append(ds.logs, log)
}

func (ds *DirtyState) Logs() LogStorage {
	return ds.logs
}
