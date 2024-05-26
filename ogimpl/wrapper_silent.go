package ogimpl

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var SilentWrapperFactory = NewSilentWrapper

type SilentWrapper struct {
	ograph.BaseWrapper
	*slog.Logger
}

func (wrapper *SilentWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Logger == nil {
		wrapper.Logger = slog.Default()
	}

	nodeName := "unknown"

	if nameable, ok := wrapper.Node.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	defer func() {
		if p := recover(); p != nil {
			wrapper.Warn("node panic", "NodeName", nodeName, "Panic", p, "Stack", string(debug.Stack()))
		}
	}()

	if err := wrapper.Node.Run(ctx, state); err != nil {
		wrapper.Warn("node failed", "NodeName", nodeName, "Error", err)
	}

	return nil
}

func NewSilentWrapper() ogcore.Node {
	return &SilentWrapper{}
}
