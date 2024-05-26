package ogimpl

import (
	"context"
	"log/slog"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var RetryWrapperFactory = func() ogcore.Node {
	return &RetryWrapper{MaxRetryTimes: 1}
}

type RetryWrapper struct {
	ograph.BaseWrapper
	*slog.Logger

	MaxRetryTimes int
}

func (wrapper *RetryWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Logger == nil {
		wrapper.Logger = slog.Default()
	}

	if err := wrapper.Node.Run(ctx, state); err != nil {
		if wrapper.MaxRetryTimes <= 0 {
			wrapper.MaxRetryTimes = 1
		}

		nodeName := "unknown"

		if nameable, ok := wrapper.Node.(ogcore.Nameable); ok {
			nodeName = nameable.Name()
		}

		for i := wrapper.MaxRetryTimes; i > 0; i-- {
			wrapper.Warn("retry failed node", "NodeName", nodeName, "Error", err)

			if err := wrapper.Node.Run(ctx, state); err != nil {
				if i == 1 {
					return err
				}
			} else {
				return nil
			}
		}
	}

	return nil
}

func NewRetryWrapper(times int) ogcore.Node {
	return &RetryWrapper{MaxRetryTimes: times}
}
