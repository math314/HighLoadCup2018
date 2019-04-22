package tester

import (
	"encoding/json"
	"hlc2018/common"
	"hlc2018/handlers"
	"log"
)

func testAccountsFilter(args *testRouterCallbackArgs) {
	ansAfa := common.RawAccountsContainer{}
	if args.status == 200 {
		if err := json.Unmarshal([]byte(args.json), &ansAfa); err != nil {
			log.Fatal(err)
		}
	}

	afa, err := handlers.AccountsFilterCore(args.url.Query())
	if args.status != 200 {
		if err == nil || args.status != err.HttpStatusCode {
			log.Fatal(args.url, "status mismatch")
		}
		return
	}

	if err != nil {
		log.Fatal(args.url, err)
	}

	if len(ansAfa.Accounts) != len(afa.Accounts) {
		log.Print("length mismatch")
		handlers.AccountsFilterCore(args.url.Query())
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa.Accounts[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Print("item mismatch")
			handlers.AccountsFilterCore(args.url.Query())
		}
	}
}

func testAccountsGroup(args *testRouterCallbackArgs) {
	ansGr := handlers.RawGroupResponses{}
	if args.status == 200 {
		if err := json.Unmarshal([]byte(args.json), &ansGr); err != nil {
			log.Fatal(err)
		}
	}

	_gr, err := handlers.AccountsGroupCore(args.url.Query())
	if args.status != 200 {
		if err == nil || args.status != err.HttpStatusCode {
			log.Fatal(args.url, "status mismatch")
		}
		return
	}

	if err != nil {
		log.Fatal(args.url, err)
	}

	if len(ansGr.Groups) != len(_gr) {
		log.Fatal("length mismatch")
	}

	for i := 0; i < len(ansGr.Groups); i++ {
		r := _gr[i].ToRawGroupResponse()
		if !ansGr.Groups[i].Equal(r) {
			log.Fatal("item mismatch")
		}
	}
}

func testAccountsRecommend(args *testRouterCallbackArgs) {
	ansAfa := common.RawAccountsContainer{}
	if args.status == 200 {
		if err := json.Unmarshal([]byte(args.json), &ansAfa); err != nil {
			log.Fatal(err)
		}
	}

	afa, err := handlers.AccountsRecommendCore(args.matched[1], args.url.Query())
	if args.status != 200 {
		if err == nil || args.status != err.HttpStatusCode {
			log.Fatal(args.url, "status mismatch")
		}
		return
	}

	if err != nil {
		log.Fatal(args.url, err)
	}

	if len(ansAfa.Accounts) != len(afa) {
		log.Fatal("length mismatch")
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Fatal("item mismatch")
		}
	}
}

func testAccountsSuggest(args *testRouterCallbackArgs) {
	ansAfa := common.RawAccountsContainer{}
	if args.status == 200 {
		if err := json.Unmarshal([]byte(args.json), &ansAfa); err != nil {
			log.Fatal(err)
		}
	}

	afa, err := handlers.AccountsSuggestCore(args.matched[1], args.url.Query())
	if args.status != 200 {
		if err == nil || args.status != err.HttpStatusCode {
			log.Fatal(args.url, "status mismatch")
		}
		return
	}

	if err != nil {
		log.Fatal(args.url, err)
	}

	if len(ansAfa.Accounts) != len(afa) {
		log.Fatal("length mismatch")
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Fatal("item mismatch")
		}
	}
}
