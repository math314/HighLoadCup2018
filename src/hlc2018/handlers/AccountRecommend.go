package handlers

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo"
	"hlc2018/common"
	"hlc2018/globals"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

type AccountRecommendParam struct {
	id     int
	wheres bytes.Buffer
	limit  int
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
	agp.addWhere(fmt.Sprintf("a.city = \"%s\"", param))
	return nil
}

func countryRecommendParser(param string, agp *AccountRecommendParam) error {
	if param == "" {
		return fmt.Errorf("country is empty")
	}
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
	arp = &AccountRecommendParam{-1, bytes.Buffer{}, -1}
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

	var accounts []common.Account
	if err := globals.DB.Select(&accounts, "SELECT id, birth, sex from accounts WHERE id = ?", arp.id); err != nil {
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}
	if len(accounts) != 1 {
		return nil, &HlcHttpError{http.StatusNotFound, err}
	}
	account := accounts[0]

	interestsCounts := globals.Is.GetSuggestInterestIds(account.ID)
	if len(interestsCounts) == 0 {
		return nil, nil
	}
	oppositeSex := 3 - account.Sex

	queryTemplate := `
SELECT a.id, a.email, a.status, IFNULL(a.fname, "") AS fname, IFNULL(a.sname, "") AS sname, a.birth, a.premium_start, a.premium_end, a.premium_now, a.status
FROM accounts as a
WHERE
 a.sex = %d
 AND a.id != %d
 %s
 AND a.id in (%s)
`
	where := ""
	if arp.wheres.Len() != 0 {
		where = " AND " + arp.wheres.String()
	}
	query1 := fmt.Sprintf(queryTemplate,
		oppositeSex, account.ID, where, common.IntIntMapJoin(interestsCounts, ","))
	log.Printf("query := %s", query1)

	var acs []*common.Account
	if err := globals.DB.Select(&acs, query1); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
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

	return acs[:retLen], nil
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
