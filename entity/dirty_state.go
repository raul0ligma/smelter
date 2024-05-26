package entity

type DirtyState struct {
	accountStorage    *AccountsStorage
	accountState      *AccountsState
	transactionsState *TransactionStorage
	blockStorage      *BlockStorage
}

func NewDirtyState() *DirtyState {
	return &DirtyState{
		accountStorage:    NewAccountsStorage(),
		accountState:      NewAccountsState(),
		transactionsState: NewTransactionStorage(),
		blockStorage:      NewBlockStorage(),
	}
}

func (ds *DirtyState) GetAccountStorage() *AccountsStorage {
	return ds.accountStorage
}

func (ds *DirtyState) GetAccountState() *AccountsState {
	return ds.accountState
}

func (ds *DirtyState) GetTransactionsState() *TransactionStorage {
	return ds.transactionsState
}

func (ds *DirtyState) GetBlockStorage() *BlockStorage {
	return ds.blockStorage
}
