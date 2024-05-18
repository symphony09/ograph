package ogcore

import (
	"maps"
	"sync"
)

type Factories struct {
	factoryMap map[string]func() Node

	sync.RWMutex
}

func (f *Factories) Add(name string, factory func() Node) {
	f.Lock()
	defer f.Unlock()

	f.factoryMap[name] = factory
}

func (f *Factories) Remove(name string) {
	f.Lock()
	defer f.Unlock()

	delete(f.factoryMap, name)
}

func (f *Factories) Clear() {
	f.Lock()
	defer f.Unlock()

	clear(f.factoryMap)
}

func (f *Factories) Get(name string) func() Node {
	f.RLock()
	defer f.RUnlock()

	return f.factoryMap[name]
}

func (f *Factories) Clone() *Factories {
	f.RLock()
	defer f.RUnlock()

	return &Factories{
		factoryMap: maps.Clone(f.factoryMap),
	}
}

func NewFactories() *Factories {
	return &Factories{factoryMap: make(map[string]func() Node)}
}
