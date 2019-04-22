package store

type StoreSource interface {
	Next() bool
	Value() int
}

type RangeStoreSource struct {
	current, end, step int
}

func NewRangeStoreSource(current, end, step int) *RangeStoreSource {
	return &RangeStoreSource{current, end, step}
}

func (bss *RangeStoreSource) Next() bool {
	bss.current += bss.step
	return bss.current != bss.end
}

func (bss *RangeStoreSource) Value() int {
	return bss.current
}

type ArrayStoreSource struct {
	src []int
	id  int
}

func NewArrayStoreSource(src []int) *ArrayStoreSource {
	return &ArrayStoreSource{src, -1}
}

func (ss *ArrayStoreSource) Next() bool {
	ss.id++
	return ss.id < len(ss.src)
}

func (ss *ArrayStoreSource) Value() int {
	return ss.src[ss.id]
}
