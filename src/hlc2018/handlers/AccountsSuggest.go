package handlers

import (
	"fmt"
	"github.com/labstack/echo"
	"hlc2018/common"
	"hlc2018/globals"
	"log"
	"net/http"
	"net/url"
)

func AccountsSuggestCore(idStr string, queryParams url.Values) ([]*common.Account, *HlcHttpError) {
	arp, err := accountsRecommendParser(idStr, queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}

	var accounts []common.Account
	if err := globals.DB.Select(&accounts, "SELECT id, sex from accounts WHERE id = ?", arp.id); err != nil {
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}
	if len(accounts) != 1 {
		return nil, &HlcHttpError{http.StatusNotFound, err}
	}
	account := accounts[0]

	orderedLiker := globals.Ls.OrderByLikeSimilarity(account.ID)
	if len(orderedLiker) == 0 {
		return nil, nil
	}

	wheres := ""
	if arp.wheres.Len() != 0 {
		wheres = " AND " + arp.wheres.String()
	}

	queryTemplate := `SELECT id FROM accounts as a WHERE sex = %d AND a.id in (%s) %s`
	query := fmt.Sprintf(queryTemplate, account.Sex, common.IntArrayJoin(orderedLiker, ","), wheres)
	log.Print(query)

	var filteredIds []int
	if err := globals.DB.Select(&filteredIds, query); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}
	if len(filteredIds) == 0 {
		return nil, nil
	}

	filteredIdsMap := map[int]struct{}{}
	for _, a := range filteredIds {
		filteredIdsMap[a] = struct{}{}
	}
	orderedRetIds := []int{}
	retIds := map[int]struct{}{}
	for _, id := range orderedLiker {
		if _, ok := filteredIdsMap[id]; !ok {
			continue
		}

		globals.Ls.GetNotLiked(account.ID, id, &retIds, &orderedRetIds, arp.limit)
		if len(retIds) == arp.limit {
			break
		}
	}

	query2 := fmt.Sprintf(`
SELECT a.id, a.email, a.status, IFNULL(a.fname, "") AS fname, IFNULL(a.sname, "") AS sname
FROM accounts AS a
WHERE a.id in (%s)
`, common.IntArrayJoin(orderedRetIds, ","))

	var retAcocuntInfo []*common.Account
	if err := globals.DB.Select(&retAcocuntInfo, query2); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}

	accountMap := map[int]*common.Account{}
	for _, a := range retAcocuntInfo {
		accountMap[a.ID] = a
	}

	var ret []*common.Account
	for _, id := range orderedRetIds {
		ret = append(ret, accountMap[id])
	}

	return ret, nil
}

func AccountsSuggestHandler(c echo.Context) error {
	acs, err := AccountsSuggestCore(c.Param("id"), c.QueryParams())
	if err != nil {
		return c.String(err.HttpStatusCode, "")
	}

	rac := common.RawAccountsContainer{[]*common.RawAccount{}}
	for _, ac := range acs {
		rac.Accounts = append(rac.Accounts, ac.ToRawAccount())
	}

	return common.JsonResponseWithoutChunking(c, http.StatusOK, &rac)
}
