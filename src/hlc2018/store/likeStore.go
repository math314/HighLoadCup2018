package store

import (
	"hlc2018/common"
	"math"
	"sort"
)

type storedLike struct {
	to, ts int
}

type storedLikeFloat64 struct {
	to int
	ts float64
}

type LikeStore struct {
	forward, backward [][]storedLike
	forwardMap        []map[int]struct{}
}

func NewLikeStore() *LikeStore {
	return &LikeStore{}
}

func (ls *LikeStore) ExtendSizeIfNeeded(nextSize int) {
	for len(ls.forward) < nextSize {
		ls.forward = append(ls.forward, []storedLike{})
		ls.backward = append(ls.backward, []storedLike{})
		ls.forwardMap = append(ls.forwardMap, map[int]struct{}{})
	}
}

func (ls *LikeStore) InsertLike(from, to, ts int) {
	if from < to {
		ls.ExtendSizeIfNeeded(to + 1)
	} else {
		ls.ExtendSizeIfNeeded(from + 1)
	}

	ls.forward[from] = append(ls.forward[from], storedLike{to, ts})
	ls.forwardMap[from][to] = struct{}{}
	ls.backward[to] = append(ls.backward[to], storedLike{from, ts})
}

func (ls *LikeStore) InsertCommonLike(like *common.Like) {
	ls.InsertLike(like.AccountIdFrom, like.AccountIdTo, like.Ts)
}

func (ls *LikeStore) CheckContainAllLikes(id int, liked []int) bool {
	for _, l := range liked {
		if _, ok := ls.forwardMap[id][l]; !ok {
			return false
		}
	}
	return true
}

func (ls *LikeStore) IdsContainAllLikes(ids []int) map[int]struct{} {
	minId := -1
	minVal := len(ls.forward)

	for _, id := range ids {
		if len(ls.backward) <= id {
			return map[int]struct{}{}
		}
		if minVal > len(ls.backward[id]) {
			minId = id
			minVal = len(ls.backward[id])
		}
	}

	ret := make(map[int]struct{}, minVal)
	for _, k := range ls.backward[minId] {
		ok := true
		for _, id := range ids {
			if _, contains := ls.forwardMap[k.to][id]; !contains {
				ok = false
				break
			}
		}
		if ok {
			ret[k.to] = struct{}{}
		}
	}

	return ret
}

func (ls *LikeStore) IdsContainAnyLikes(ids []int) []int {
	mp := map[int]struct{}{}
	for _, id := range ids {
		if len(ls.backward) <= id {
			continue
		}
		for _, k := range ls.backward[id] {
			mp[k.to] = struct{}{}
		}
	}

	var ret []int
	for k, _ := range mp {
		ret = append(ret, k)
	}

	return ret
}

func composed(mp []storedLike) []storedLikeFloat64 {
	ret := map[int][]int{}
	for _, sl := range mp {
		ret[sl.to] = append(ret[sl.to], sl.ts)
	}
	var ret2 []storedLikeFloat64
	for id, vals := range ret {
		sum := 0
		for _, x := range vals {
			sum += x
		}
		ret2 = append(ret2, storedLikeFloat64{id, float64(sum) / float64(len(vals))})
	}
	return ret2
}

func (ls *LikeStore) OrderByLikeSimilarity(id int) []int {
	mp := map[int]float64{}

	for _, sl := range composed(ls.forward[id]) {
		for _, otherSl := range composed(ls.backward[sl.to]) {
			if otherSl.to == id {
				continue
			}
			div := float64(sl.ts - otherSl.ts)
			if div == 0 {
				div = 1
			}
			add := 1 / math.Abs(div)
			mp[otherSl.to] += add
		}
	}

	type P struct {
		id  int
		val float64
	}
	var vp []P

	for k, v := range mp {
		vp = append(vp, P{k, v})
	}
	sort.Slice(vp, func(i, j int) bool {
		return vp[i].val > vp[j].val
	})

	var ret []int
	for _, v := range vp {
		ret = append(ret, v.id)
	}

	return ret
}

func (ls *LikeStore) GetNotLiked(id, othersId int, mp *map[int]struct{}, ret *[]int, limit int) {
	var vp []int
	for _, sl := range ls.forward[othersId] {
		if _, alreadyLiked := ls.forwardMap[id][sl.to]; alreadyLiked {
			continue
		}
		vp = append(vp, sl.to)
	}

	sort.Slice(vp, func(i, j int) bool {
		return vp[i] > vp[j]
	})

	for _, v := range vp {
		if _, found := (*mp)[v]; found {
			continue
		}
		(*mp)[v] = struct{}{}
		*ret = append(*ret, v)
		if len(*ret) == limit {
			return
		}
	}
}
