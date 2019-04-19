package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type RawPremium struct {
	Start  int `json:"start"`
	Finish int `json:"finish"`
}

type RawAccount struct {
	ID        int         `json:"id,omitempty"`
	Fname     string      `json:"fname,omitempty"`
	Sname     string      `json:"sname,omitempty"`
	Email     string      `json:"email,omitempty"`
	Interests []string    `json:"interests,omitempty"`
	Status    string      `json:"status,omitempty"`
	Premium   *RawPremium `json:"premium,omitempty"`
	Sex       string      `json:"sex,omitempty"`
	Phone     string      `json:"phone,omitempty"`
	Likes     []struct {
		Ts int `json:"ts"`
		ID int `json:"id"`
	} `json:"likes,omitempty"`
	Birth   int    `json:"birth,omitempty"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
	Joined  int    `json:"joined,omitempty"`
}

type RawAccountsContainer struct {
	Accounts []RawAccount `json:"accounts"`
}

type Account struct {
	ID            int    `db:"id"`
	Fname         string `db:"fname"`
	Sname         string `db:"sname"`
	Email         string `db:"email"`
	Status        int8   `db:"status"`
	Premium_start int    `db:"premium_start"`
	Premium_end   int    `db:"premium_end"`
	Sex           int8   `db:"sex"`
	Phone         string `db:"phone"`
	Birth         int    `db:"birth"`
	City          string `db:"city"`
	Country       string `db:"country"`
	Joined        int    `db:"joined"`
}

type AccountContainer struct {
	Accounts []Account
}

type Interest struct {
	AccountId int
	Interest  string
}

type Like struct {
	AccountIdFrom int
	AccountIdTo   int
	Ts            int
}

func SliceIndex(s []string, val string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == val {
			return i
		}
	}
	return -1
}

var STATUSES = []string{"свободны", "заняты", "всё сложно"}
var SEXES = []string{"m", "f"}

func StatusFromString(s string) int8 {
	return int8(SliceIndex(STATUSES, s) + 1)
}

func SexFromString(s string) int8 {
	return int8(SliceIndex(SEXES, s) + 1)
}

func (rawAccount *RawAccount) ToAccount() Account {
	var a Account
	a.ID = rawAccount.ID
	a.Fname = rawAccount.Fname
	a.Sname = rawAccount.Sname
	a.Email = rawAccount.Email
	a.Status = StatusFromString(rawAccount.Status)
	if rawAccount.Premium != nil {
		a.Premium_start = rawAccount.Premium.Start
		a.Premium_end = rawAccount.Premium.Finish
	}
	a.Sex = SexFromString(rawAccount.Sex)
	a.Phone = rawAccount.Phone
	a.Birth = rawAccount.Birth
	a.City = rawAccount.City
	a.Country = rawAccount.Country
	a.Joined = rawAccount.Joined

	return a
}

func (rawAccount *RawAccount) ToInterests() []Interest {
	var interests []Interest
	for _, i := range rawAccount.Interests {
		interest := Interest{rawAccount.ID, i}
		interests = append(interests, interest)
	}

	return interests
}

func (rawAccount *RawAccount) ToLikes() []Like {
	var likes []Like
	for _, l := range rawAccount.Likes {
		like := Like{rawAccount.ID, l.ID, l.Ts}
		likes = append(likes, like)
	}

	return likes
}

func (l *RawAccount) Equal(r *RawAccount) bool {
	if l.ID != r.ID {
		return false
	}
	if l.Fname != r.Fname {
		return false
	}
	if l.Sname != r.Sname {
		return false
	}
	if l.Email != r.Email {
		return false
	}
	if l.Status != r.Status {
		return false
	}
	if l.Premium != nil && r.Premium != nil {
		if l.Premium.Start != r.Premium.Start {
			return false
		}
		if l.Premium.Finish != r.Premium.Finish {
			return false
		}
	} else if l.Premium != nil || r.Premium != nil {
		return false
	}
	if l.Sex != r.Sex {
		return false
	}
	if l.Phone != r.Phone {
		return false
	}
	if l.Birth != r.Birth {
		return false
	}
	if l.City != r.City {
		return false
	}
	if l.Country != r.Country {
		return false
	}
	if l.Joined != r.Joined {
		return false
	}

	return true
}

func (a *Account) ToRawAccount() RawAccount {
	r := RawAccount{}
	r.ID = a.ID
	r.Fname = a.Fname
	r.Sname = a.Sname
	r.Email = a.Email
	r.Status = ""
	if a.Status != 0 {
		r.Status = STATUSES[a.Status-1]
	}
	if a.Premium_start != 0 {
		r.Premium = &RawPremium{a.Premium_start, a.Premium_end}
	}
	r.Sex = ""
	if a.Sex != 0 {
		r.Sex = SEXES[a.Sex-1]
	}
	r.Phone = a.Phone
	r.Birth = a.Birth
	r.City = a.City
	r.Country = a.Country
	r.Joined = a.Joined

	return r
}

func (ac *AccountContainer) ToRawAccountsContainer() RawAccountsContainer {
	rac := RawAccountsContainer{[]RawAccount{}}
	for _, a := range ac.Accounts {
		rac.Accounts = append(rac.Accounts, a.ToRawAccount())
	}

	return rac
}

type oneLineBuilder struct {
	b strings.Builder
}

func (o *oneLineBuilder) appendString(s string) {
	if o.b.Len() != 0 {
		o.b.WriteString(",")
	}
	if s == "" {
		o.b.WriteString("\\N")
	} else {
		if strings.Contains(s, "\"") {
			log.Fatal(fmt.Printf("%s contains \"", s))
		}
		o.b.WriteString(s)
	}
}

func (o *oneLineBuilder) appendInt(i int) {
	if o.b.Len() != 0 {
		o.b.WriteString(",")
	}
	o.b.WriteString(strconv.Itoa(i))
}

func (o *oneLineBuilder) build() string {
	o.b.WriteString("\n")
	return o.b.String()
}

func (a *Account) Oneline() string {
	olb := oneLineBuilder{strings.Builder{}}
	olb.appendInt(a.ID)
	olb.appendString(a.Email)
	olb.appendString(a.Fname)
	olb.appendString(a.Sname)
	olb.appendString(a.Phone)
	olb.appendInt(int(a.Sex))
	olb.appendInt(a.Birth)
	olb.appendString(a.Country)
	olb.appendString(a.City)
	olb.appendInt(a.Joined)
	olb.appendInt(int(a.Status))
	olb.appendInt(a.Premium_start)
	olb.appendInt(a.Premium_end)
	return olb.build()
}

func (i *Interest) Oneline() string {
	olb := oneLineBuilder{strings.Builder{}}
	olb.appendInt(i.AccountId)
	olb.appendString(i.Interest)
	return olb.build()
}

func (l *Like) Oneline() string {
	olb := oneLineBuilder{strings.Builder{}}
	olb.appendInt(l.AccountIdFrom)
	olb.appendInt(l.AccountIdTo)
	olb.appendInt(l.Ts)
	return olb.build()
}
