package store

import (
	"fmt"
	"hlc2018/common"
)

type InterestStore struct {
	StringIndex
}

func NewInterestStore() *InterestStore {
	return &InterestStore{*NewStringIndex()}
}

func (is *InterestStore) InsertCommonInterest(interest *common.Interest) {
	is.SetString(interest.AccountId, interest.Interest)
}

func (is *InterestStore) ContainsAllFromInterests(vs []string) map[int]struct{} {
	var mp map[int]struct{}
	for _, s := range vs {
		interestId, found := is.sim.GetWithFound(s)
		if !found {
			return map[int]struct{}{}
		}
		if mp == nil {
			mp = is.stringIdToPk[interestId]
		} else {
			mp = common.MapIntersect(mp, is.stringIdToPk[interestId])
		}
	}
	return mp
}

func (is *InterestStore) ContainsAnyFromInterests(vs []string) map[int]struct{} {
	mp := map[int]struct{}{}
	for _, s := range vs {
		interestId, found := is.sim.GetWithFound(s)
		if !found {
			continue
		}
		for k, _ := range is.stringIdToPk[interestId] {
			mp[k] = struct{}{}
		}
	}
	return mp
}

func (is *InterestStore) ContainsAll(id int, vs []string) bool {
	for _, s := range vs {
		interestId, found := is.sim.GetWithFound(s)
		if !found {
			return false
		}
		if _, ok := is.pkToStringId[id][interestId]; !ok {
			return false
		}
	}
	return true
}

func (is *InterestStore) ContainsAny(id int, vs []string) bool {
	for _, s := range vs {
		interestId, found := is.sim.GetWithFound(s)
		if !found {
			continue
		}
		if _, ok := is.pkToStringId[id][interestId]; ok {
			return true
		}
	}
	return false
}

func (is *InterestStore) GetInterestStrings(id int) []string {
	ret := []string{}
	for interestId, _ := range is.pkToStringId[id] {
		ret = append(ret, is.sim.strings[interestId])
	}
	return ret
}

func (is *InterestStore) GetCommonInterests(id int) []*common.Interest {
	var ret []*common.Interest
	for interestId, _ := range is.pkToStringId[id] {
		ret = append(ret, &common.Interest{id, is.sim.strings[interestId]})
	}
	return ret
}

func (is *InterestStore) GetSuggestInterestIds(id int) map[int]int {
	mp := map[int]int{}
	for interestId, _ := range is.pkToStringId[id] {
		for k, _ := range is.stringIdToPk[interestId] {
			mp[k]++
		}
	}

	return mp
}

func (is *InterestStore) UpdateInterests(id int, interests []string) error {
	if id >= len(is.pkToStringId) {
		return fmt.Errorf("id out of range : %d", id)
	}
	is.StringIndex.DeleteStringsFromPk(id)
	for _, s := range interests {
		is.StringIndex.SetString(id, s)
	}
	return nil
}
