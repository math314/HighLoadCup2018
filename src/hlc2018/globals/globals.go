package globals

import (
	"hlc2018/store"
)

var (
	As = store.NewAccountStore()
	Ls = store.NewLikeStore(As)
	Is = store.NewInterestStore()
)
