package handlers

import (
	"encoding/json"
	"github.com/labstack/echo"
	"hlc2018/common"
	"hlc2018/globals"
	"io/ioutil"
	"log"
	"net/http"
)

type RawLikesContainer struct {
	Likes []struct {
		Likee int `json:"likee"`
		Ts    int `json:"ts"`
		Liker int `json:"liker"`
	} `json:"likes"`
}

func (rlc *RawLikesContainer) ToLikes() []*common.Like {
	var ret []*common.Like
	for _, l := range rlc.Likes {
		ret = append(ret, &common.Like{l.Liker, l.Likee, l.Ts})
	}
	return ret
}

func AccountsLikesHandlerCore(j []byte) error {
	var rlc RawLikesContainer
	if err := json.Unmarshal([]byte(j), &rlc); err != nil {
		return err
	}

	likes := rlc.ToLikes()
	for _, i := range likes {
		if err := globals.Ls.IsValidCommonLike(i); err != nil {
			return err
		}
	}
	for _, i := range likes {
		if err := globals.Ls.InsertCommonLikeWithoutRangeCheck(i); err != nil {
			return err
		}
	}

	return nil
}

func AccountsLikesHandler(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Fatal(err)
	}
	err = AccountsLikesHandlerCore(body)
	if err != nil {
		log.Print(err)
		return c.String(http.StatusBadRequest, "")
	}
	return c.JSON(http.StatusAccepted, map[string]struct{}{})
}
