package ograph

import (
	"fmt"
	"reflect"

	"github.com/symphony09/ograph/ogcore"
)

func LoadPrivateState[SK ~string, SV any](state ogcore.State, key string) SV {
	var zeroVal SV

	if v, ok := state.Get(SK(key)); !ok {
		return zeroVal
	} else if ret, ok := v.(SV); !ok {
		return zeroVal
	} else {
		return ret
	}
}

func LoadState[SV any](state ogcore.State, key string) SV {
	return LoadPrivateState[string, SV](state, key)
}

func SavePrivateState[SK ~string](state ogcore.State, key string, val any, overwrite bool) {
	state.Update(SK(key), func(oldVal any) any {
		if oldVal != nil && !overwrite {
			return oldVal
		} else {
			return val
		}
	})
}

func SaveState(state ogcore.State, key string, val any, overwrite bool) {
	SavePrivateState[string](state, key, val, overwrite)
}

func UpdatePrivateState[SK ~string, SV any](state ogcore.State, key string, updateFunc func(oldVal SV) (val SV)) error {
	var err error

	state.Update(SK(key), func(val any) any {
		if oldVal, ok := val.(SV); val != nil && !ok {
			err = fmt.Errorf("update state value failed, unexpected type:%v", reflect.TypeOf(val))
			return oldVal
		} else {
			newVal := updateFunc(oldVal)
			return newVal
		}
	})

	return err
}

func UpdateState[SV any](state ogcore.State, key string, updateFunc func(oldVal SV) (val SV)) error {
	return UpdatePrivateState[string, SV](state, key, updateFunc)
}
