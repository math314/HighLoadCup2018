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
	sim       *StringIdMapper
	idToValue []map[int]struct{}
	valueToId []map[int]struct{}
}

func NewStringIndex() *StringIndex {
	is := &StringIndex{newStringIdMapper(), nil, nil}
	is.insertIfNeeded("")
	return is
}

func (si *StringIndex) ExtendSizeIfNeeded(nextSize int) {
	for len(si.idToValue) < nextSize {
		si.idToValue = append(si.idToValue, map[int]struct{}{})
	}
}

func (si *StringIndex) insertIfNeeded(s string) int {
	val, added := si.sim.InsertStringIfNeeded(s)
	if added {
		si.valueToId = append(si.valueToId, map[int]struct{}{})
	}
	return val
}

func (si *StringIndex) SetValue(id int, s string) int {
	si.ExtendSizeIfNeeded(id + 1)
	insertedId := si.insertIfNeeded(s)
	si.idToValue[id][insertedId] = struct{}{}
	si.valueToId[insertedId][id] = struct{}{}
	return insertedId
}
