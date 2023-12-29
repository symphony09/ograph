package ograph

import (
	"context"
	"sync"

	"github.com/symphony09/ograph/ogcore"
)

type BaseNode struct {
	name string

	Action ogcore.Action
}

func (node BaseNode) Run(ctx context.Context, state ogcore.State) error {
	if node.Action != nil {
		return node.Action(ctx, state)
	}
	return nil
}

func (node *BaseNode) Name() string {
	return node.name
}

func (node *BaseNode) SetName(name string) {
	node.name = name
}

type FuncNode struct {
	BaseNode

	RunFunc func(ctx context.Context, state ogcore.State) error
}

func (node *FuncNode) Run(ctx context.Context, state ogcore.State) error {
	if node.RunFunc != nil {
		return node.RunFunc(ctx, state)
	} else {
		return nil
	}
}

func NewFuncNode(runFunc func(ctx context.Context, state ogcore.State) error) *FuncNode {
	return &FuncNode{
		RunFunc: runFunc,
	}
}

type BaseCluster struct {
	BaseNode

	Group   []ogcore.Node
	NodeMap map[string]ogcore.Node
}

func (cluster *BaseCluster) Join(nodes []ogcore.Node) {
	cluster.Group = append(cluster.Group, nodes...)

	if cluster.NodeMap == nil {
		cluster.NodeMap = make(map[string]ogcore.Node)
	}

	for _, node := range nodes {
		if nameable, ok := node.(ogcore.Nameable); ok {
			cluster.NodeMap[nameable.Name()] = node
		}
	}
}

func (cluster BaseCluster) Run(ctx context.Context, state ogcore.State) error {
	for _, node := range cluster.Group {
		if err := node.Run(ctx, state); err != nil {
			return err
		}
	}

	return nil
}

type BaseWrapper struct {
	BaseNode

	ogcore.Node
}

func (wrapper *BaseWrapper) Wrap(node ogcore.Node) {
	wrapper.Node = node
}

func (wrapper BaseWrapper) Run(ctx context.Context, state ogcore.State) error {
	return wrapper.Node.Run(ctx, state)
}

type BaseState struct {
	store map[any]any

	sync.RWMutex
}

func (state *BaseState) Get(key any) (any, bool) {
	state.RLock()
	defer state.RUnlock()

	if val, ok := state.store[key]; !ok {
		return nil, false
	} else {
		return val, true
	}
}

func (state *BaseState) Set(key any, val any) {
	state.Lock()
	defer state.Unlock()

	state.store[key] = val
}

func (state *BaseState) Update(key any, updateFunc func(val any) any) {
	state.Lock()
	defer state.Unlock()

	oldVal := state.store[key]
	newVal := updateFunc(oldVal)
	state.store[key] = newVal
}

func NewState() *BaseState {
	state := new(BaseState)

	state.store = make(map[any]any)

	return state
}
