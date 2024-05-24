package ogimpl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var DebugWrapperFactory = func() ogcore.Node {
	return &DebugWrapper{}
}

type DebugWrapper struct {
	ograph.BaseWrapper
	*slog.Logger
}

func (wrapper *DebugWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Logger == nil {
		wrapper.Logger = slog.Default()
	}

	var idKey traceIdKey
	var traceId string
	if id, ok := state.Get(idKey); ok {
		traceId, _ = id.(string)
	}

	nodeName := "unknown"
	if nameable, ok := wrapper.Node.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	hookState := NewHookState(state, func(state ogcore.State, event string, key any, completed bool) {
		if (event == "get" && completed) || (event == "set" && !completed) {
			return
		}

		var at string
		if !completed {
			at = "before_" + event
		} else {
			at = "after_" + event
		}

		val, _ := state.Get(key)

		keyStr := fmt.Sprintf("(%T)%v", key, key)
		valStr := fmt.Sprintf("(%T)%v", val, val)

		wrapper.Info("state debug hook", "NodeName", nodeName,
			"At", at, "Key", keyStr, "Value", valStr, "TraceID", traceId)
	})

	return wrapper.Node.Run(ctx, hookState)
}
