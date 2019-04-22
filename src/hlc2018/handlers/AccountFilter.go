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
	"strings"
	"time"
)

type AccountsFilterParams struct {
	selects map[string]struct{}
	wheres  bytes.Buffer
	limit   int
}

func (sb *AccountsFilterParams) addSelect(s string) {
	sb.selects[s] = struct{}{}
}

func (sb *AccountsFilterParams) addWhere(s string) {
	if sb.wheres.Len() != 0 {
		sb.wheres.WriteString(" AND ")
	}
	sb.wheres.WriteString(s)
}

func SexEqFilter(param string, sb *AccountsFilterParams) error {
	sex := common.SexFromString(param)
	if sex == 0 {
		return fmt.Errorf("%s is not valid sex", param)
	}
	sb.addSelect("sex")
	sb.addWhere(fmt.Sprintf("sex = %d", sex))
	return nil
}

func emailDomainFilter(param string, sb *AccountsFilterParams) error {
	if strings.Contains(param, "%") {
		return fmt.Errorf("domain (%s) cannot contain \"%%\"", param)
	}
	sb.addSelect("email")
	sb.addWhere(fmt.Sprintf("email like \"%s\"", "%"+param))
	return nil
}

func emailLtFilter(param string, sb *AccountsFilterParams) error {
	sb.addSelect("email")
	sb.addWhere(fmt.Sprintf("email <= \"%s\"", param))
	return nil
}

func emailGtFilter(param string, sb *AccountsFilterParams) error {
	sb.addSelect("email")
	sb.addWhere(fmt.Sprintf("email >= \"%s\"", param))
	return nil
}

func StatusEqFilter(param string, sb *AccountsFilterParams) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	sb.addSelect("status")
	sb.addWhere(fmt.Sprintf("status = %d", status))
	return nil
}

func StatusNeqFilter(param string, sb *AccountsFilterParams) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	sb.addSelect("status")
	sb.addWhere(fmt.Sprintf("status != %d", status))
	return nil
}

func fnameAnyFilter(param string, sb *AccountsFilterParams) error {
	sb.addSelect("fname")
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")
	sb.addWhere(fmt.Sprintf("fname in (%s)", arg))
	return nil
}

func snameStartsFilter(param string, sb *AccountsFilterParams) error {
	sb.addSelect("sname")
	sb.addWhere(fmt.Sprintf("sname LIKE \"%s\"", param+"%"))
	return nil
}

func phoneCodeFilter(param string, sb *AccountsFilterParams) error {
	if len(param) != 3 {
		return fmt.Errorf("phone code param length should be 3 but %s", len(param))
	}
	for _, c := range param {
		if c < '0' || c > '9' {
			return fmt.Errorf("phone code param should be [0-9] : %s", param)
		}
	}
	sb.addSelect("phone")
	sb.addWhere(fmt.Sprintf("phone LIKE \"%s\"", "%("+param+")%"))
	return nil
}

func cityAnyFilter(param string, sb *AccountsFilterParams) error {
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")

	sb.addSelect("city")
	sb.addWhere(fmt.Sprintf("city in (%s)", arg))
	return nil
}

func birthLtFilter(param string, sb *AccountsFilterParams) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	sb.addSelect("birth")
	sb.addWhere(fmt.Sprintf("birth < %d", birth))
	return nil
}

func birthGtFilter(param string, sb *AccountsFilterParams) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	sb.addSelect("birth")
	sb.addWhere(fmt.Sprintf("birth > %d", birth))
	return nil
}

func birthYearFilter(param string, sb *AccountsFilterParams) error {
	year, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth year (%s)", param)
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	sb.addSelect("birth")
	sb.addWhere(fmt.Sprintf("birth >= %d", from.Unix()))
	sb.addWhere(fmt.Sprintf("birth < %d", after.Unix()))
	return nil
}

func premiumNowFilter(param string, sb *AccountsFilterParams) error {
	sb.addSelect("premium_start")
	sb.addSelect("premium_end")
	sb.addWhere("premium_now = 1")
	return nil
}

func premiumNullFilter(param string, sb *AccountsFilterParams) error {
	if param == "0" {
		sb.addSelect("premium_start")
		sb.addSelect("premium_end")
		sb.addWhere("premium_start != 0")
	} else if param == "1" {
		sb.addSelect("premium_start")
		sb.addSelect("premium_end")
		sb.addWhere("premium_start = 0")
	} else {
		return fmt.Errorf("premium param is not valid (%s)", param)
	}
	return nil
}

func limitFilter(param string, sb *AccountsFilterParams) error {
	limit, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse limit (%s)", param)
	}
	if limit <= 0 {
		return fmt.Errorf("limit should be positive (%s)", param)
	}
	sb.limit = limit
	return nil
}

func interestsAnyFilter(param string, sb *AccountsFilterParams) error {
	names := strings.Split(param, ",")

	ids := globals.Is.ContainsAnyFromInterests(names)
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	sb.addWhere(fmt.Sprintf("id in (%s)", common.IntSetJoin(ids, ",")))
	return nil
}

func interestsContainsFilter(param string, sb *AccountsFilterParams) error {
	names := strings.Split(param, ",")

	ids := globals.Is.ContainsAllFromInterests(names)
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	sb.addWhere(fmt.Sprintf("id in (%s)", common.IntSetJoin(ids, ",")))
	return nil
}

func likesContainsFilter(param string, sb *AccountsFilterParams) error {
	var liked []int
	for _, s := range strings.Split(param, ",") {
		id, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("failed to parse likes (%s)", param)
		}
		liked = append(liked, id)
	}

	liker := globals.Ls.IdsContainAllLikes(liked)
	if len(liker) == 0 {
		liker = []int{-1}
	}

	sb.addWhere(fmt.Sprintf("id in (%s)", common.IntArrayJoin(liker, ",")))
	return nil
}

func noopFilter(param string, sb *AccountsFilterParams) error {
	return nil
}

type FilterFunc func(param string, sb *AccountsFilterParams) error

func eqFilterGenerator(name string) FilterFunc {
	return func(param string, sb *AccountsFilterParams) error {
		sb.addSelect(name)
		sb.addWhere(fmt.Sprintf("%s = \"%s\"", name, param))
		return nil
	}
}

func nullFilterGenerator(name string) FilterFunc {
	return func(param string, sb *AccountsFilterParams) error {
		if param == "0" {
			sb.addSelect(name)
			sb.addWhere(fmt.Sprintf("%s IS NOT NULL", name))
		} else if param == "1" {
			sb.addWhere(fmt.Sprintf("%s IS NULL", name))
		} else {
			return fmt.Errorf("%s param is not valid (%s)", name, param)
		}
		return nil
	}
}

var filterFuncs = map[string]FilterFunc{
	"sex_eq":             SexEqFilter, // 1/2
	"email_domain":       emailDomainFilter,
	"email_lt":           emailLtFilter,
	"email_gt":           emailGtFilter,
	"status_eq":          StatusEqFilter,             // 1/3
	"status_neq":         StatusNeqFilter,            // 2/3
	"fname_eq":           eqFilterGenerator("fname"), // 1/100 ~ 1/150
	"fname_any":          fnameAnyFilter,
	"fname_null":         nullFilterGenerator("fname"), // 1/15
	"sname_eq":           eqFilterGenerator("sname"),   // 1 / 1000
	"sname_starts":       snameStartsFilter,
	"sname_null":         nullFilterGenerator("sname"),   // 1/4
	"phone_code":         phoneCodeFilter,                // 1/200 ~ 1/300 NOTE: only (900) ~ (999) are available
	"phone_null":         nullFilterGenerator("phone"),   // 1/2
	"country_eq":         eqFilterGenerator("country"),   // 1/40 ~ 1/100
	"country_null":       nullFilterGenerator("country"), // 1/6
	"city_eq":            eqFilterGenerator("city"),      // 1/300
	"city_any":           cityAnyFilter,                  // ?
	"city_null":          nullFilterGenerator("city"),    //  1/3
	"birth_lt":           birthLtFilter,
	"birth_gt":           birthGtFilter,
	"birth_year":         birthYearFilter,
	"interests_contains": interestsContainsFilter, // 1/30 if length == 1
	"interests_any":      interestsAnyFilter,      // 1/30 if length == 1
	"likes_contains":     likesContainsFilter,
	"premium_now":        premiumNowFilter,  // 1/10
	"premium_null":       premiumNullFilter, // 2/3
	"limit":              limitFilter,
	"query_id":           noopFilter,
}

func accountsFilterParser(queryParams url.Values) (sb AccountsFilterParams, err error) {
	sb = AccountsFilterParams{map[string]struct{}{}, bytes.Buffer{}, -1}
	sb.addSelect("id")
	sb.addSelect("email")

	for field, param := range queryParams {
		if param[0] == "" {
			err = fmt.Errorf("parameter cannot be empty (field = %s)", field)
			return
		}

		fun, found := filterFuncs[field]
		if !found {
			err = fmt.Errorf("filter (%s) not found", field)
			return
		}
		if len(param) != 1 {
			err = fmt.Errorf("multiple params in filter (%s)", field)
			return
		}
		if err = fun(param[0], &sb); err != nil {
			return
		}
	}
	if sb.limit == -1 {
		err = fmt.Errorf("limit is not specified")
		return
	}

	return
}

func AccountsFilterCore(queryParams url.Values) (*common.AccountContainer, *HlcHttpError) {
	sb, err := accountsFilterParser(queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}

	selectCluster := bytes.Buffer{}
	for k, _ := range sb.selects {
		if selectCluster.Len() != 0 {
			selectCluster.WriteString(", ")
		}
		selectCluster.WriteString(k)
	}

	whereCluster := ""
	if sb.wheres.Len() != 0 {
		whereCluster = "WHERE " + sb.wheres.String()
	}

	query := fmt.Sprintf("SELECT %s FROM accounts %s ORDER BY id DESC LIMIT %d", selectCluster.String(), whereCluster, sb.limit)
	log.Printf("query := %s", query)

	var afas common.AccountContainer
	if err := globals.DB.Select(&afas.Accounts, query); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}

	return &afas, nil
}

func AccountsFilterHandler(c echo.Context) error {
	afas, err := AccountsFilterCore(c.QueryParams())
	if err != nil {
		return c.String(err.HttpStatusCode, "")
	}

	return common.JsonResponseWithoutChunking(c, http.StatusOK, afas.ToRawAccountsContainer())
}
