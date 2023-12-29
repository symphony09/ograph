package ogimpl

import (
	"context"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var ConditionWrapperFactory = func() ogcore.Node {
	return &ConditionWrapper{}
}

type ConditionWrapper struct {
	ograph.BaseWrapper

	Switch string

	Condition func(ctx context.Context, state ogcore.State) bool
}

func (wrapper *ConditionWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Switch != "" {
		if wrapper.Switch == "On" {
			return wrapper.Node.Run(ctx, state)
		} else {
			return nil
		}
	} else if wrapper.Condition != nil {
		if ok := wrapper.Condition(ctx, state); ok {
			return wrapper.Node.Run(ctx, state)
		} else {
			return nil
		}
	}

	return nil
}

func NewConditionWrapperFactory(
	cond func(ctx context.Context, state ogcore.State) bool,
) func() ogcore.Node {

	return func() ogcore.Node {
		return &ConditionWrapper{
			Condition: cond,
		}
	}
}
