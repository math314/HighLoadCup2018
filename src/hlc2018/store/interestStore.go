package store

import (
	"hlc2018/common"
)

type InterestStore struct {
	interestMapper StringIdMapper
	idToInterests  []map[int]struct{}
	InterestsToId  []map[int]struct{}
}

func NewInterestStore() *InterestStore {
	is := &InterestStore{StringIdMapper{map[string]int{}, nil}, nil, nil}
	is.insertInterestIfNeeded("")
	return is
}

func (is *InterestStore) ExtendSizeIfNeeded(nextSize int) {
	for len(is.idToInterests) < nextSize {
		is.idToInterests = append(is.idToInterests, map[int]struct{}{})
	}
}

func (is *InterestStore) insertInterestIfNeeded(s string) int {
	val, added := is.interestMapper.insertStringIfNeeded(s)
	if added {
		is.InterestsToId = append(is.InterestsToId, map[int]struct{}{})
	}
	return val
}

func (is *InterestStore) setInterests(id int, s string) {
	is.ExtendSizeIfNeeded(id + 1)
	interestId := is.insertInterestIfNeeded(s)
	is.idToInterests[id][interestId] = struct{}{}
	is.InterestsToId[interestId][id] = struct{}{}
}

func (is *InterestStore) InsertCommonInterest(interest *common.Interest) {
	is.setInterests(interest.AccountId, interest.Interest)
}

func (is *InterestStore) ContainsAllFromInterests(vs []string) map[int]struct{} {
	var mp map[int]struct{}
	for _, s := range vs {
		interestId, found := is.interestMapper.get(s)
		if !found {
			return map[int]struct{}{}
		}
		if mp == nil {
			mp = is.InterestsToId[interestId]
		} else {
			mp = common.MapIntersect(mp, is.InterestsToId[interestId])
		}
	}
	return mp
}

func (is *InterestStore) ContainsAnyFromInterests(vs []string) map[int]struct{} {
	mp := map[int]struct{}{}
	for _, s := range vs {
		interestId, found := is.interestMapper.get(s)
		if !found {
			continue
		}
		for k, _ := range is.InterestsToId[interestId] {
			mp[k] = struct{}{}
		}
	}
	return mp
}

func (is *InterestStore) ContainsAll(id int, vs []string) bool {
	for _, s := range vs {
		interestId, found := is.interestMapper.get(s)
		if !found {
			return false
		}
		if _, ok := is.idToInterests[id][interestId]; !ok {
			return false
		}
	}
	return true
}

func (is *InterestStore) ContainsAny(id int, vs []string) bool {
	for _, s := range vs {
		interestId, found := is.interestMapper.get(s)
		if !found {
			continue
		}
		if _, ok := is.idToInterests[id][interestId]; ok {
			return true
		}
	}
	return false
}

func (is *InterestStore) GetCommonInterests(id int) []*common.Interest {
	var ret []*common.Interest
	for interestId, _ := range is.idToInterests[id] {
		ret = append(ret, &common.Interest{id, is.interestMapper.strings[interestId]})
	}
	return ret
}

func (is *InterestStore) GetSuggestInterestIds(id int) map[int]int {
	mp := map[int]int{}
	for interestId, _ := range is.idToInterests[id] {
		for k, _ := range is.InterestsToId[interestId] {
			mp[k]++
		}
	}

	return mp
}
