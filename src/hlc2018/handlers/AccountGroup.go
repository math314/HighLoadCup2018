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

type AccountGroupParam struct {
	keys            map[string]struct{}
	limit           int
	order           int
	sexEq           int8
	likeContain     int
	countryEq       string
	cityEq          string
	joinedYear      common.JoinedYear
	statusEq        int8
	interestContain string
	birthYear       int
}

type AccountGroupFunc func(param string, agp *AccountGroupParam) error

func sexGroupParser(param string, agp *AccountGroupParam) error {
	sex := common.SexFromString(param)
	if sex == 0 {
		return fmt.Errorf("%s is not valid sex", param)
	}
	agp.sexEq = sex
	return nil
}

func likesGroupParser(param string, agp *AccountGroupParam) error {
	like, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse like (%s)", param)
	}
	liker := globals.Ls.IdsContainAnyLikes([]int{like})
	if len(liker) == 0 {
		liker = []int{-1}
	}

	agp.likeContain = like
	return nil
}

func countryGroupParser(param string, agp *AccountGroupParam) error {
	agp.countryEq = param
	return nil
}

func keysGroupParser(param string, agp *AccountGroupParam) error {
	validKeys := []string{"sex", "status", "interests", "country", "city"}
	for _, k := range strings.Split(param, ",") {
		if common.SliceIndex(validKeys, k) == -1 {
			return fmt.Errorf("invalid keys (%s)", k)
		}
		agp.keys[k] = struct{}{}
	}
	return nil
}

func joinedGroupParser(param string, agp *AccountGroupParam) error {
	joined, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse joined (%s)", param)
	}
	jy := common.ToJoinedYear(joined)
	agp.joinedYear = jy

	return nil
}

func statusGroupParser(param string, agp *AccountGroupParam) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	agp.statusEq = status
	return nil
}

func orderGroupParser(param string, agp *AccountGroupParam) error {
	order, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse order (%s)", param)
	}
	if order != 1 && order != -1 {
		return fmt.Errorf("invalid order (%d)", order)
	}
	agp.order = order
	return nil
}

func limitGroupParser(param string, agp *AccountGroupParam) error {
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

func interestsGroupParser(param string, agp *AccountGroupParam) error {
	ids := globals.Is.ContainsAnyFromInterests([]string{param})
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	agp.interestContain = param
	return nil
}

func birthGroupParser(param string, agp *AccountGroupParam) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}

	agp.birthYear = birth
	return nil
}

func cityGroupParser(param string, agp *AccountGroupParam) error {
	agp.cityEq = param
	return nil
}

func noopGroupParser(param string, agp *AccountGroupParam) error {
	return nil
}

var accountGroupFuncs = map[string]AccountGroupFunc{
	"sex":       sexGroupParser,
	"likes":     likesGroupParser,
	"country":   countryGroupParser,
	"keys":      keysGroupParser,
	"joined":    joinedGroupParser,
	"query_id":  noopGroupParser,
	"status":    statusGroupParser,
	"order":     orderGroupParser,
	"limit":     limitGroupParser,
	"interests": interestsGroupParser,
	"birth":     birthGroupParser,
	"city":      cityGroupParser,
}

type RawGroupResponse struct {
	Sex       string `json:"sex,omitempty"`
	Status    string `json:"status,omitempty"`
	Interests string `json:"interests,omitempty"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`
	Count     int    `json:"count,omitempty"`
}

type RawGroupResponses struct {
	Groups []*RawGroupResponse `json:"groups"`
}

type GroupResponse struct {
	Sex       int8
	Status    int8
	Interests string
	Country   string
	City      string
}

type GroupResponseCount struct {
	GroupResponse
	Count int
}

func (gr *GroupResponseCount) ToRawGroupResponse() *RawGroupResponse {
	r := RawGroupResponse{}
	if gr.Sex != 0 {
		r.Sex = common.SEXES[gr.Sex-1]
	}
	if gr.Status != 0 {
		r.Status = common.STATUSES[gr.Status-1]
	}
	r.Interests = gr.Interests
	r.Country = gr.Country
	r.City = gr.City
	r.Count = gr.Count
	return &r
}

func (l *RawGroupResponse) Equal(r *RawGroupResponse) bool {
	if l.Sex != r.Sex {
		return false
	}
	if l.Status != r.Status {
		return false
	}
	if l.Interests != r.Interests {
		return false
	}
	if l.Country != r.Country {
		return false
	}
	if l.City != r.City {
		return false
	}
	if l.Count != r.Count {
		return false
	}
	return true
}

func SplitGroupParamsIntoStoreAndFilter(originalAgp *AccountGroupParam) (*AccountGroupParam, store.StoreSource) {
	agp := *originalAgp
	//?
	if agp.likeContain != 0 {
		mp := globals.Ls.IdsContainAllLikes([]int{agp.likeContain})

		var ids []int
		for id, _ := range mp {
			ids = append(ids, id)
		}

		agp.likeContain = 0
		return &agp, store.NewArrayStoreSource(ids)
	}

	// 1/30 if length == 1
	if len(agp.interestContain) > 0 {
		mp := globals.Is.ContainsAllFromInterests([]string{agp.interestContain})

		var ids []int
		for id, _ := range mp {
			ids = append(ids, id)
		}

		agp.interestContain = ""
		return &agp, store.NewArrayStoreSource(ids)
	}

	// default because there're no index
	return &agp, globals.As.NewRangeAccountStoreSource()
}

func GenFilterFromAccountsGroupParams(agp *AccountGroupParam) store.StoreFilterFunc {
	return func(id int) bool {
		me := globals.As.GetStoredAccountWithoutError(id)

		if agp.likeContain != 0 {
			result := globals.Ls.CheckContainAllLikes(id, []int{agp.likeContain})
			if !result {
				return false
			}
		}

		if agp.interestContain != "" {
			result := globals.Is.ContainsAny(id, []string{agp.interestContain})
			if !result {
				return false
			}
		}

		if agp.sexEq != 0 {
			if me.Sex != agp.sexEq {
				return false
			}
		}

		if agp.countryEq != "" {
			if me.Country != globals.As.GetCountryId(agp.countryEq) {
				return false
			}
		}

		if agp.cityEq != "" {
			if me.City != globals.As.GetCityId(agp.cityEq) {
				return false
			}
		}

		if agp.statusEq != 0 {
			if me.Status != agp.statusEq {
				return false
			}
		}

		if agp.joinedYear.Int8 != 0 {
			if me.JoinedYear != agp.joinedYear {
				return false
			}
		}

		if agp.birthYear != 0 {
			from := int(time.Date(agp.birthYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
			to := int(time.Date(agp.birthYear+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
			ok := from <= me.Birth && me.Birth < to
			if !ok {
				return false
			}

		}

		return true
	}
}

func filterIdsFromGroupParam(originalAgp *AccountGroupParam) []int {
	agp, ss := SplitGroupParamsIntoStoreAndFilter(originalAgp)
	sff := GenFilterFromAccountsGroupParams(agp)

	// filter without limit
	ret := store.ApplyFilter(ss, sff, 1e8)

	return ret
}

func grouping(ids []int, agp *AccountGroupParam) []GroupResponseCount {
	mp := map[GroupResponse]int{}
	for _, id := range ids {
		a := globals.As.GetStoredAccountWithoutError(id)
		gr := GroupResponse{}
		var interests []string
		for key, _ := range agp.keys {
			switch key {
			case "country":
				gr.Country = globals.As.IdToCountry(a.Country)
			case "city":
				gr.City = globals.As.IdToCity(a.City)
			case "sex":
				gr.Sex = a.Sex
			case "status":
				gr.Status = a.Status
			case "interests":
				interests = globals.Is.GetInterestStrings(id)
			}
		}
		if interests == nil {
			mp[gr]++
		} else {
			for _, i := range interests {
				gr.Interests = i
				mp[gr]++
			}
		}
	}

	ret := []GroupResponseCount{}
	for k, v := range mp {
		ret = append(ret, GroupResponseCount{k, v})
	}

	return ret
}

func sorting(grc []GroupResponseCount, agp *AccountGroupParam) {
	less := func(i, j int) bool {
		if grc[i].Count != grc[j].Count {
			return grc[i].Count < grc[j].Count
		}
		if grc[i].Country != grc[j].Country {
			return grc[i].Country < grc[j].Country
		}
		if grc[i].City != grc[j].City {
			return grc[i].City < grc[j].City
		}
		if grc[i].Interests != grc[j].Interests {
			return grc[i].Interests < grc[j].Interests
		}
		if grc[i].Sex != grc[j].Sex {
			return grc[i].Sex < grc[j].Sex
		}
		if grc[i].Status != grc[j].Status {
			return grc[i].Status < grc[j].Status
		}
		return false
	}

	sortFunc := less
	if agp.order == -1 {
		sortFunc = func(i, j int) bool {
			return less(j, i)
		}
	}

	sort.Slice(grc, sortFunc)
}

func accountsGroupParser(queryParams url.Values) (agp *AccountGroupParam, err error) {
	agp = &AccountGroupParam{
		keys:  map[string]struct{}{},
		limit: -1,
		order: 0,
	}

	for field, param := range queryParams {
		if param[0] == "" {
			err = fmt.Errorf("parameter cannot be empty (field = %s)", field)
			return
		}
		fun, found := accountGroupFuncs[field]
		if !found {
			err = fmt.Errorf("filter (%s) not found", field)
			return
		}
		if len(param) != 1 {
			err = fmt.Errorf("multiple params in filter (%s)", field)
			return
		}
		if err = fun(param[0], agp); err != nil {
			return
		}
	}
	if agp.limit == -1 {
		err = fmt.Errorf("limit is not specified")
		return
	}
	if agp.order == 0 {
		err = fmt.Errorf("order is not specified")
		return
	}
	if len(agp.keys) == 0 {
		err = fmt.Errorf("keys is not specified")
		return
	}
	if len(agp.keys) > 2 {
		log.Printf("keys length = %d : %s", len(agp.keys), common.StringSetJoin(agp.keys, ","))
	}

	return
}

func AccountsGroupCore(queryParams url.Values) ([]GroupResponseCount, *HlcHttpError) {
	agp, err := accountsGroupParser(queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}

	ids := filterIdsFromGroupParam(agp)
	grc := grouping(ids, agp)
	sorting(grc, agp)

	limit := agp.limit
	if limit > len(grc) {
		limit = len(grc)
	}
	return grc[:limit], nil
}

func AccountsGroupHandler(c echo.Context) error {
	grs, err := AccountsGroupCore(c.QueryParams())
	if err != nil {
		return c.String(err.HttpStatusCode, "")
	}

	rgr := RawGroupResponses{[]*RawGroupResponse{}}
	for _, g := range grs {
		rgr.Groups = append(rgr.Groups, g.ToRawGroupResponse())
	}

	return common.JsonResponseWithoutChunking(c, http.StatusOK, &rgr)
}
