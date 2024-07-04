package internal

import (
	"context"
	"time"

	"github.com/symphony09/ograph/ogcore"
)

type Profiler struct {
	Data ProfileData
}

type ProfileData struct {
	NodeCostTime map[string]time.Duration
}

func (profiler *Profiler) ProxyRunNode(ctx context.Context, state ogcore.State, node ogcore.Node, name string) error {
	if profiler.Data.NodeCostTime == nil {
		profiler.Data.NodeCostTime = make(map[string]time.Duration)
	}

	start := time.Now()
	defer func() {
		profiler.Data.NodeCostTime[name] = time.Since(start)
	}()

	return node.Run(ctx, state)
}
