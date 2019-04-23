package globals

import (
	"hlc2018/store"
)

var (
	Ls = store.NewLikeStore()
	Is = store.NewInterestStore()
	As = store.NewAccountStore()
)
