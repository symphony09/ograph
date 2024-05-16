package test

import (
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogimpl"
)

func TestGuardState(t *testing.T) {
	state := ograph.NewState()
	guardState := ogimpl.NewGuardState(state, func(key any) (flag int) {
		if key == "rw" {
			return ogimpl.AllowRead | ogimpl.AllowWrite
		}

		if key == "r" {
			return ogimpl.AllowRead
		}

		if key == "w" {
			return ogimpl.AllowWrite
		}

		return 0
	})

	state.Set("r", 0)
	guardState.Set("r", 1)

	if v, _ := guardState.Get("r"); v != 0 {
		t.Error(v)
	}

	if v, _ := state.Get("r"); v != 0 {
		t.Error(v)
	}

	state.Set("w", 0)
	guardState.Set("w", 2)

	if v, ok := guardState.Get("w"); ok {
		t.Error(v)
	}

	if v, _ := state.Get("w"); v != 2 {
		t.Error(v)
	}

	guardState.Update("rw", func(val any) any { return 3 })

	if v, _ := guardState.Get("rw"); v != 3 {
		t.Error(v)
	}

	if v, _ := state.Get("rw"); v != 3 {
		t.Error(v)
	}
}
