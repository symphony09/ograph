package ogimpl

import (
	"sync"

	"github.com/symphony09/ograph/ogcore"
)

type OverlayState struct {
	Upper map[any]any
	Lower ogcore.State

	sync.RWMutex
}

func (state *OverlayState) Get(key any) (any, bool) {
	if val, ok := state.Upper[key]; ok {
		state.RLock()
		defer state.RUnlock()

		return val, ok
	}

	return state.Lower.Get(key)
}

func (state *OverlayState) Set(key any, val any) {
	state.Lock()
	defer state.Unlock()

	state.Upper[key] = val
}

func (state *OverlayState) Update(key any, updateFunc func(val any) any) {
	state.Lock()
	defer state.Unlock()

	if val, ok := state.Upper[key]; ok {
		state.Upper[key] = updateFunc(val)
	} else {
		val, _ := state.Lower.Get(key)
		state.Upper[key] = updateFunc(val)
	}
}

func (state *OverlayState) Sync() {
	for k, v := range state.Upper {
		state.Lower.Set(k, v)
		delete(state.Upper, k)
	}
}

func NewOverlayState(state ogcore.State) *OverlayState {
	overState := &OverlayState{
		Upper: make(map[any]any),
		Lower: state,
	}

	return overState
}
