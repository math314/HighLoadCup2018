package globals

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"hlc2018/store"
	"log"
	"os"
	"time"
)

var (
	DB *sqlx.DB
	Ls = store.NewLikeStore()
	Is = store.NewInterestStore()
	As = store.NewAccountStore()
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

	log.Printf("Connecting to DB: %q", dsn)
	var err error
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	DB.SetMaxOpenConns(20)
	DB.SetConnMaxLifetime(5 * time.Minute)
	log.Printf("Succeeded to connect DB.")
}
