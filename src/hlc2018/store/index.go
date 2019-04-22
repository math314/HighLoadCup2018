package store

type StringIdMapper struct {
	stringToInt map[string]int
	strings     []string
}

func (sim *StringIdMapper) insertStringIfNeeded(s string) (int, bool) {
	val, ok := sim.stringToInt[s]
	if ok {
		return val, false
	}
	newVal := len(sim.stringToInt)
	sim.stringToInt[s] = newVal
	sim.strings = append(sim.strings, s)

	return newVal, true
}

func (sim *StringIdMapper) get(s string) (int, bool) {
	val, ok := sim.stringToInt[s]
	return val, ok
}
