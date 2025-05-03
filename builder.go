package ograph

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/symphony09/eventd"
	"github.com/symphony09/ograph/global"
	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Builder struct {
	Factories *ogcore.Factories
}

func (builder *Builder) RegisterPrototype(name string, prototype ogcore.Cloneable) *Builder {
	factory := func() ogcore.Node {
		return prototype.Clone()
	}

	builder.RegisterFactory(name, factory)
	return builder
}

func (builder *Builder) RegisterFactory(name string, factory func() ogcore.Node) *Builder {
	if builder.Factories == nil {
		builder.Factories = global.Factories.Clone()
	}

	builder.Factories.Add(name, factory)
	return builder
}

func (builder *Builder) build(graph *PGraph, eventBus *eventd.EventBus[ogcore.State]) (*internal.Worker, error) {
	if builder.Factories == nil {
		builder.Factories = global.Factories.Clone()
	}

	txManager := internal.NewTransactionManager()

	workGraph, err := internal.MapToNewGraph(graph, func(e *Element) (ogcore.Node, error) {
		return builder.doBuild(e, txManager, eventBus)
	})

	if err != nil {
		return nil, err
	}

	workGraph.Optimize()

	worker := internal.NewWorker(workGraph)
	worker.SetTxManager(txManager)

	return worker, nil
}

func (builder *Builder) doBuild(element *Element, txManager *internal.TransactionManager, eventBus *eventd.EventBus[ogcore.State]) (ogcore.Node, error) {
	if element.Virtual {
		return nil, nil
	}

	var node ogcore.Node

	factory := element.PrivateFactory
	if factory == nil {
		factory = builder.Factories.Get(element.FactoryName)
	}

	if element.Singleton != nil {
		node = element.Singleton
	} else if factory != nil {
		node = factory()

		if err := builder.doInit(node, element.ParamsMap); err != nil {
			return nil, fmt.Errorf("can't init node %s, err: %v", element.Name, err)
		}

		if len(element.SubElements) > 0 {
			if cluster, ok := node.(ogcore.Cluster); ok {
				subNodes := make([]ogcore.Node, 0, len(element.SubElements))

				for _, subElem := range element.SubElements {
					if subElem.Virtual {
						continue
					}

					if subNode, err := builder.doBuild(subElem, txManager, eventBus); err != nil {
						return nil, err
					} else {
						subNodes = append(subNodes, subNode)
					}
				}

				cluster.Join(subNodes)
			}
		}
	} else {
		return nil, fmt.Errorf("can't build node %s, factory of %s not found", element.Name, element.FactoryName)
	}

	if nameable, ok := node.(ogcore.Nameable); ok {
		nameable.SetName(element.Name)
	}

	if eventNode, ok := node.(ogcore.EventNode); ok {
		eventNode.AttachBus(eventBus)
	}

	if txNode, ok := node.(ogcore.Transactional); ok {
		node = txManager.Manage(txNode)
	}

	if pipeline, ok := node.(*Pipeline); ok {
		pipeline.Subscribe(func(event string, obj ogcore.State) bool {
			eventBus.Emit(event, obj)
			return true
		}, eventd.On(".*"))
	}

	seenWrapper := make(map[string]bool)

	if len(element.Wrappers) > 0 {
		for _, wrapperName := range element.Wrappers {
			// wrappers should be unique, avoid parameter confusion
			if seenWrapper[wrapperName] {
				continue
			} else {
				seenWrapper[wrapperName] = true
			}

			wrapperFactoryName := wrapperName

			if element.WrapperAlias != nil {
				if name := element.WrapperAlias[wrapperName]; name != "" {
					wrapperFactoryName = name
				}
			}

			var wrapperNode ogcore.Node

			if factory := builder.Factories.Get(wrapperFactoryName); factory != nil {
				wrapperNode = factory()
			} else {
				return nil, fmt.Errorf("can't build wrapper for %s, factory of %s not found", element.Name, wrapperFactoryName)
			}

			if nameable, ok := wrapperNode.(ogcore.Nameable); ok {
				nameable.SetName(element.Name)
			}

			if err := builder.doInit(wrapperNode, element.filterParams(wrapperName)); err != nil {
				return nil, fmt.Errorf("can't init wrapper %s, err: %v", element.Name, err)
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
		if len(params) == 0 {
			return nil
		}

		decoderConfig := &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToIPHookFunc(),
				mapstructure.StringToIPNetHookFunc(),
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			),
			Result: &node,
		}

		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			return err
		}

		if params != nil {
			if err := decoder.Decode(params); err != nil {
				return err
			}
		}
	}

	return nil
}
