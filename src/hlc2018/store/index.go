package store

type StringIdMapper struct {
	stringToInt map[string]int
	strings     []string
}

func newStringIdMapper() *StringIdMapper {
	ret := &StringIdMapper{map[string]int{}, nil}
	return ret
}

func (sim *StringIdMapper) InsertStringIfNeeded(s string) (int, bool) {
	val, ok := sim.stringToInt[s]
	if ok {
		return val, false
	}
	newVal := len(sim.stringToInt)
	sim.stringToInt[s] = newVal
	sim.strings = append(sim.strings, s)

	return newVal, true
}

func (sim *StringIdMapper) GetWithFound(s string) (int, bool) {
	val, ok := sim.stringToInt[s]
	return val, ok
}

func (sim *StringIdMapper) Get(s string) int {
	val, ok := sim.stringToInt[s]
	if ok {
		return val
	} else {
		return -1
	}
}

type StringIndex struct {
	sim          *StringIdMapper
	pkToStringId []map[int]struct{}
	stringIdToPk []map[int]struct{}
}

func NewStringIndex() *StringIndex {
	is := &StringIndex{newStringIdMapper(), nil, nil}
	is.insertIfNeeded("")
	return is
}

func (si *StringIndex) ExtendSizeIfNeeded(nextSize int) {
	for len(si.pkToStringId) < nextSize {
		si.pkToStringId = append(si.pkToStringId, map[int]struct{}{})
	}
}

func (si *StringIndex) insertIfNeeded(s string) int {
	val, added := si.sim.InsertStringIfNeeded(s)
	if added {
		si.stringIdToPk = append(si.stringIdToPk, map[int]struct{}{})
	}
	return val
}

func (si *StringIndex) SetString(pk int, s string) int {
	si.ExtendSizeIfNeeded(pk + 1)
	insertedId := si.insertIfNeeded(s)
	si.pkToStringId[pk][insertedId] = struct{}{}
	si.stringIdToPk[insertedId][pk] = struct{}{}
	return insertedId
}

func (si *StringIndex) DeleteStringsFromPk(pk int) {
	for currentSID, _ := range si.pkToStringId[pk] {
		delete(si.pkToStringId[pk], currentSID)
	}
	si.pkToStringId[pk] = map[int]struct{}{}
}

func (si *StringIndex) ConvertStringToStringId(s string) int {
	return si.sim.Get(s)
}

func (si *StringIndex) StringIdToString(sid int) string {
	return si.sim.strings[sid]
}
