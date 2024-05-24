package ogimpl

import (
	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

type StateHookFn func(state ogcore.State, event string, key any, completed bool)

type HookState struct {
	Base ogcore.State

	Hooks []StateHookFn
}

func NewHookState(base ogcore.State, hooks ...StateHookFn) *HookState {
	if base == nil {
		base = ograph.NewState()
	}

	return &HookState{Base: base, Hooks: hooks}
}

func (state *HookState) Get(key any) (any, bool) {
	for _, hook := range state.Hooks {
		hook(state.Base, "get", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "get", key, true)
		}
	}()

	return state.Base.Get(key)
}

func (state *HookState) Set(key any, val any) {
	for _, hook := range state.Hooks {
		hook(state.Base, "set", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "set", key, true)
		}
	}()

	state.Base.Set(key, val)
}

func (state *HookState) Update(key any, updateFunc func(val any) any) {
	for _, hook := range state.Hooks {
		hook(state.Base, "update", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "update", key, true)
		}
	}()

	state.Base.Update(key, updateFunc)
}
