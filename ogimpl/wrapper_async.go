package ogimpl

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var AsyncWrapperFactory = func() ogcore.Node {
	return &AsyncWrapper{}
}

type AsyncWrapper struct {
	ograph.BaseWrapper
	*slog.Logger
}

func (wrapper *AsyncWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Logger == nil {
		wrapper.Logger = slog.Default()
	}

	nodeName := "unknown"
	if nameable, ok := wrapper.Node.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	overState := NewOverlayState(state)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				wrapper.Error("node panic", "NodeName", nodeName, "Panic", p, "Stack", string(debug.Stack()))
			}
		}()

		if err := wrapper.Node.Run(ctx, overState); err != nil {
			wrapper.Error("node failed", "NodeName", nodeName, "Error", err)
		}
	}()

	return nil
}
