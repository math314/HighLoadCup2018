package store

import (
	"fmt"
	"hlc2018/common"
)

type storedAccount struct {
	ID              int
	Fname           string
	Sname           string
	Email           string
	Premium_start   int
	Premium_end     int
	Premium_now     bool
	Status          int8
	Sex             int8
	CompressedPhone int
	Birth           int
	City            string
	Country         string
	JoinedYear      int8
}

type AccountStore struct {
	accounts []*storedAccount
}

func (as *AccountStore) ExtendSizeIfNeeded(nextSize int) {
	for len(as.accounts) < nextSize {
		as.accounts = append(as.accounts, nil)
	}
}

func (as *AccountStore) addAccountCommon(account *common.Account) error {
	as.ExtendSizeIfNeeded(account.ID + 1)
	if as.accounts[account.ID] != nil {
		return fmt.Errorf("failed to add a new account : %d is already used", account.ID)
	}
	return nil
}
