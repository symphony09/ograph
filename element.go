package ograph

import (
	"context"
	"strings"

	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Element struct {
	Virtual     bool
	Name        string
	FactoryName string
	Wrappers    []string
	ParamsMap   map[string]any

	Singleton ogcore.Node `json:"-"`

	SubElements []*Element
}

func (e *Element) SetVirtual(isVirtual bool) *Element {
	e.Virtual = isVirtual
	return e
}

func (e *Element) UseFactory(name string, subPNodes ...*Element) *Element {
	e.FactoryName = name
	e.SubElements = append(e.SubElements, subPNodes...)
	return e
}

func (e *Element) Wrap(wrappers ...string) *Element {
	e.Wrappers = append(e.Wrappers, wrappers...)
	return e
}

func (e *Element) UseNode(node ogcore.Node) *Element {
	e.Singleton = node
	return e
}

func (e *Element) UseFn(fn func() error) *Element {
	e.Singleton = &BaseNode{
		Action: func(ctx context.Context, state ogcore.State) error {
			return fn()
		}}

	return e
}

func (e *Element) Params(key string, val any) *Element {
	if e.ParamsMap == nil {
		e.ParamsMap = make(map[string]any)
	}

	e.ParamsMap[key] = val

	return e
}

func (e *Element) filterParams(owner string) map[string]any {
	if e.ParamsMap == nil {
		return nil
	}

	ret := make(map[string]any)

	for key, val := range e.ParamsMap {
		if belong, paramKey, ok := strings.Cut(key, "."); ok {
			if belong == owner {
				ret[paramKey] = val
			}
		} else {
			ret[key] = val
		}
	}

	return ret
}

type PGraph = internal.Graph[*Element]

func NewElement(name string) *Element {
	return &Element{Name: name}
}
