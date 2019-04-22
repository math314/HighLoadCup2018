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

func RunTest() {
	testCases := LoadGetPhase("./testdata/answers/phase_1_get.answ")
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
