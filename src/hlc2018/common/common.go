package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Account struct {
	ID            int    `db:"id"`
	Fname         string `db:"fname"`
	Sname         string `db:"sname"`
	Email         string `db:"email"`
	Status        int8   `db:"status"`
	Premium_start int    `db:"premium_start"`
	Premium_end   int    `db:"premium_start"`
	Sex           int8   `db:"sex"`
	Phone         string `db:"phone"`
	Birth         int    `db:"birth"`
	City          string `db:"city"`
	Country       string `db:"country"`
	Joined        int    `db:"joined"`
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

func sliceIndex(s []string, val string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == val {
			return i
		}
	}
	return -1
}

func StatusFromString(s string) int8 {
	return int8(sliceIndex([]string{"свободны", "заняты", "всё сложно"}, s))
}

func SexFromString(s string) int8 {
	return int8(sliceIndex([]string{"m", "f"}, s))
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
