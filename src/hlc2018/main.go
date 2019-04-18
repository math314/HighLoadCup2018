package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"hlc2018/common"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	db *sqlx.DB
)

func init() {
	db_name := os.Getenv("MYSQL_DATABASE")
	if db_name == "" {
		db_name = "hlc2018"
	}
	db_host := os.Getenv("MYSQL_HOST")
	if db_host == "" {
		db_host = "127.0.0.1"
	}
	db_port := os.Getenv("MYSQL_PORT")
	if db_port == "" {
		db_port = "3306"
	}
	db_user := os.Getenv("MYSQL_USER")
	if db_user == "" {
		db_user = "hlc"
	}
	db_password := os.Getenv("MYSQL_PASSWORD")
	if db_password == "" {
		db_password = "hlc"
	}
	db_password = ":" + db_password

	dsn := fmt.Sprintf("%s%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		db_user, db_password, db_host, db_port, db_name)

	log.Printf("Connecting to db: %q", dsn)
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Printf("Succeeded to connect db.")
}

type sqlBuilder struct {
	where bytes.Buffer
	limit int
}

func (sb *sqlBuilder) addWhere(s string) {
	if sb.where.Len() != 0 {
		sb.where.WriteString(" AND ")
	}
	sb.where.WriteString(s)
}

func SexEqFilter(param string, sb *sqlBuilder) error {
	sex := common.SexFromString(param)
	if sex == -1 {
		return fmt.Errorf("%s is not valid sex", param)
	}
	sb.addWhere(fmt.Sprintf("sex = %d", sex))
	return nil
}

func emailDomainFilter(param string, sb *sqlBuilder) error {
	if strings.Contains(param, "%") {
		return fmt.Errorf("domain (%s) cannot contain \"%%\"", param)
	}
	sb.addWhere(fmt.Sprintf("email like \"%s\"", "%"+param))
	return nil
}

func emailLtFilter(param string, sb *sqlBuilder) error {
	sb.addWhere(fmt.Sprintf("email <= \"%s\"", param))
	return nil
}

func emailGtFilter(param string, sb *sqlBuilder) error {
	sb.addWhere(fmt.Sprintf("email >= \"%s\"", param))
	return nil
}

func StatusEqFilter(param string, sb *sqlBuilder) error {
	status := common.StatusFromString(param)
	if status == -1 {
		return fmt.Errorf("%s is not valid status", param)
	}
	sb.addWhere(fmt.Sprintf("status = %d", status))
	return nil
}

func StatusNeqFilter(param string, sb *sqlBuilder) error {
	status := common.StatusFromString(param)
	if status == -1 {
		return fmt.Errorf("%s is not valid status", param)
	}
	sb.addWhere(fmt.Sprintf("status != %d", status))
	return nil
}

func fnameAnyFilter(param string, sb *sqlBuilder) error {
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")
	sb.addWhere(fmt.Sprintf("fname in (%s)", arg))
	return nil
}

func snameStartsFilter(param string, sb *sqlBuilder) error {
	sb.addWhere(fmt.Sprintf("sname LIKE \"%s\"", param+"%"))
	return nil
}

func phoneCodeFilter(param string, sb *sqlBuilder) error {
	if len(param) != 3 {
		return fmt.Errorf("phone code param length should be 3 but %s", len(param))
	}
	for _, c := range param {
		if c < '0' || c > '9' {
			return fmt.Errorf("phone code param should be [0-9] : %s", param)
		}
	}
	sb.addWhere(fmt.Sprintf("phone LIKE \"%s\"", "%("+param+")%"))
	return nil
}

func cityAnyFilter(param string, sb *sqlBuilder) error {
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")

	sb.addWhere(fmt.Sprintf("city in (%s)", arg))
	return nil
}

func birthLtFilter(param string, sb *sqlBuilder) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	sb.addWhere(fmt.Sprintf("birth < %d", birth))
	return nil
}

func birthGtFilter(param string, sb *sqlBuilder) error {
	birth, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth (%s)", param)
	}
	sb.addWhere(fmt.Sprintf("birth > %d", birth))
	return nil
}

func birthYearFilter(param string, sb *sqlBuilder) error {
	year, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse birth year (%s)", param)
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	sb.addWhere(fmt.Sprintf("birth >= %d", from.Unix()))
	sb.addWhere(fmt.Sprintf("birth < %d", after.Unix()))
	return nil
}

func premiumNowFilter(param string, sb *sqlBuilder) error {
	//y, m, d := time.Now().In(time.UTC).Date()
	from := time.Date(2019, 1, 24, 0, 0, 0, 0, time.UTC)
	after := time.Date(2019, 1, 25, 0, 0, 0, 0, time.UTC)

	sb.addWhere(fmt.Sprintf("premium_start <= %d", from.Unix()))
	sb.addWhere(fmt.Sprintf("premium_end >= %d", after.Unix()))
	return nil
}

func limitFilter(param string, sb *sqlBuilder) error {
	limit, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("failed to parse limit (%s)", param)
	}
	sb.limit = limit
	return nil
}

func interestsAnyFilter(param string, sb *sqlBuilder) error {
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")

	sb.addWhere(fmt.Sprintf("id in (SELECT DISTINCT(account_id) from interests WHERE interest in (%s))", arg))
	return nil
}

func interestsContainsFilter(param string, sb *sqlBuilder) error {
	names := strings.Split(param, ",")
	for i := 0; i < len(names); i++ {
		names[i] = "\"" + names[i] + "\""
	}
	arg := strings.Join(names, ",")

	paramLen := strings.Count(param, ",") + 1
	sb.addWhere(fmt.Sprintf("id in (SELECT account_id from interests WHERE interest in (%s) GROUP BY account_id HAVING COUNT(account_id) = %d)", arg, paramLen))
	return nil
}

func likesContainsFilter(param string, sb *sqlBuilder) error {
	for _, s := range strings.Split(param, ",") {
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("failed to parse likes (%s)", param)
		}
	}

	paramLen := strings.Count(param, ",") + 1
	sb.addWhere(fmt.Sprintf("id in (SELECT account_id_from from likes WHERE account_id_to in (%s) GROUP BY account_id_from HAVING COUNT(account_id_from) = %d)", param, paramLen))
	return nil
}

func premiumNullFilter(param string, sb *sqlBuilder) error {
	if param == "0" {
		sb.addWhere("premium_start != 0")
	} else if param == "1" {
		sb.addWhere("premium_start = 0")
	} else {
		return fmt.Errorf("premium param is not valid (%s)", param)
	}
	return nil
}

func noopFilter(param string, sb *sqlBuilder) error {
	return nil
}

type FilterFunc func(param string, sb *sqlBuilder) error

func eqFilterGenerator(name string) FilterFunc {
	return func(param string, sb *sqlBuilder) error {
		sb.addWhere(fmt.Sprintf("%s = \"%s\"", name, param))
		return nil
	}
}

func nullFilterGenerator(name string) FilterFunc {
	return func(param string, sb *sqlBuilder) error {
		if param == "0" {
			sb.addWhere(fmt.Sprintf("%s IS NOT NULL", name))
		} else if param == "1" {
			sb.addWhere(fmt.Sprintf("%s IS NULL", name))
		} else {
			return fmt.Errorf("%s param is not valid (%s)", name, param)
		}
		return nil
	}
}

var usualFilters = map[string]FilterFunc{
	"sex_eq":             SexEqFilter,
	"email_domain":       emailDomainFilter,
	"email_lt":           emailLtFilter,
	"email_gt":           emailGtFilter,
	"status_eq":          StatusEqFilter,
	"status_neq":         StatusNeqFilter,
	"fname_eq":           eqFilterGenerator("fname"),
	"fname_any":          fnameAnyFilter,
	"fname_null":         nullFilterGenerator("fname"),
	"sname_eq":           eqFilterGenerator("sname"),
	"sname_starts":       snameStartsFilter,
	"sname_null":         nullFilterGenerator("sname"),
	"phone_code":         phoneCodeFilter,
	"phone_null":         nullFilterGenerator("phone"),
	"country_eq":         eqFilterGenerator("country"),
	"country_null":       nullFilterGenerator("country"),
	"city_eq":            eqFilterGenerator("city"),
	"city_any":           cityAnyFilter,
	"city_null":          nullFilterGenerator("city"),
	"birth_lt":           birthLtFilter,
	"birth_gt":           birthGtFilter,
	"birth_year":         birthYearFilter,
	"interests_contains": interestsContainsFilter,
	"interests_any":      interestsAnyFilter,
	"likes_contains":     likesContainsFilter,
	"premium_now":        premiumNowFilter,
	"premium_null":       premiumNullFilter,
	"limit":              limitFilter,
	"query_id":           noopFilter,
}

func accountsFilter(queryParams url.Values) (sb sqlBuilder, err error) {
	sb = sqlBuilder{bytes.Buffer{}, -1}

	for field, param := range queryParams {
		fun, found := usualFilters[field]
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

type AccountsFilterAccount struct {
	Accounts []struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	} `json:"accounts"`
}

type FilterError struct {
	httpStatusCode int
	Err            error
}

func (e *FilterError) Error() string {
	return "status: " + strconv.Itoa(e.httpStatusCode) + ", error: " + e.Err.Error()
}

func accountsFilterCore(queryParams url.Values) (*AccountsFilterAccount, *FilterError) {
	sb, err := accountsFilter(queryParams)
	if err != nil {
		log.Print(err)
		return &AccountsFilterAccount{}, &FilterError{http.StatusBadRequest, err}
	}
	whereCluster := ""
	if sb.where.Len() != 0 {
		whereCluster = "WHERE " + sb.where.String()
	}

	query := fmt.Sprintf("SELECT id, email FROM accounts %s ORDER BY id DESC LIMIT %d", whereCluster, sb.limit)
	log.Printf("query := %s", query)

	afas := AccountsFilterAccount{Accounts: []struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}{}}
	if err := db.Select(&afas.Accounts, query); err != nil {
		log.Print(err)
		return &AccountsFilterAccount{}, &FilterError{http.StatusInternalServerError, err}
	}
	return &afas, nil
}

func accountsFilterHandler(c echo.Context) error {
	afas, err := accountsFilterCore(c.QueryParams())
	if err != nil {
		return c.String(err.httpStatusCode, "")
	}

	return c.JSON(http.StatusOK, afas)
}

func accountsGroupHandler(c echo.Context) error {
	return nil
}

func accountsRecommendHandler(c echo.Context) error {
	return nil
}

func accountsSuggestHandler(c echo.Context) error {
	return nil
}

func accountsNewHandler(c echo.Context) error {
	return nil
}

func accountsIdHandler(c echo.Context) error {
	return nil
}

func accountsLikesHandler(c echo.Context) error {
	return nil
}

func loadAnsw(pathRegex string, callback func(url *url.URL, status int, json string)) {
	fp, err := os.Open("./testdata/answers/phase_1_get.answ")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	r := regexp.MustCompile(pathRegex)
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		tmp := strings.Split(line, "\t")
		url, err := url.Parse(tmp[1])
		if err != nil {
			log.Fatal(err)
		}
		if !r.MatchString(url.Path) {
			continue
		}

		status, err := strconv.Atoi(tmp[2])
		if err != nil {
			log.Fatal(err)
		}
		j := ""
		if len(tmp) > 3 {
			j = tmp[3]
		}
		callback(url, status, j)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func testAccountsFilter(c echo.Context) error {
	loadAnsw(`/accounts/filter/`, func(url *url.URL, status int, j string) {
		ansAfa := AccountsFilterAccount{}
		if status == 200 {
			if err := json.Unmarshal([]byte(j), &ansAfa); err != nil {
				log.Fatal(err)
			}
		}

		afa, err := accountsFilterCore(url.Query())
		if status != 200 {
			if err == nil || status != err.httpStatusCode {
				log.Fatal(url, "status mismatch")
			}
			return
		}

		if err != nil {
			log.Fatal(url, err)
		}

		mp := map[int]string{}
		for _, a := range afa.Accounts {
			mp[a.ID] = a.Email
		}
		idsStr := []string{}
		for _, a := range ansAfa.Accounts {
			idsStr = append(idsStr, strconv.Itoa(a.ID))
			if mp[a.ID] != a.Email {
				log.Printf("id = %d not found", a.ID)
			}
		}
		log.Print(strings.Join(idsStr, ","))
	})

	return c.HTML(http.StatusOK, "tested")
}

func httpMain() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "request:\"${method} ${uri}\" status:${status} latency:${latency} (${latency_human}) bytes:${bytes_out}\n",
	}))

	e.GET("/accounts/filter/", accountsFilterHandler)
	e.GET("/accounts/group/", accountsGroupHandler)
	e.GET("/accounts/:id/recommend/", accountsRecommendHandler)
	e.GET("/accounts/:id/suggest/", accountsSuggestHandler)
	e.POST("/accounts/new/", accountsNewHandler)
	e.POST("/accounts/:id/", accountsIdHandler)
	e.POST("/accounts/likes/", accountsLikesHandler)

	e.GET("/tests/filter", testAccountsFilter)
	e.Start(":" + port)
}

type RawAccountsContainer struct {
	Accounts []struct {
		ID        int      `json:"id"`
		Fname     string   `json:"fname"`
		Sname     string   `json:"sname"`
		Email     string   `json:"email"`
		Interests []string `json:"interests"`
		Status    string   `json:"status"`
		Premium   struct {
			Start  int `json:"start"`
			Finish int `json:"finish"`
		} `json:"premium"`
		Sex   string `json:"sex"`
		Phone string `json:"phone"`
		Likes []struct {
			Ts int `json:"ts"`
			ID int `json:"id"`
		} `json:"likes"`
		Birth   int    `json:"birth"`
		City    string `json:"city"`
		Country string `json:"country"`
		Joined  int    `json:"joined"`
	} `json:"accounts"`
}

func loadDataInFile(tableName string, fields string, data []byte) {
	if err := ioutil.WriteFile("/tmp/tmpload.txt", data, 0644); err != nil {
		log.Fatal(err)
	}

	query := fmt.Sprintf("LOAD DATA INFILE '/tmp/tmpload.txt' INTO TABLE %s FIELDS TERMINATED BY ',' LINES TERMINATED BY '\n'", tableName)
	if fields != "" {
		query = fmt.Sprintf("%s %s", query, fields)
	}
	result, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(result)
}

func mysqlDataLoader() {
	r, err := zip.OpenReader("/tmp/data/data.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Fatal(err)
		}

		var ac RawAccountsContainer
		if err := json.Unmarshal(b, &ac); err != nil {
			log.Fatal(err)
		}

		accounts := make([]common.Account, 0)
		interests := make([]common.Interest, 0)
		likes := make([]common.Like, 0)

		for _, rawAccount := range ac.Accounts {
			var a common.Account
			a.ID = rawAccount.ID
			a.Fname = rawAccount.Fname
			a.Sname = rawAccount.Sname
			a.Email = rawAccount.Email
			a.Status = common.StatusFromString(rawAccount.Status)
			a.Premium_start = rawAccount.Premium.Start
			a.Premium_end = rawAccount.Premium.Finish
			a.Sex = common.SexFromString(rawAccount.Sex)
			a.Phone = rawAccount.Phone
			a.Birth = rawAccount.Birth
			a.City = rawAccount.City
			a.Country = rawAccount.Country
			a.Joined = rawAccount.Joined
			accounts = append(accounts, a)

			for _, i := range rawAccount.Interests {
				interest := common.Interest{a.ID, i}
				interests = append(interests, interest)
			}

			for _, l := range rawAccount.Likes {
				like := common.Like{a.ID, l.ID, l.Ts}
				likes = append(likes, like)
			}
		}

		sb := bytes.Buffer{}
		for _, a := range accounts {
			sb.WriteString(a.Oneline())
		}
		loadDataInFile("accounts", "", sb.Bytes())

		sb = bytes.Buffer{}
		for _, i := range interests {
			sb.WriteString(i.Oneline())
		}
		loadDataInFile("interests", "(account_id, interest)", sb.Bytes())

		sb = bytes.Buffer{}
		for _, l := range likes {
			sb.WriteString(l.Oneline())
		}
		loadDataInFile("likes", "(account_id_from, account_id_to, ts)", sb.Bytes())
	}
}

func main() {
	load := flag.Bool("l", false, "run mysqlDataLoader")
	flag.Parse()
	if *load {
		mysqlDataLoader()
	} else {
		httpMain()
	}
}
