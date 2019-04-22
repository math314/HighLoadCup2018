package handlers

import (
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
	"strings"
	"time"
)

type Tribool int8

const (
	TUndefined Tribool = iota
	TFalse
	TTrue
)

type AccountsFilterParams struct {
	selects           map[string]struct{}
	limit             int
	sexEq             int8
	emailDomain       string
	emailLt           string
	emailGt           string
	statusEq          int8
	statusNeq         int8
	fnameEq           string
	fnameAny          []string
	fnameNull         Tribool
	snameEq           string
	snameStarts       string
	snameNull         Tribool
	phoneCode         int
	phoneNull         Tribool
	countryEq         string
	countryNull       Tribool
	cityEq            string
	cityAny           []string
	cityNull          Tribool
	birthLt           int
	birthGt           int
	birthYear         int
	interestsContains []string
	interestsAny      []string
	likeContains      []int
	premiumNow        Tribool
	premiumNull       Tribool
}

func (afp *AccountsFilterParams) addSelect(s string) {
	afp.selects[s] = struct{}{}
}

func SexEqFilter(param string, afp *AccountsFilterParams) error {
	sex := common.SexFromString(param)
	if sex == 0 {
		return fmt.Errorf("%s is not valid sex", param)
	}
	afp.addSelect("sex")
	afp.sexEq = sex
	return nil
}

func emailDomainFilter(param string, afp *AccountsFilterParams) error {
	if strings.Contains(param, "%") {
		return fmt.Errorf("domain (%s) cannot contain \"%%\"", param)
	}
	afp.emailDomain = param
	return nil
}

func emailLtFilter(param string, afp *AccountsFilterParams) error {
	afp.emailLt = param
	return nil
}

func emailGtFilter(param string, afp *AccountsFilterParams) error {
	afp.emailGt = param
	return nil
}

func StatusEqFilter(param string, afp *AccountsFilterParams) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	afp.addSelect("status")
	afp.statusEq = status
	return nil
}

func StatusNeqFilter(param string, afp *AccountsFilterParams) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	afp.addSelect("status")
	afp.statusNeq = status
	return nil
}

func fnameAnyFilter(param string, afp *AccountsFilterParams) error {
	names := strings.Split(param, ",")
	afp.addSelect("fname")
	afp.fnameAny = names
	return nil
}

func snameStartsFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("sname")
	afp.snameStarts = param
	return nil
}

func phoneCodeFilter(param string, afp *AccountsFilterParams) error {
	if len(param) != 3 {
		return fmt.Errorf("phone code param length should be 3 but %s", len(param))
	}
	for _, c := range param {
		if c < '0' || c > '9' {
			return fmt.Errorf("phone code param should be [0-9] : %s", param)
		}
	}
	afp.addSelect("phone")
	var err error
	afp.phoneCode, err = strconv.Atoi(param)
	return err
}

func cityAnyFilter(param string, afp *AccountsFilterParams) error {
	names := strings.Split(param, ",")

	afp.addSelect("city")
	afp.cityAny = names
	return nil
}

func birthLtFilter(param string, afp *AccountsFilterParams) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	afp.addSelect("birth")
	afp.birthLt = birth
	return nil
}

func birthGtFilter(param string, afp *AccountsFilterParams) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	afp.addSelect("birth")
	afp.birthGt = birth
	return nil
}

func birthYearFilter(param string, afp *AccountsFilterParams) error {
	year, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth year (%s)", param)
	}

	afp.addSelect("birth")
	afp.birthYear = year
	return nil
}

func premiumNowFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("premium_start")
	afp.addSelect("premium_end")
	afp.premiumNow = TTrue
	return nil
}

func premiumNullFilter(param string, afp *AccountsFilterParams) error {
	if param == "0" {
		afp.addSelect("premium_start")
		afp.addSelect("premium_end")
		afp.premiumNull = TFalse
	} else if param == "1" {
		afp.addSelect("premium_start")
		afp.addSelect("premium_end")
		afp.premiumNull = TTrue
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

func interestsAnyFilter(param string, afp *AccountsFilterParams) error {
	names := strings.Split(param, ",")

	ids := globals.Is.ContainsAnyFromInterests(names)
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	afp.interestsAny = names
	return nil
}

func interestsContainsFilter(param string, afp *AccountsFilterParams) error {
	names := strings.Split(param, ",")

	ids := globals.Is.ContainsAllFromInterests(names)
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	afp.interestsContains = names
	return nil
}

func likesContainsFilter(param string, afp *AccountsFilterParams) error {
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

	afp.likeContains = liked
	return nil
}

func noopFilter(param string, sb *AccountsFilterParams) error {
	return nil
}

type FilterFunc func(param string, afp *AccountsFilterParams) error

func fnameEqFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("fname")
	afp.fnameEq = param
	return nil
}

func snameEqFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("sname")
	afp.snameEq = param
	return nil
}

func countryEqFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("country")
	afp.countryEq = param
	return nil
}

func cityEqFilter(param string, afp *AccountsFilterParams) error {
	afp.addSelect("city")
	afp.cityEq = param
	return nil
}

func nullFilterParser(param string, afp *AccountsFilterParams, name string) (Tribool, error) {
	afp.addSelect(name)
	if param == "0" {
		return TFalse, nil
	} else if param == "1" {
		return TTrue, nil
	} else {
		return TUndefined, fmt.Errorf("%s param is not valid (%s)", name, param)
	}
}

func fnameNullFilter(param string, afp *AccountsFilterParams) error {
	var err error
	afp.fnameNull, err = nullFilterParser(param, afp, "fname")
	return err
}

func snameNullFilter(param string, afp *AccountsFilterParams) error {
	var err error
	afp.snameNull, err = nullFilterParser(param, afp, "sname")
	return err
}

func phoneNullFilter(param string, afp *AccountsFilterParams) error {
	var err error
	afp.phoneNull, err = nullFilterParser(param, afp, "phone")
	return err
}

func countryNullFilter(param string, afp *AccountsFilterParams) error {
	var err error
	afp.countryNull, err = nullFilterParser(param, afp, "country")
	return err
}

func cityNullFilter(param string, afp *AccountsFilterParams) error {
	var err error
	afp.cityNull, err = nullFilterParser(param, afp, "city")
	return err
}

var filterFuncs = map[string]FilterFunc{
	"sex_eq":             SexEqFilter, // 1/2
	"email_domain":       emailDomainFilter,
	"email_lt":           emailLtFilter,
	"email_gt":           emailGtFilter,
	"status_eq":          StatusEqFilter,  // 1/3
	"status_neq":         StatusNeqFilter, // 2/3
	"fname_eq":           fnameEqFilter,   // 1/100 ~ 1/150
	"fname_any":          fnameAnyFilter,
	"fname_null":         fnameNullFilter, // 1/15
	"sname_eq":           snameEqFilter,   // 1 / 1000
	"sname_starts":       snameStartsFilter,
	"sname_null":         snameNullFilter,   // 1/4
	"phone_code":         phoneCodeFilter,   // 1/200 ~ 1/300 NOTE: only (900) ~ (999) are available
	"phone_null":         phoneNullFilter,   // 1/2
	"country_eq":         countryEqFilter,   // 1/40 ~ 1/100
	"country_null":       countryNullFilter, // 1/6
	"city_eq":            cityEqFilter,      // 1/300
	"city_any":           cityAnyFilter,     // ?
	"city_null":          cityNullFilter,    //  1/3
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

func accountsFilterParser(queryParams url.Values) (afp *AccountsFilterParams, err error) {
	afp = &AccountsFilterParams{
		selects: map[string]struct{}{},
		limit:   -1,
	}
	afp.addSelect("id")
	afp.addSelect("email")

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
		if err = fun(param[0], afp); err != nil {
			return
		}
	}
	if afp.limit == -1 {
		err = fmt.Errorf("limit is not specified")
		return
	}

	return
}

type StoreFilterFunc func(id int) bool

func ApplyFilter(ss store.StoreSource, filter StoreFilterFunc, limit int) []int {
	var ret []int
	for ss.Next() {
		val := ss.Value()
		if filter(val) {
			ret = append(ret, val)
			if len(ret) == limit {
				return ret
			}
		}
	}
	return ret
}

func IsNullStoreFilterString(val string, b Tribool) bool {
	if b == TTrue {
		return val == ""
	} else if b == TFalse {
		return val != ""
	} else {
		log.Fatal("TUndefined is provided to IsNullStoreFilterString")
		return false
	}
}

func IsNullStoreFilterInt(val int, b Tribool) bool {
	if b == TTrue {
		return val == 0
	} else if b == TFalse {
		return val != 0
	} else {
		log.Fatal("TUndefined is provided to IsNullStoreFilterInt")
		return false
	}
}

func GenFilterFromAccountsFilterParams(afp *AccountsFilterParams) StoreFilterFunc {
	return func(id int) bool {
		if len(afp.likeContains) > 0 {
			result := globals.Ls.CheckContainAllLikes(id, afp.likeContains)
			if !result {
				return false
			}
		}
		if len(afp.interestsContains) > 0 {
			result := globals.Is.ContainsAll(id, afp.interestsContains)
			if !result {
				return false
			}
		}

		// 1/30 if length == 1
		if len(afp.interestsAny) > 0 {
			result := globals.Is.ContainsAny(id, afp.interestsAny)
			if !result {
				return false
			}
		}

		me := globals.As.GetStoredAccountWithoutError(id)

		//  "sex_eq":             SexEqFilter, // 1/2
		if afp.sexEq != 0 {
			if me.Sex != afp.sexEq {
				return false
			}
		}

		if afp.emailDomain != "" {
			domain := strings.Split(me.Email, "@")[1]
			if domain != afp.emailDomain {
				return false
			}
		}

		if afp.emailLt != "" {
			if me.Email > afp.emailLt {
				return false
			}
		}

		if afp.emailGt != "" {
			if me.Email < afp.emailGt {
				return false
			}
		}

		if afp.statusEq != 0 {
			if me.Status != afp.statusEq {
				return false
			}
		}

		if afp.statusNeq != 0 {
			if me.Status == afp.statusNeq {
				return false
			}
		}

		if afp.fnameEq != "" {
			if me.Fname != afp.fnameEq {
				return false
			}
		}

		if len(afp.fnameAny) > 0 {
			ok := false
			for _, f := range afp.fnameAny {
				if me.Fname == f {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}

		if afp.fnameNull != TUndefined {
			ok := IsNullStoreFilterString(me.Fname, afp.fnameNull)
			if !ok {
				return false
			}
		}

		if afp.snameEq != "" {
			if me.Sname != afp.snameEq {
				return false
			}
		}

		if afp.snameStarts != "" {
			if len(me.Sname) < len(afp.snameStarts) {
				return false
			}
			if me.Sname[:len(afp.snameStarts)] != afp.snameStarts {
				return false
			}
		}

		if afp.snameNull != TUndefined {
			ok := IsNullStoreFilterString(me.Sname, afp.snameNull)
			if !ok {
				return false
			}
		}

		if afp.phoneCode != 0 {
			if !me.Phone.HasPhoneCode(afp.phoneCode) {
				return false
			}
		}

		if afp.phoneNull != TUndefined {
			if afp.phoneNull == TTrue {
				if me.Phone.Int != 0 {
					return false
				}
			} else {
				if me.Phone.Int == 0 {
					return false
				}
			}
		}

		if afp.countryEq != "" {
			if me.Country != globals.As.GetCountryId(afp.countryEq) {
				return false
			}
		}

		if afp.countryNull != TUndefined {
			ok := IsNullStoreFilterInt(me.Country, afp.countryNull)
			if !ok {
				return false
			}
		}

		if afp.cityEq != "" {
			if me.City != globals.As.GetCityId(afp.cityEq) {
				return false
			}
		}

		if len(afp.cityAny) > 0 {
			ok := false
			for _, f := range afp.cityAny {
				if me.City == globals.As.GetCityId(f) {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}

		if afp.cityNull != TUndefined {
			ok := IsNullStoreFilterInt(me.City, afp.cityNull)
			if !ok {
				return false
			}
		}

		if afp.birthLt != 0 {
			if me.Birth > afp.birthLt {
				return false
			}
		}

		if afp.birthGt != 0 {
			if me.Birth < afp.birthGt {
				return false
			}
		}

		if afp.birthYear != 0 {
			from := int(time.Date(afp.birthYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
			to := int(time.Date(afp.birthYear+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
			ok := from <= me.Birth && me.Birth < to
			if !ok {
				return false
			}
		}

		if afp.premiumNow != TUndefined {
			if !me.Premium_now {
				return false
			}
		}

		if afp.premiumNull != TUndefined {
			if afp.premiumNull == TTrue {
				if me.Premium_start != 0 {
					return false
				}
			} else {
				if me.Premium_start == 0 {
					return false
				}
			}
		}

		return true
	}
}

func SplitParamsIntoStoreAndFilter(originalAfp *AccountsFilterParams) (*AccountsFilterParams, store.StoreSource) {
	afp := *originalAfp
	//?
	if len(afp.likeContains) > 0 {
		liker := globals.Ls.IdsContainAllLikes(afp.likeContains)
		sort.Sort(sort.Reverse(sort.IntSlice(liker)))

		afp.likeContains = nil
		return &afp, store.NewArrayStoreSource(liker)
	}

	// 1/30 if length == 1
	if len(afp.interestsContains) > 0 {
		mp := globals.Is.ContainsAllFromInterests(afp.interestsContains)

		var ids []int
		for id, _ := range mp {
			ids = append(ids, id)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(ids)))

		afp.interestsContains = nil
		return &afp, store.NewArrayStoreSource(ids)
	}

	// 1/30 if length == 1
	if len(afp.interestsAny) > 0 {
		mp := globals.Is.ContainsAnyFromInterests(afp.interestsAny)
		var ids []int
		for id, _ := range mp {
			ids = append(ids, id)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(ids)))

		afp.interestsAny = nil
		return &afp, store.NewArrayStoreSource(ids)
	}

	//
	//  "sex_eq":             SexEqFilter, // 1/2
	//	"email_domain":       emailDomainFilter,
	//	"email_lt":           emailLtFilter,
	//	"email_gt":           emailGtFilter,
	//	"status_eq":          StatusEqFilter,  // 1/3
	//	"status_neq":         StatusNeqFilter, // 2/3
	//	"fname_eq":           fnameEqFilter,   // 1/100 ~ 1/150
	//	"fname_any":          fnameAnyFilter,
	//	"fname_null":         fnameNullFilter, // 1/15
	//	"sname_eq":           snameEqFilter,   // 1 / 1000
	//	"sname_starts":       snameStartsFilter,
	//	"sname_null":         snameNullFilter,   // 1/4
	//	"phone_code":         phoneCodeFilter,   // 1/200 ~ 1/300 NOTE: only (900) ~ (999) are available
	//	"phone_null":         phoneNullFilter,   // 1/2
	//	"country_eq":         countryEqFilter,   // 1/40 ~ 1/100
	//	"country_null":       countryNullFilter, // 1/6
	//	"city_eq":            cityEqFilter,      // 1/300
	//	"city_any":           cityAnyFilter,     // ?
	//	"city_null":          cityNullFilter,    //  1/3
	//	"birth_lt":           birthLtFilter,
	//	"birth_gt":           birthGtFilter,
	//	"birth_year":         birthYearFilter,
	//	"premium_now":        premiumNowFilter,  // 1/10
	//	"premium_null":       premiumNullFilter, // 2/3

	// default because there're no index
	return &afp, globals.As.NewRangeAccountStoreSource()
}

func filterIds(originalAfp *AccountsFilterParams) []int {
	afp, ss := SplitParamsIntoStoreAndFilter(originalAfp)
	sff := GenFilterFromAccountsFilterParams(afp)

	ret := ApplyFilter(ss, sff, afp.limit)

	return ret
}

func AccountsFilterCore(queryParams url.Values) (*common.AccountContainer, *HlcHttpError) {
	afp, err := accountsFilterParser(queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}
	ansIds := filterIds(afp)

	afas := common.AccountContainer{}
	for _, id := range ansIds {
		a := globals.As.GetStoredAccountWithoutError(id)
		r := common.Account{
			ID:    a.ID,
			Email: a.Email,
		}
		if _, found := afp.selects["sex"]; found {
			r.Sex = a.Sex
		}
		if _, found := afp.selects["status"]; found {
			r.Status = a.Status
		}
		if _, found := afp.selects["fname"]; found {
			r.Fname = a.Fname
		}
		if _, found := afp.selects["sname"]; found {
			r.Sname = a.Sname
		}
		if _, found := afp.selects["phone"]; found {
			r.Phone = a.Phone.String()
		}
		if _, found := afp.selects["city"]; found {
			r.City = globals.As.IdToCity(a.City)
		}
		if _, found := afp.selects["country"]; found {
			r.Country = globals.As.IdToCountry(a.Country)
		}
		if _, found := afp.selects["birth"]; found {
			r.Birth = a.Birth
		}
		if _, found := afp.selects["premium_start"]; found {
			r.Premium_start = a.Premium_start
		}
		if _, found := afp.selects["premium_end"]; found {
			r.Premium_end = a.Premium_end
		}
		afas.Accounts = append(afas.Accounts, &r)
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
