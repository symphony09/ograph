package ogimpl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var ChooseClusterFactory = func() ogcore.Node {
	return &ChooseCluster{}
}

type ChooseCluster struct {
	ograph.BaseCluster
	*slog.Logger

	ChooseNode string

	ChooseFn func(ctx context.Context, state ogcore.State) int
}

func (cluster *ChooseCluster) Run(ctx context.Context, state ogcore.State) error {
	if cluster.Logger == nil {
		cluster.Logger = slog.Default()
	}

	var chosenNode ogcore.Node

	if cluster.ChooseNode != "" {
		chosenNode = cluster.NodeMap[cluster.ChooseNode]
	} else if cluster.ChooseFn != nil {
		n := cluster.ChooseFn(ctx, state)
		if n > 0 || n <= len(cluster.Group) {
			chosenNode = cluster.Group[n-1]
		}
	}

	if chosenNode == nil {
		return nil
	}

	nodeName := "unknown"
	if nameable, ok := chosenNode.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	if err := chosenNode.Run(ctx, state); err != nil {
		return fmt.Errorf("chosen node (%s) failed, err: %w", nodeName, err)
	} else {
		cluster.Info("choose cluster finish", "Chosen", nodeName)
		return nil
	}
}

func NewChooseClusterFactory(
	chooseFn func(ctx context.Context, state ogcore.State) int,
) func() ogcore.Node {

	return func() ogcore.Node {
		return &ChooseCluster{
			ChooseFn: chooseFn,
		}
	}
}
