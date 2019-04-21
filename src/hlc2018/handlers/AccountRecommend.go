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

	interests := globals.Is.GetCommonInterests(account.ID)
	if len(interests) == 0 {
		return nil, nil
	}

	joinedInterests := bytes.Buffer{}
	for _, i := range interests {
		if joinedInterests.Len() != 0 {
			joinedInterests.WriteString(", ")
		}
		joinedInterests.WriteString("\"")
		joinedInterests.WriteString(i.Interest)
		joinedInterests.WriteString("\"")
	}

	oppositeSex := 3 - account.Sex

	queryTemplate := `
SELECT a.id, a.email, a.status, IFNULL(a.fname, "") AS fname, IFNULL(a.sname, "") AS sname, a.birth, a.premium_start, a.premium_end
FROM accounts as a, (
 SELECT account_id, COUNT(account_id) AS ` + "`count`" + `
 FROM interests
 WHERE interest in (%s)
 GROUP BY account_id
 ) as b
WHERE
 a.sex = %d
 AND a.id != %d
 AND %s
 AND a.id = b.account_id
ORDER BY
a.premium_now DESC
, a.status_for_recommend DESC
, b.count DESC
, abs(a.birth - %d) ASC
LIMIT %d
`
	wheres1 := arp.wheres
	if wheres1.Len() != 0 {
		wheres1.WriteString(" AND ")
	}
	wheres1.WriteString("b.count != 0")

	query1 := fmt.Sprintf(queryTemplate,
		joinedInterests.String(), oppositeSex, account.ID, wheres1.String(), account.Birth, arp.limit)
	log.Printf("query := %s", query1)

	var acs []*common.Account
	if err := globals.DB.Select(&acs, query1); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}
	return acs, nil
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
