package ograph

import (
	"context"
	"maps"
	"strings"

	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Element struct {
	Virtual     bool `json:"Virtual,omitempty"`
	Name        string
	FactoryName string         `json:"FactoryName,omitempty"`
	Wrappers    []string       `json:"Wrappers,omitempty"`
	ParamsMap   map[string]any `json:"ParamsMap,omitempty"`
	DefaultImpl string         `json:"DefaultImpl,omitempty"`

	Singleton ogcore.Node `json:"-"`

	PrivateFactory func() ogcore.Node `json:"-"`

	SubElements  []*Element `json:"SubElements,omitempty"`
	ImplElements []*Element `json:"ImplElements,omitempty"`
}

func (e *Element) SetVirtual(isVirtual bool) *Element {
	e.Virtual = isVirtual
	return e
}

func (e *Element) AsVirtual() *Element {
	e.Virtual = true
	return e
}

func (e *Element) UseFactory(name string, subElements ...*Element) *Element {
	e.FactoryName = name
	e.SubElements = append(e.SubElements, subElements...)
	return e
}

func (e *Element) UsePrivateFactory(factory func() ogcore.Node, subElements ...*Element) *Element {
	e.PrivateFactory = factory
	e.SubElements = append(e.SubElements, subElements...)
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

func (e *Element) Implement(virtualElem *Element, isDefault bool) *Element {
	virtualElem.ImplElements = append(virtualElem.ImplElements, e)

	if isDefault {
		virtualElem.DefaultImpl = e.Name
	}

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

func (e *Element) GetRequiredFactories() map[string]bool {
	factoriesMap := make(map[string]bool)

	if e.FactoryName != "" {
		factoriesMap[e.FactoryName] = true
	}

	for _, subElem := range e.SubElements {
		maps.Copy(factoriesMap, subElem.GetRequiredFactories())
	}

	for _, implElem := range e.ImplElements {
		maps.Copy(factoriesMap, implElem.GetRequiredFactories())
	}

	return factoriesMap
}

type PGraph = internal.Graph[*Element]

func NewElement(name string) *Element {
	return &Element{Name: name}
}
