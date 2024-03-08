package ogimpl

type FastState struct {
	store map[any]any
}

func (state *FastState) Get(key any) (any, bool) {
	if val, ok := state.store[key]; !ok {
		return nil, false
	} else {
		return val, true
	}
}

func (state *FastState) Set(key any, val any) {
	state.store[key] = val
}

func (state *FastState) Update(key any, updateFunc func(val any) any) {
	oldVal := state.store[key]
	newVal := updateFunc(oldVal)
	state.store[key] = newVal
}

func NewFastState() *FastState {
	state := new(FastState)

	state.store = make(map[any]any)

	return state
}
