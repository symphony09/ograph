package ogimpl

import (
	"context"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var LoopWrapperFactory = func() ogcore.Node {
	return &LoopWrapper{LoopTimes: 1}
}

type LoopWrapper struct {
	ograph.BaseWrapper

	LoopTimes int
}

func (wrapper *LoopWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.LoopTimes < 0 {
		wrapper.LoopTimes = 1
	}

	for i := 0; i < wrapper.LoopTimes; i++ {
		if err := wrapper.Node.Run(ctx, state); err != nil {
			return err
		}
	}

	return nil
}
