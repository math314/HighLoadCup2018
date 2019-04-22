package handlers

import (
	"bytes"
	"database/sql"
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

type AccountGroupParam struct {
	keys   map[string]struct{}
	froms  map[string]struct{}
	wheres bytes.Buffer
	limit  int
	order  int
}

func (agp *AccountGroupParam) addFrom(s string) {
	agp.froms[s] = struct{}{}
}

func (agp *AccountGroupParam) addWhere(s string) {
	if agp.wheres.Len() != 0 {
		agp.wheres.WriteString(" AND ")
	}
	agp.wheres.WriteString(s)
}

type AccountGroupFunc func(param string, agp *AccountGroupParam) error

func sexGroupParser(param string, agp *AccountGroupParam) error {
	sex := common.SexFromString(param)
	if sex == 0 {
		return fmt.Errorf("%s is not valid sex", param)
	}
	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("a.sex = %d", sex))
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

	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("a.id in (%s)", common.IntArrayJoin(liker, ",")))
	return nil
}

func countryGroupParser(param string, agp *AccountGroupParam) error {
	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("a.country = \"%s\"", param))
	return nil
}

func keysGroupParser(param string, agp *AccountGroupParam) error {
	validKeys := []string{"sex", "status", "interests", "country", "city"}
	for _, k := range strings.Split(param, ",") {
		if common.SliceIndex(validKeys, k) == -1 {
			return fmt.Errorf("invalid keys (%s)", k)
		}
		agp.keys[k] = struct{}{}

		if k == "interests" {
			agp.addFrom("accounts")
			agp.addFrom("interests")
			agp.addWhere("a.id = i.account_id")
		} else {
			agp.addFrom("accounts")
		}
	}
	return nil
}

func joinedGroupParser(param string, agp *AccountGroupParam) error {
	joined, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse joined (%s)", param)
	}

	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("a.joined_year = %d", joined-2000))

	return nil
}

func statusGroupParser(param string, agp *AccountGroupParam) error {
	status := common.StatusFromString(param)
	if status == 0 {
		return fmt.Errorf("%s is not valid status", param)
	}
	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("a.status = %d", status))
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
	agp.addFrom("accounts")
	ids := globals.Is.ContainsAnyFromInterests([]string{param})
	if len(ids) == 0 {
		ids[-1] = struct{}{}
	}
	agp.addWhere(fmt.Sprintf("a.id in (%s)", common.IntSetJoin(ids, ",")))
	return nil
}

func birthGroupParser(param string, agp *AccountGroupParam) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}

	from := time.Date(birth, 1, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(birth+1, 1, 1, 0, 0, 0, 0, time.UTC)

	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("birth >= %d", from.Unix()))
	agp.addWhere(fmt.Sprintf("birth < %d", after.Unix()))

	return nil
}

func cityGroupParser(param string, agp *AccountGroupParam) error {
	agp.addFrom("accounts")
	agp.addWhere(fmt.Sprintf("city = \"%s\"", param))
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
	Sex       int            `db:"sex"`
	Status    int            `db:"status"`
	Interests string         `db:"interests"`
	Country   sql.NullString `db:"country"`
	City      sql.NullString `db:"city"`
	Count     int            `db:"count"`
}

func (gr *GroupResponse) ToRawGroupResponse() *RawGroupResponse {
	r := RawGroupResponse{}
	if gr.Sex != 0 {
		r.Sex = common.SEXES[gr.Sex-1]
	}
	if gr.Status != 0 {
		r.Status = common.STATUSES[gr.Status-1]
	}
	r.Interests = gr.Interests
	r.Country = gr.Country.String
	r.City = gr.City.String
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

func accountsGroupParser(queryParams url.Values) (agp *AccountGroupParam, err error) {
	agp = &AccountGroupParam{map[string]struct{}{}, map[string]struct{}{}, bytes.Buffer{}, -1, 0}

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

func AccountsGroupCore(queryParams url.Values) ([]GroupResponse, *HlcHttpError) {
	agp, err := accountsGroupParser(queryParams)
	if err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusBadRequest, err}
	}

	selectCluster := bytes.Buffer{}
	selectCluster.WriteString("COUNT(*) AS `count`")
	for k, _ := range agp.keys {
		selectCluster.WriteString(", ")
		if k == "interests" {
			selectCluster.WriteString("i.interest AS interests")
		} else {
			selectCluster.WriteString("a." + k + " AS " + k)
		}
	}

	fromCluster := bytes.Buffer{}
	for k, _ := range agp.froms {
		if fromCluster.Len() != 0 {
			fromCluster.WriteString(", ")
		}

		fromCluster.WriteString(k + " AS ")
		fromCluster.WriteByte(k[0])
	}

	whereClusterWithWHERE := ""
	if agp.wheres.Len() != 0 {
		whereClusterWithWHERE = "WHERE " + agp.wheres.String()
	}

	groupByCluster := bytes.Buffer{}
	for k, _ := range agp.keys {
		if groupByCluster.Len() != 0 {
			groupByCluster.WriteString(", ")
		}
		groupByCluster.WriteString(k)
	}

	orderByCluster := bytes.Buffer{}
	{
		order := "ASC"
		if agp.order == -1 {
			order = "DESC"
		}
		orderByCluster.WriteString("`count` " + order)
		for _, v := range []string{"country", "city", "interests", "sex", "status"} {
			if _, ok := agp.keys[v]; !ok {
				continue
			}
			orderByCluster.WriteString(", " + v + " " + order)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s /*WHERE*/%s GROUP BY %s ORDER BY %s LIMIT %d",
		selectCluster.String(), fromCluster.String(), whereClusterWithWHERE, groupByCluster.String(),
		orderByCluster.String(), agp.limit)
	log.Printf("query := %s", query)

	var grs []GroupResponse
	if err := globals.DB.Select(&grs, query); err != nil {
		log.Print(err)
		return nil, &HlcHttpError{http.StatusInternalServerError, err}
	}

	return grs, nil
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
