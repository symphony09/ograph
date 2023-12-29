package ogcore

import (
	"maps"
	"sync"
)

type Factories struct {
	newNodeFuncs map[string]func() Node

	sync.RWMutex
}

func (f *Factories) Add(name string, newNode func() Node) {
	f.Lock()
	defer f.Unlock()

	f.newNodeFuncs[name] = newNode
}

func (f *Factories) Remove(name string) {
	f.Lock()
	defer f.Unlock()

	delete(f.newNodeFuncs, name)
}

func (f *Factories) Clear() {
	f.Lock()
	defer f.Unlock()

	clear(f.newNodeFuncs)
}

func (f *Factories) Get(name string) func() Node {
	f.RLock()
	defer f.RUnlock()

	return f.newNodeFuncs[name]
}

func (f *Factories) Clone() *Factories {
	f.RLock()
	defer f.RUnlock()

	return &Factories{
		newNodeFuncs: maps.Clone(f.newNodeFuncs),
	}
}

func NewFactories() *Factories {
	return &Factories{newNodeFuncs: make(map[string]func() Node)}
}
