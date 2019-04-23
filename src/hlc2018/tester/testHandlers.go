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
		log.Fatal("length mismatch")
		handlers.AccountsFilterCore(args.url.Query())
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa.Accounts[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Fatal("item mismatch")
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
		handlers.AccountsGroupCore(args.url.Query())
		return
	}

	for i := 0; i < len(ansGr.Groups); i++ {
		r := _gr[i].ToRawGroupResponse()
		if !ansGr.Groups[i].Equal(r) {
			log.Fatal("item mismatch")
			handlers.AccountsGroupCore(args.url.Query())
			return
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
		handlers.AccountsRecommendCore(args.matched[1], args.url.Query())
		return
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Fatal("item mismatch")
			handlers.AccountsRecommendCore(args.matched[1], args.url.Query())
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
		return
	}

	for i := 0; i < len(ansAfa.Accounts); i++ {
		r := afa[i].ToRawAccount()
		if !ansAfa.Accounts[i].Equal(r) {
			log.Fatal("item mismatch")
			return
		}
	}
}

func testAccountsInsert(args *testRouterCallbackArgs) {
	err := handlers.AccountsInsertHandlerCore([]byte(args.json))
	if err != nil {
		if args.status != 400 {
			log.Fatal("status error")
			handlers.AccountsInsertHandlerCore([]byte(args.json))
		}
	} else {
		if args.status != 201 {
			log.Fatal("status error")
			handlers.AccountsInsertHandlerCore([]byte(args.json))
		}
	}
}

func testAccountsUpdate(args *testRouterCallbackArgs) {
	err := handlers.AccountsUpdateHandlerCore(args.matched[1], []byte(args.json))
	if err != nil {
		if args.status != err.HttpStatusCode {
			log.Fatal("status error")
			handlers.AccountsUpdateHandlerCore(args.matched[1], []byte(args.json))
		}
	} else {
		if args.status != 202 {
			log.Fatal("status error")
			handlers.AccountsUpdateHandlerCore(args.matched[1], []byte(args.json))
		}
	}
}

func testAccountsLike(args *testRouterCallbackArgs) {
	err := handlers.AccountsLikesHandlerCore([]byte(args.json))
	if err != nil {
		if args.status != 400 {
			log.Fatal("status error")
		}
	} else {
		if args.status != 202 {
			log.Fatal("status error")
		}
	}
}
