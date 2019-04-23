package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"hlc2018/common"
	"hlc2018/globals"
	"hlc2018/handlers"
	"hlc2018/tester"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func InsertAccount(tx *sql.Tx, account *common.Account) error {
	args := account.InsertArgs()
	var placeHolders []string
	for i := 0; i < len(args); i++ {
		placeHolders = append(placeHolders, "?")
	}
	if _, err := tx.Exec(fmt.Sprintf("INSERT INTO accounts VALUES(%s)", strings.Join(placeHolders, ",")), args); err != nil {
		return err
	}

	return nil
}

func InsertInterests(tx *sql.Tx, interests []*common.Interest) error {
	for _, v := range interests {
		if _, err := tx.Exec("INSERT INTO interests(account_id, interest) VALUES(?,?)", v.AccountId, v.Interest); err != nil {
			return err
		}
	}
	return nil
}

func DeleteInterests(tx *sql.Tx, accountId int) error {
	_, err := tx.Exec("DELETE FROM interests WHERE account_id = ?", accountId)
	return err
}

func InsertLikes(tx *sql.Tx, likes []*common.Like) error {
	for _, v := range likes {
		globals.Ls.InsertCommonLike(v)
	}
	return nil
}

func accountsNewHandlerCore(rc *common.RawAccount) *handlers.HlcHttpError {
	return nil
}

func accountsNewHandler(c echo.Context) error {
	ra := common.RawAccount{}
	if err := c.Bind(&ra); err != nil {
		log.Fatal(err)
	}
	if err := accountsNewHandlerCore(&ra); err != nil {
		return c.HTML(err.HttpStatusCode, "")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{})
}

func accountsUpdateHandlerCore(rc *common.RawAccount) *handlers.HlcHttpError {
	return nil
}

func accountsIdHandler(c echo.Context) error {
	ra := common.RawAccount{}
	if err := c.Bind(&ra); err != nil {
		log.Fatal(err)
	}
	if err := accountsUpdateHandlerCore(&ra); err != nil {
		return c.HTML(err.HttpStatusCode, "")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{})
}

func accountsLikesHandler(c echo.Context) error {
	return nil
}

func httpMain() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if port == "8080" {
		tester.RunTest()
	}

	e := echo.New()
	if port == "8080" {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "request:\"${method} ${uri}\" status:${status} latency:${latency} (${latency_human}) bytes:${bytes_out}\n",
		}))
	}

	e.GET("/accounts/filter/", handlers.AccountsFilterHandler)
	e.GET("/accounts/group/", handlers.AccountsGroupHandler)
	e.GET("/accounts/:id/recommend/", handlers.AccountsRecommendHandler)
	e.GET("/accounts/:id/suggest/", handlers.AccountsSuggestHandler)
	//e.POST("/accounts/new/", accountsNewHandler)
	//e.POST("/accounts/:id/", accountsIdHandler)
	//e.POST("/accounts/likes/", accountsLikesHandler)

	e.Any("/*", func(context echo.Context) error {
		return context.String(http.StatusNotFound, "")
	})

	e.Start(":" + port)
}

func loadZip() {
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

		var ac common.RawAccountsContainer
		if err := json.Unmarshal(b, &ac); err != nil {
			log.Fatal(err)
		}

		var accounts []*common.Account
		var interests []*common.Interest
		var likes []*common.Like
		for _, rawAccount := range ac.Accounts {
			accounts = append(accounts, rawAccount.ToAccount())
			for _, i := range rawAccount.ToInterests() {
				interests = append(interests, i)
			}
			for _, l := range rawAccount.ToLikes() {
				likes = append(likes, l)
			}
		}

		for _, i := range accounts {
			globals.As.InsertAccountCommon(i)
		}

		for _, i := range interests {
			globals.Is.InsertCommonInterest(i)
		}

		for _, l := range likes {
			globals.Ls.InsertCommonLike(l)
		}
	}
}

func main() {
	loadZip()
	httpMain()
}
