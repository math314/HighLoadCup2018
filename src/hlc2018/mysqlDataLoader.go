package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	db *sqlx.DB
)

func init() {
	db_name := os.Getenv("MYSQL_DATABASE")
	if db_name == "" {
		db_name = "hlc2018"
	}
	db_host := os.Getenv("MYSQL_HOST")
	if db_host == "" {
		db_host = "127.0.0.1"
	}
	db_port := os.Getenv("MYSQL_PORT")
	if db_port == "" {
		db_port = "3306"
	}
	db_user := os.Getenv("MYSQL_USER")
	if db_user == "" {
		db_user = "hlc"
	}
	db_password := os.Getenv("MYSQL_PASSWORD")
	if db_password == "" {
		db_password = "hlc"
	}
	db_password = ":" + db_password

	dsn := fmt.Sprintf("%s%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		db_user, db_password, db_host, db_port, db_name)

	log.Printf("Connecting to db: %q", dsn)
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Printf("Succeeded to connect db.")
}

type RawAccountsContainer struct {
	Accounts []struct {
		ID        int      `json:"id"`
		Fname     string   `json:"fname"`
		Sname     string   `json:"sname"`
		Email     string   `json:"email"`
		Interests []string `json:"interests"`
		Status    string   `json:"status"`
		Premium   struct {
			Start  int `json:"start"`
			Finish int `json:"finish"`
		} `json:"premium"`
		Sex   string `json:"sex"`
		Phone string `json:"phone"`
		Likes []struct {
			Ts int `json:"ts"`
			ID int `json:"id"`
		} `json:"likes"`
		Birth   int    `json:"birth"`
		City    string `json:"city"`
		Country string `json:"country"`
		Joined  int    `json:"joined"`
	} `json:"accounts"`
}

type Account struct {
	ID            int    `json:"id"`
	Fname         string `json:"fname"`
	Sname         string `json:"sname"`
	Email         string `json:"email"`
	Status        int8   `json:"status"`
	Premium_start int
	Premium_end   int
	Sex           int8   `json:"sex"`
	Phone         string `json:"phone"`
	Birth         int    `json:"birth"`
	City          string `json:"city"`
	Country       string `json:"country"`
	Joined        int    `json:"joined"`
}

type OneLineBuilder struct {
	b strings.Builder
}

func (o *OneLineBuilder) appendString(s string) {
	if o.b.Len() != 0 {
		o.b.WriteString(",")
	}
	if s == "" {
		o.b.WriteString("\\N")
	} else {
		o.b.WriteString("\"")
		o.b.WriteString(s)
		o.b.WriteString("\"")
	}
}

func (o *OneLineBuilder) appendInt(i int) {
	if o.b.Len() != 0 {
		o.b.WriteString(",")
	}
	o.b.WriteString(strconv.Itoa(i))
}

func (o *OneLineBuilder) build() string {
	o.b.WriteString("\n")
	return o.b.String()
}

func (a *Account) oneline() string {
	olb := OneLineBuilder{strings.Builder{}}
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

func sliceIndex(s []string, val string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == val {
			return i
		}
	}
	return -1
}

func loadDataInFile(tableName string, data []byte) {
	if err := ioutil.WriteFile("/tmp/tmpload.txt", data, 0644); err != nil {
		log.Fatal(err)
	}

	query := fmt.Sprintf("LOAD DATA INFILE '/tmp/tmpload.txt' INTO TABLE %s FIELDS TERMINATED BY ',' LINES TERMINATED BY '\n'", tableName)
	result, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(result)
}

func main() {
	r, err := zip.OpenReader("/tmp/data/data.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Fatal(err)
		}

		var ac RawAccountsContainer
		if err := json.Unmarshal(b, &ac); err != nil {
			log.Fatal(err)
		}

		type msi map[string]interface{}
		accounts := make([]Account, 0)

		for _, rawAccount := range ac.Accounts {
			var a Account
			a.ID = rawAccount.ID
			a.Fname = rawAccount.Fname
			a.Sname = rawAccount.Sname
			a.Email = rawAccount.Email
			a.Status = int8(sliceIndex([]string{"свободны", "заняты", "всё сложно"}, rawAccount.Status))
			a.Premium_start = rawAccount.Premium.Start
			a.Premium_end = rawAccount.Premium.Finish
			a.Sex = int8(sliceIndex([]string{"m", "f"}, rawAccount.Sex))
			a.Phone = rawAccount.Phone
			a.Birth = rawAccount.Birth
			a.City = rawAccount.City
			a.Country = rawAccount.Country
			a.Joined = rawAccount.Joined
			accounts = append(accounts, a)
		}

		sb := bytes.Buffer{}
		for _, a := range accounts {
			sb.WriteString(a.oneline())
		}

		loadDataInFile("accounts", sb.Bytes())
	}

}
