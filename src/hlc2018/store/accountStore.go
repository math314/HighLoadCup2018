package store

import (
	"fmt"
	"hlc2018/common"
	"strconv"
)

type CompressedPhone struct {
	Int int
}

func CompressedPhoneFromString(phone string) (CompressedPhone, error) {
	if phone == "" {
		return CompressedPhone{0}, nil
	}
	use := phone[3:5] + phone[6:]
	ret, err := strconv.Atoi(use)
	return CompressedPhone{ret}, err
}

const K_PHONE = 10000000

func (cp CompressedPhone) String() string {
	if cp.Int == 0 {
		return ""
	}
	return fmt.Sprintf("8(9%02d)%07d", cp.Int/K_PHONE, cp.Int%K_PHONE)
}

func (cp CompressedPhone) HasPhoneCode(code int) bool {
	if cp.Int == 0 {
		return false
	}
	cpCode := 900 + cp.Int/K_PHONE
	return cpCode == code
}

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
	Phone         CompressedPhone
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

	cp, err := CompressedPhoneFromString(a.Phone)
	if err != nil {
		return err
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
		Phone:         cp,
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
