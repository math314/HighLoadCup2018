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
	City          int
	Country       int
	JoinedYear    common.JoinedYear
}

type AccountStore struct {
	countryIndex *StringIndex
	cityIndex    *StringIndex
	accounts     []*StoredAccount
	emailToPK    map[string]int
}

func NewAccountStore() *AccountStore {
	return &AccountStore{
		countryIndex: NewStringIndex(),
		cityIndex:    NewStringIndex(),
		accounts:     nil,
		emailToPK:    map[string]int{},
	}
}

func (as *AccountStore) GetCountryId(country string) int {
	return as.countryIndex.ConvertStringToStringId(country)
}

func (as *AccountStore) IdToCountry(id int) string {
	return as.countryIndex.StringIdToString(id)
}

func (as *AccountStore) GetCityId(country string) int {
	return as.cityIndex.ConvertStringToStringId(country)
}

func (as *AccountStore) IdToCity(id int) string {
	return as.cityIndex.StringIdToString(id)
}

func (as *AccountStore) ExtendSizeIfNeeded(nextSize int) {
	for len(as.accounts) < nextSize {
		as.accounts = append(as.accounts, nil)
	}
}

func (as *AccountStore) InsertAccountCommon(a *common.Account) error {
	if a.ID == 0 {
		return fmt.Errorf("id is not provided")
	}
	as.ExtendSizeIfNeeded(a.ID + 1)
	if as.accounts[a.ID] != nil {
		return fmt.Errorf("failed to add a new account : %d is already used", a.ID)
	}
	if other, found := as.emailToPK[a.Email]; found {
		return fmt.Errorf("email is already registered. %d is using. your id : %d", other, a.ID)
	}

	cp, err := CompressedPhoneFromString(a.Phone)
	if err != nil {
		return err
	}
	cityCode := as.cityIndex.SetString(a.ID, a.City)
	countryCode := as.countryIndex.SetString(a.ID, a.Country)
	as.emailToPK[a.Email] = a.ID

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
		City:          cityCode,
		Country:       countryCode,
		JoinedYear:    a.JoinedYear,
	}
	as.accounts[a.ID] = nw

	return nil
}

func (as *AccountStore) UpdateAccountCommon(a *common.Account) error {
	if a.ID >= len(as.accounts) || as.accounts[a.ID] == nil {
		return fmt.Errorf("%d is already not registered yet", a.ID)
	}

	// if a.Fname != ""
	// if a.Sname != ""

	if a.Email != "" {
		if other, found := as.emailToPK[a.Email]; found {
			return fmt.Errorf("email is already registered. %d is using. your id : %d", other, a.ID)
		}
	}

	//Premium_start: a.Premium_start,
	//Premium_end:   a.Premium_end,
	//Premium_now:   a.Premium_now,
	//Status:        a.Status,
	//Sex:           a.Sex,

	cp, err := CompressedPhoneFromString(a.Phone)
	if err != nil {
		return err
	}

	//Birth:         a.Birth,
	//City:          cityCode,
	//Country:       countryCode,
	//JoinedYear:    a.JoinedYear,

	cityCode := as.cityIndex.SetString(a.ID, a.City)
	countryCode := as.countryIndex.SetString(a.ID, a.Country)

	me := as.accounts[a.ID]
	if a.Fname != "" {
		me.Fname = a.Fname
	}
	if a.Sname != "" {
		me.Sname = a.Sname
	}
	if a.Email != "" {
		delete(as.emailToPK, me.Email)
		me.Email = a.Email
		as.emailToPK[me.Email] = me.ID
	}
	if a.Premium_start != 0 {
		me.Premium_start = a.Premium_start
		me.Premium_end = a.Premium_end
		me.Premium_now = a.Premium_now
	}
	if a.Status != 0 {
		me.Status = a.Status
	}
	if a.Sex != 0 {
		me.Sex = a.Sex
	}
	if a.Phone != "" {
		me.Phone = cp
	}
	if a.Birth != 0 {
		me.Birth = a.Birth
	}
	if a.City != "" {
		as.cityIndex.DeleteStringsFromPk(me.ID)
		me.City = as.cityIndex.ConvertStringToStringId(a.City)
		as.cityIndex.SetString(me.ID, a.City)
	}
	if a.Country != "" {
		as.countryIndex.DeleteStringsFromPk(me.ID)
		me.Country = as.countryIndex.ConvertStringToStringId(a.Country)
		as.countryIndex.SetString(me.ID, a.Country)
	}
	if a.JoinedYear.Int8 != 0 {
		me.JoinedYear = a.JoinedYear
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
		City:          cityCode,
		Country:       countryCode,
		JoinedYear:    a.JoinedYear,
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
