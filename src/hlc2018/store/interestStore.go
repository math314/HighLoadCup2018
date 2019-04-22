package store

import (
	"hlc2018/common"
)

type InterestStore struct {
	StringIndex
}

func NewInterestStore() *InterestStore {
	return &InterestStore{*NewStringIndex()}
}

func (is *InterestStore) InsertCommonInterest(interest *common.Interest) {
	is.SetValue(interest.AccountId, interest.Interest)
}

func (is *InterestStore) ContainsAllFromInterests(vs []string) map[int]struct{} {
	var mp map[int]struct{}
	for _, s := range vs {
		interestId, found := is.sim.GetWithFound(s)
		if !found {
			return map[int]struct{}{}
		}
		if mp == nil {
			mp = is.valueToId[interestId]
		} else {
			mp = common.MapIntersect(mp, is.valueToId[interestId])
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
		for k, _ := range is.valueToId[interestId] {
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
		if _, ok := is.idToValue[id][interestId]; !ok {
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
		if _, ok := is.idToValue[id][interestId]; ok {
			return true
		}
	}
	return false
}

func (is *InterestStore) GetCommonInterests(id int) []*common.Interest {
	var ret []*common.Interest
	for interestId, _ := range is.idToValue[id] {
		ret = append(ret, &common.Interest{id, is.sim.strings[interestId]})
	}
	return ret
}

func (is *InterestStore) GetSuggestInterestIds(id int) map[int]int {
	mp := map[int]int{}
	for interestId, _ := range is.idToValue[id] {
		for k, _ := range is.valueToId[interestId] {
			mp[k]++
		}
	}

	return mp
}
