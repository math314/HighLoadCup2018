package main

import (
	"archive/zip"
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
)

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
	e.POST("/accounts/new/", handlers.AccountsInsertHandler)
	e.POST("/accounts/:id/", handlers.AccountsUpdateHandler)
	e.POST("/accounts/likes/", handlers.AccountsLikesHandler)

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
			a, err := rawAccount.ToAccount()
			if err != nil {
				log.Fatal(err)
			}
			accounts = append(accounts, a)
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
