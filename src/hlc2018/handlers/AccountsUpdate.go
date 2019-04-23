package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"hlc2018/common"
	"hlc2018/globals"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func AccountsUpdateHandlerCore(idStr string, j []byte) *HlcHttpError {
	var ra common.RawAccount
	if err := json.Unmarshal([]byte(j), &ra); err != nil {
		return &HlcHttpError{http.StatusNotFound, err}
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return &HlcHttpError{http.StatusBadRequest, err}
	}
	ra.ID = id

	if len(ra.Likes) != 0 {
		return &HlcHttpError{http.StatusBadRequest, fmt.Errorf("likes cannot be updated")}
	}

	a, err := ra.ToAccount()
	if err != nil {
		return &HlcHttpError{http.StatusBadRequest, err}
	}

	if _, err := globals.As.GetStoredAccount(id); err != nil {
		return &HlcHttpError{http.StatusNotFound, fmt.Errorf("account not found")}
	}

	if err := globals.As.UpdateAccountCommon(a); err != nil {
		return &HlcHttpError{http.StatusBadRequest, err}
	}

	if ra.Interests != nil {
		if err := globals.Is.UpdateInterests(a.ID, ra.Interests); err != nil {
			return &HlcHttpError{http.StatusBadRequest, err}
		}
	}

	return nil
}

func AccountsUpdateHandler(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Fatal(err)
	}
	herr := AccountsUpdateHandlerCore(c.Param("id"), body)
	if herr != nil {
		log.Print(herr)
		return c.String(herr.HttpStatusCode, "")
	}
	return c.JSON(http.StatusAccepted, map[string]struct{}{})
}
