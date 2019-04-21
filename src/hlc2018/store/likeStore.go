package store

import "hlc2018/common"

type LikeStore struct {
	forward, backward []map[int]int
}

func New() *LikeStore {
	return &LikeStore{}
}

func (ls *LikeStore) ExtendSizeIfNeeded(nextSize int) {
	for len(ls.forward) < nextSize {
		ls.forward = append(ls.forward, map[int]int{})
		ls.backward = append(ls.backward, map[int]int{})
	}
}

func (ls *LikeStore) InsertLike(from, to, ts int) {
	if from < to {
		ls.ExtendSizeIfNeeded(to + 1)
	} else {
		ls.ExtendSizeIfNeeded(from + 1)
	}

	ls.forward[from][to] = ts
	ls.backward[to][from] = ts
}

func (ls *LikeStore) InsertCommonLike(like *common.Like) {
	ls.InsertLike(like.AccountIdFrom, like.AccountIdTo, like.Ts)
}

func (ls *LikeStore) DeleteLikes(id int) {
	for k, _ := range ls.forward[id] {
		delete(ls.backward[k], id)
	}
	ls.forward[id] = map[int]int{}
}

func (ls *LikeStore) IdsContainAllLikes(ids []int) []int {
	minId := -1
	minVal := len(ls.forward)

	for _, id := range ids {
		if minVal > len(ls.backward[id]) {
			minId = id
			minVal = len(ls.backward[id])
		}
	}

	ret := make([]int, 0, minVal)
	for k, _ := range ls.backward[minId] {
		ok := true
		for _, id := range ids {
			if _, contains := ls.forward[k][id]; !contains {
				ok = false
				break
			}
		}
		if ok {
			ret = append(ret, k)
		}
	}

	return ret
}

func (ls *LikeStore) IdsContainAnyLikes(ids []int) []int {
	mp := map[int]struct{}{}
	for _, id := range ids {
		for k, _ := range ls.backward[id] {
			mp[k] = struct{}{}
		}
	}

	var ret []int
	for k, _ := range mp {
		ret = append(ret, k)
	}

	return ret
}
