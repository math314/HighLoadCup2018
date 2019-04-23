package handlers

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"hlc2018/common"
	"hlc2018/globals"
	"io/ioutil"
	"net/http"
)

func AccountsInsertHandlerCore(j []byte) error {
	var ra common.RawAccount
	if err := json.Unmarshal([]byte(j), &ra); err != nil {
		return err
	}

	a, err := ra.ToAccount()
	if err != nil {
		return err
	}
	interests := ra.ToInterests()
	likes := ra.ToLikes()

	if err := globals.As.InsertAccountCommon(a); err != nil {
		return err
	}
	for _, i := range interests {
		globals.Is.InsertCommonInterest(i)
	}
	for _, i := range likes {
		if err := globals.Ls.InsertCommonLike(i); err != nil {
			return err
		}
	}

	return nil
}

func AccountsInsertHandler(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Fatal(err)
	}
	err = AccountsInsertHandlerCore(body)
	if err != nil {
		log.Print(err)
		return c.String(http.StatusBadRequest, "")
	}
	return c.JSON(http.StatusCreated, map[string]struct{}{})
}
