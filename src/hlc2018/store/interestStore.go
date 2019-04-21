package store

import "hlc2018/common"

type InterestStore struct {
	interestToInt map[string]int
	intToInterest []string
	idToInterests []map[int]struct{}
	InterestsToId []map[int]struct{}
}

func NewInterestStore() *InterestStore {
	is := &InterestStore{map[string]int{}, nil, nil, nil}
	is.insertInterestIfNeeded("")
	return is
}

func (is *InterestStore) ExtendSizeIfNeeded(nextSize int) {
	for len(is.idToInterests) < nextSize {
		is.idToInterests = append(is.idToInterests, map[int]struct{}{})
	}
}

func (is *InterestStore) insertInterestIfNeeded(s string) int {
	val, ok := is.interestToInt[s]
	if ok {
		return val
	}
	newVal := len(is.interestToInt)
	is.interestToInt[s] = newVal
	is.intToInterest = append(is.intToInterest, s)
	is.InterestsToId = append(is.InterestsToId, map[int]struct{}{})
	return newVal
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

func mapIntersect(l, r map[int]struct{}) map[int]struct{} {
	if len(l) > len(r) {
		l, r = r, l
	}
	ret := map[int]struct{}{}
	for k, _ := range l {
		if _, ok := r[k]; ok {
			ret[k] = struct{}{}
		}
	}
	return ret
}

func (is *InterestStore) ContainsAllFromInterests(vs []string) map[int]struct{} {
	var mp map[int]struct{}
	for _, s := range vs {
		interestId, found := is.interestToInt[s]
		if !found {
			return map[int]struct{}{}
		}
		if mp == nil {
			mp = is.InterestsToId[interestId]
		} else {
			mp = mapIntersect(mp, is.InterestsToId[interestId])
		}
	}
	return mp
}

func (is *InterestStore) ContainsAnyFromInterests(vs []string) map[int]struct{} {
	mp := map[int]struct{}{}
	for _, s := range vs {
		interestId, found := is.interestToInt[s]
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
		interestId, found := is.interestToInt[s]
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
		interestId, found := is.interestToInt[s]
		if !found {
			continue
		}
		if _, ok := is.idToInterests[id][interestId]; !ok {
			return true
		}
	}
	return false
}

func (is *InterestStore) GetCommonInterests(id int) []*common.Interest {
	var ret []*common.Interest
	for interestId, _ := range is.idToInterests[id] {
		ret = append(ret, &common.Interest{id, is.intToInterest[interestId]})
	}
	return ret
}
