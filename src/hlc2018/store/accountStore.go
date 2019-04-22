package store

import (
	"fmt"
	"hlc2018/common"
)

type StoredAccount struct {
	ID            int
	Fname         string
	Sname         string
	Email         string
	Premium_start int
	Premium_end   int
	Premium_now   bool
	Status        int8
	Sex           int8
	Phone         string
	Birth         int
	City          string
	Country       string
	JoinedYear    int8
}

type AccountStore struct {
	accounts []*StoredAccount
}

func NewAccountStore() *AccountStore {
	return &AccountStore{}
}

func (as *AccountStore) ExtendSizeIfNeeded(nextSize int) {
	for len(as.accounts) < nextSize {
		as.accounts = append(as.accounts, nil)
	}
}

func (as *AccountStore) InsertAccountCommon(a *common.Account) error {
	as.ExtendSizeIfNeeded(a.ID + 1)
	if as.accounts[a.ID] != nil {
		return fmt.Errorf("failed to add a new account : %d is already used", a.ID)
	}

	nw := &StoredAccount{
		ID:            a.ID,
		Fname:         a.Fname,
		Sname:         a.Sname,
		Email:         a.Email,
		Premium_start: a.Premium_start,
		Premium_end:   a.Premium_end,
		Premium_now:   a.Premium_now,
		Status:        a.Status,
		Sex:           a.Sex,
		Phone:         a.Phone,
		Birth:         a.Birth,
		City:          a.City,
		Country:       a.Country,
		JoinedYear:    a.JoinedYear.Int8,
	}
	as.accounts[a.ID] = nw

	return nil
}

func (as *AccountStore) NewRangeAccountStoreSource() *RangeStoreSource {
	return NewRangeStoreSource(len(as.accounts), 0, -1)
}

func (as *AccountStore) GetStoredAccount(id int) (*StoredAccount, error) {
	if len(as.accounts) <= id {
		return nil, fmt.Errorf("account not found")
	}
	return as.accounts[id], nil
}

func (as *AccountStore) GetStoredAccountWithoutError(id int) *StoredAccount {
	return as.accounts[id]
}
