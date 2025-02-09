package ogcore

import (
	"context"

	"github.com/symphony09/eventd"
)

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

type Transactional interface {
	Node
	Commit()
	Rollback()
}

type EventNode interface {
	Node
	AttachBus(bus *eventd.EventBus[State])
}
