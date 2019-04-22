package handlers

import (
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

	account, err := globals.As.GetStoredAccount(arp.id)
	if err != nil {
		return nil, &HlcHttpError{http.StatusNotFound, err}
	}

	orderedLiker := globals.Ls.OrderByLikeSimilarity(account.ID)
	if len(orderedLiker) == 0 {
		return nil, nil
	}

	var filteredOrderedLiker []int
	for _, id := range orderedLiker {
		a := globals.As.GetStoredAccountWithoutError(id)
		if a.Sex != account.Sex {
			continue
		}
		if arp.country != "" {
			if arp.country != a.Country {
				continue
			}
		}
		if arp.city != "" {
			if arp.city != a.City {
				continue
			}
		}

		filteredOrderedLiker = append(filteredOrderedLiker, id)
	}

	if len(filteredOrderedLiker) == 0 {
		return nil, nil
	}

	orderedRetIds := []int{}
	retIds := map[int]struct{}{}
	for _, id := range filteredOrderedLiker {
		globals.Ls.GetNotLiked(account.ID, id, &retIds, &orderedRetIds, arp.limit)
		if len(retIds) == arp.limit {
			break
		}
	}

	var ret []*common.Account
	for _, id := range orderedRetIds {
		a := globals.As.GetStoredAccountWithoutError(id)
		ret = append(ret, &common.Account{
			ID:     a.ID,
			Email:  a.Email,
			Status: a.Status,
			Fname:  a.Fname,
			Sname:  a.Sname,
		})
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
