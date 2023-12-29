package ogimpl

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var TraceWrapperFactory = func() ogcore.Node {
	return &TraceWrapper{}
}

type traceIdKey string

type TraceWrapper struct {
	ograph.BaseWrapper
	*slog.Logger
}

func (wrapper *TraceWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Logger == nil {
		wrapper.Logger = slog.Default()
	}

	nodeName := "unknown"
	if nameable, ok := wrapper.Node.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	var idKey traceIdKey
	var traceId string
	if id, ok := state.Get(idKey); ok {
		traceId, _ = id.(string)
	}

	start := time.Now()

	defer func() {
		if p := recover(); p != nil {
			wrapper.Error("node panic", "NodeName", nodeName,
				"Panic", p, "Stack", string(debug.Stack()), "TraceID", traceId)
			panic(p)
		}

		wrapper.Info("node finish", "NodeName", nodeName,
			"TimeCost", time.Since(start), "TraceID", traceId)
	}()

	wrapper.Info("node start", "NodeName", nodeName, "TraceID", traceId)

	return wrapper.Node.Run(ctx, state)
}

func SetTraceId(state ogcore.State, id string) {
	var key traceIdKey
	state.Set(key, id)
}
