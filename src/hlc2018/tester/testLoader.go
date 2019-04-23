package tester

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type testRouter struct {
	pathRegex *regexp.Regexp
	fun       getPhaseFunc
}

type testCase struct {
	url    *url.URL
	status int
	json   string
}

type testRouterCallbackArgs struct {
	testCase
	matched []string
}

type getPhaseFunc func(args *testRouterCallbackArgs)

var testRouters = []testRouter{
	testRouter{regexp.MustCompile("/accounts/filter/$"), testAccountsFilter},
	testRouter{regexp.MustCompile("/accounts/group/$"), testAccountsGroup},
	testRouter{regexp.MustCompile("/accounts/([0-9]+)/recommend/$"), testAccountsRecommend},
	testRouter{regexp.MustCompile("/accounts/([0-9]+)/suggest/$"), testAccountsSuggest},
	testRouter{regexp.MustCompile("/accounts/new/$"), testAccountsInsert},
	testRouter{regexp.MustCompile("/accounts/([0-9]+)/$"), testAccountsUpdate},
	testRouter{regexp.MustCompile("/accounts/likes/$"), testAccountsLike},
}

func LoadGetPhase(fileName string) []*testCase {
	fp, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	var ret []*testCase
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		tmp := strings.Split(line, "\t")
		url, err := url.Parse(tmp[1])
		if err != nil {
			log.Fatal(err)
		}

		status, err := strconv.Atoi(tmp[2])
		if err != nil {
			log.Fatal(err)
		}
		j := ""
		if len(tmp) > 3 {
			j = tmp[3]
		}

		ret = append(ret, &testCase{url, status, j})
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return ret
}

type ammo struct {
	id   int
	url  *url.URL
	json string
}

func LoadAmmo(ammoFile string) []*ammo {
	fp, err := os.Open(ammoFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	var ret []*ammo
	scanner := bufio.NewScanner(fp)
	for {
		var lines []string
		for i := 0; i < 10; i++ {
			if !scanner.Scan() {
				return ret
			}
			text := scanner.Text()
			lines = append(lines, text)
			if text == "" {
				break
			}
		}
		if lines[len(lines)-3] == "Content-Length: 0" {
			lines = append(lines, "")
		} else {
			scanner.Scan()
			lines = append(lines, scanner.Text())
		}

		tmp := strings.Split(lines[1], " ")
		url, err := url.Parse(tmp[1])
		if err != nil {
			log.Fatal(err)
		}

		id, err := strconv.Atoi(url.Query().Get("query_id"))
		if err != nil {
			log.Fatal(err)
		}

		a := &ammo{id, url, lines[len(lines)-1]}
		ret = append(ret, a)
	}
	return ret
}

type answ struct {
	id     int
	status int
}

func LoadAnsw(answFile string) []*answ {
	fp, err := os.Open(answFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	var ret []*answ
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		tmp := strings.Split(line, "\t")
		url, err := url.Parse(tmp[1])
		if err != nil {
			log.Fatal(err)
		}

		status, err := strconv.Atoi(tmp[2])
		if err != nil {
			log.Fatal(err)
		}
		id, err := strconv.Atoi(url.Query().Get("query_id"))
		if err != nil {
			log.Fatal(err)
		}

		ret = append(ret, &answ{id, status})
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return ret
}

func LoadPostPhase(ammoFile string, answFile string) []*testCase {
	ammos := LoadAmmo(ammoFile)
	answs := LoadAnsw(answFile)
	mp := map[int]*ammo{}
	for _, ammo := range ammos {
		mp[ammo.id] = ammo
	}

	var cases []*testCase
	for _, answ := range answs {
		ammo := mp[answ.id]
		cases = append(cases, &testCase{ammo.url, answ.status, ammo.json})
	}

	return cases
}

func RunTestCases(testCases []*testCase) {
	for _, testCase := range testCases {
		var routed bool
		for _, route := range testRouters {
			matched := route.pathRegex.FindStringSubmatch(testCase.url.Path)
			if matched == nil {
				continue
			}

			routed = true
			args := &testRouterCallbackArgs{*testCase, matched}

			before := time.Now().UnixNano()
			route.fun(args)
			after := time.Now().UnixNano()
			elaspedNano := after - before
			log.Printf("%d ms (%s)", elaspedNano/1000/1000, testCase.url.String())

		}
		if !routed {
			if testCase.status != http.StatusNotFound {
				log.Fatalf("%s is not routed", testCase.url.String())
			}
		}
	}
}

func RunTest() {
	testCases1 := LoadGetPhase("./testdata/answers/phase_1_get.answ")
	RunTestCases(testCases1)
	testCases2 := LoadPostPhase("./testdata/ammo/phase_2_post.ammo", "./testdata/answers/phase_2_post.answ")
	RunTestCases(testCases2)
	testCases3 := LoadGetPhase("./testdata/answers/phase_3_get.answ")
	RunTestCases(testCases3)
}
