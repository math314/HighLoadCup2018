package handlers

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo"
	"hlc2018/common"
	"hlc2018/globals"
	"hlc2018/store"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

type AccountRecommendParam struct {
	id     int
	wheres bytes.Buffer // todo: delete
	limit  int

	// adding for recommend
	city    string
	country string
}

func (arp *AccountRecommendParam) addWhere(s string) {
	if arp.wheres.Len() != 0 {
		arp.wheres.WriteString(" AND ")
	}
	arp.wheres.WriteString(s)
}

func idRecommendParser(param string, agp *AccountRecommendParam) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse id (%s)", param)
	}
	agp.id = id
	return nil
}

func limitRecommendParser(param string, agp *AccountRecommendParam) error {
	limit, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse limit (%s)", param)
	}
	if limit <= 0 {
		return fmt.Errorf("limit should be positive (%s)", param)
	}
	agp.limit = limit
	return nil
}

func cityRecommendParser(param string, agp *AccountRecommendParam) error {
	if param == "" {
		return fmt.Errorf("city is empty")
	}
	agp.city = param
	agp.addWhere(fmt.Sprintf("a.city = \"%s\"", param))
	return nil
}

func countryRecommendParser(param string, agp *AccountRecommendParam) error {
	if param == "" {
		return fmt.Errorf("country is empty")
	}
	agp.country = param
	agp.addWhere(fmt.Sprintf("a.country = \"%s\"", param))
	return nil
}

func noopRecommendParser(param string, agp *AccountRecommendParam) error {
	return nil
}

type AccountRecommendFunc func(param string, agp *AccountRecommendParam) error

var accountRecommendFuncs = map[string]AccountRecommendFunc{
	"limit":    limitRecommendParser,
	"city":     cityRecommendParser,
	"country":  countryRecommendParser,
	"query_id": noopRecommendParser,
}

func accountsRecommendParser(idStr string, queryParams url.Values) (arp *AccountRecommendParam, err error) {
	arp = &AccountRecommendParam{-1, bytes.Buffer{}, -1, "", ""}
	if err = idRecommendParser(idStr, arp); err != nil {
		return
	}

	for field, param := range queryParams {
		if param[0] == "" {
			err = fmt.Errorf("parameter cannot be empty (field = %s)", field)
			return
		}
		fun, found := accountRecommendFuncs[field]
		if !found {
			err = fmt.Errorf("filter (%s) not found", field)
			return
		}
		if len(param) != 1 {
			err = fmt.Errorf("multiple params in filter (%s)", field)
			return
		}
		if err = fun(param[0], arp); err != nil {
			return
		}
	}
	if arp.limit == -1 {
		err = fmt.Errorf("limit is not specified")
		return
	}
	if arp.id == -1 {
		err = fmt.Errorf("id is not specified")
		return
	}

	return
}

func AccountsRecommendCore(idStr string, queryParams url.Values) ([]*common.Account, *HlcHttpError) {
	arp, err := accountsRecommendParser(idStr, queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}

	account, err := globals.As.GetStoredAccount(arp.id)
	if err != nil {
		return nil, &HlcHttpError{http.StatusNotFound, err}
	}

	arpCountryId := globals.As.GetCountryId(arp.country)
	arpCityId := globals.As.GetCityId(arp.city)

	interestsCounts := globals.Is.GetSuggestInterestIds(account.ID)
	filteredInterestingsCounts := map[int]int{}
	for k, v := range interestsCounts {
		a := globals.As.GetStoredAccountWithoutError(k)
		if k == account.ID {
			continue
		}
		if a.Sex+account.Sex != 3 {
			continue
		}
		if arpCountryId != 0 {
			if arpCountryId != a.Country {
				continue
			}
		}
		if arpCityId != 0 {
			if arpCityId != a.City {
				continue
			}
		}
		filteredInterestingsCounts[k] = v
	}

	if len(filteredInterestingsCounts) == 0 {
		return nil, nil
	}

	type StoredAccountCount struct {
		*store.StoredAccount
		count int
	}

	var acs []StoredAccountCount
	for id, cnt := range filteredInterestingsCounts {
		acs = append(acs, StoredAccountCount{globals.As.GetStoredAccountWithoutError(id), cnt})
	}

	sort.Slice(acs, func(i, j int) bool {
		if acs[i].Premium_now != acs[j].Premium_now {
			return acs[i].Premium_now
		}
		c, d := acs[i].Status, acs[j].Status
		if c != d {
			if c == 3 {
				return true
			} else if d == 3 {
				return false
			} else {
				//2,1 or 1,2
				return c == 1
			}
		}
		x := interestsCounts[acs[i].ID]
		y := interestsCounts[acs[j].ID]
		if x != y {
			return x > y
		}
		a := common.AbsInt(account.Birth - acs[i].Birth)
		b := common.AbsInt(account.Birth - acs[j].Birth)
		return a < b
	})

	retLen := arp.limit
	if retLen > len(acs) {
		retLen = len(acs)
	}

	var converted []*common.Account
	for i := 0; i < retLen; i++ {
		converted = append(converted, &common.Account{
			ID:            acs[i].ID,
			Email:         acs[i].Email,
			Status:        acs[i].Status,
			Fname:         acs[i].Fname,
			Sname:         acs[i].Sname,
			Birth:         acs[i].Birth,
			Premium_start: acs[i].Premium_start,
			Premium_end:   acs[i].Premium_end,
		})
	}

	return converted, nil
}

func AccountsRecommendHandler(c echo.Context) error {
	acs, err := AccountsRecommendCore(c.Param("id"), c.QueryParams())
	if err != nil {
		return c.String(err.HttpStatusCode, "")
	}

	rac := common.RawAccountsContainer{[]*common.RawAccount{}}
	for _, ac := range acs {
		rac.Accounts = append(rac.Accounts, ac.ToRawAccount())
	}

	return common.JsonResponseWithoutChunking(c, http.StatusOK, &rac)
}
