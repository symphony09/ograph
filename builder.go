package ograph

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/symphony09/ograph/global"
	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Builder struct {
	Factories *ogcore.Factories
}

func (builder *Builder) RegisterPrototype(name string, prototype ogcore.Cloneable) *Builder {
	newNode := func() ogcore.Node {
		return prototype.Clone()
	}

	builder.RegisterFactory(name, newNode)
	return builder
}

func (builder *Builder) RegisterFactory(name string, newNode func() ogcore.Node) *Builder {
	if builder.Factories == nil {
		builder.Factories = global.Factories.Clone()
	}

	builder.Factories.Add(name, newNode)
	return builder
}

func (builder *Builder) build(graph *PGraph) (*internal.Worker, error) {
	if builder.Factories == nil {
		builder.Factories = global.Factories.Clone()
	}

	workGraph, err := internal.MapToNewGraph[*Element, ogcore.Node](graph, builder.doBuild)
	if err != nil {
		return nil, err
	}

	workGraph.Optimize()

	worker := internal.NewWorker(workGraph)
	return worker, nil
}

func (builder *Builder) doBuild(element *Element) (ogcore.Node, error) {
	if element.Virtual {
		vn := &ogcore.VirtualNode{
			VirtualName: element.Name,
			DefaultImpl: element.DefaultImpl,
			Implements:  make(map[string]ogcore.Node),
		}

		for _, impl := range element.ImplElements {
			if node, err := builder.doBuild(impl); err != nil {
				return nil, err
			} else {
				vn.Implements[impl.Name] = node
			}
		}

		return vn, nil
	}

	var node ogcore.Node

	if element.Singleton != nil {
		node = element.Singleton
	} else if newNode := builder.Factories.Get(element.FactoryName); newNode != nil {
		node = newNode()

		if err := builder.doInit(node, element.ParamsMap); err != nil {
			return nil, fmt.Errorf("cann't init node %s, err: %v", element.Name, err)
		}

		if len(element.SubElements) > 0 {
			if cluster, ok := node.(ogcore.Cluster); ok {
				subNodes := make([]ogcore.Node, 0, len(element.SubElements))

				for _, subElem := range element.SubElements {
					if subElem.Virtual {
						continue
					}

					if subNode, err := builder.doBuild(subElem); err != nil {
						return nil, err
					} else {
						subNodes = append(subNodes, subNode)
					}
				}

				cluster.Join(subNodes)
			}
		}
	} else {
		return nil, fmt.Errorf("cann't build node %s, factory of %s not found", element.Name, element.FactoryName)
	}

	if nameable, ok := node.(ogcore.Nameable); ok {
		nameable.SetName(element.Name)
	}

	if len(element.Wrappers) > 0 {
		for _, wrapperFactoryName := range element.Wrappers {
			var wrapperNode ogcore.Node

			if newNode := builder.Factories.Get(wrapperFactoryName); newNode != nil {
				wrapperNode = newNode()
			} else {
				return nil, fmt.Errorf("cann't build wrapper for %s, factory of %s not found", element.Name, wrapperFactoryName)
			}

			if nameable, ok := wrapperNode.(ogcore.Nameable); ok {
				nameable.SetName(element.Name)
			}

			if err := builder.doInit(wrapperNode, element.filterParams(wrapperFactoryName)); err != nil {
				return nil, fmt.Errorf("cann't init wrapper %s, err: %v", element.Name, err)
			}

			if wrapper, ok := wrapperNode.(ogcore.Wrapper); ok {
				wrapper.Wrap(node)
				node = wrapperNode
			}
		}
	}

	return node, nil
}

func (builder *Builder) doInit(node any, params map[string]any) error {
	if initializeable, ok := node.(ogcore.Initializeable); ok {
		if err := initializeable.Init(params); err != nil {
			return err
		}
	} else {
		if params != nil {
			if err := mapstructure.Decode(params, &node); err != nil {
				return err
			}
		}
	}

	return nil
}
