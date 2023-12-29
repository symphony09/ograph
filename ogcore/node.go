package ogcore

import "context"

type Node interface {
	Run(ctx context.Context, state State) error
}

type Cluster interface {
	Join(nodes []Node)
}

type Wrapper interface {
	Wrap(node Node)
}

type Nameable interface {
	Name() string

	SetName(name string)
}

type Cloneable interface {
	Node
	Clone() Cloneable
}

type Initializeable interface {
	Init(params map[string]any) error
}
