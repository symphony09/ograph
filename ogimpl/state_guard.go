package ogimpl

import (
	"github.com/symphony09/ograph/ogcore"
)

const (
	AllowRead = 1 << iota
	AllowWrite
)

type GuardState struct {
	guard  func(key any) int
	target ogcore.State
}

func (state *GuardState) Get(key any) (any, bool) {
	flag := state.guard(key)

	if flag&AllowRead == 0 {
		return nil, false
	} else {
		return state.target.Get(key)
	}
}

func (state *GuardState) Set(key any, val any) {
	flag := state.guard(key)

	if flag&AllowWrite == 0 {
		return
	} else {
		state.target.Set(key, val)
	}
}

func (state *GuardState) Update(key any, updateFunc func(val any) any) {
	flag := state.guard(key)

	if flag&AllowWrite == 0 || flag&AllowRead == 0 {
		return
	} else {
		state.target.Update(key, updateFunc)
	}
}

func NewGuardState(state ogcore.State, guard func(key any) (flag int)) *GuardState {
	if guard == nil {
		guard = func(key any) (flag int) {
			return AllowRead | AllowWrite
		}
	}

	return &GuardState{
		target: state,
		guard:  guard,
	}
}
