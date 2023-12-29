package ogimpl

import (
	"context"
	"fmt"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
	"golang.org/x/sync/errgroup"
)

var ParallelClusterFactory = func() ogcore.Node {
	return &ParallelCluster{}
}

type ParallelCluster struct {
	ograph.BaseCluster
}

func (cluster *ParallelCluster) Run(ctx context.Context, state ogcore.State) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, node := range cluster.Group {
		node := node

		g.Go(func() error {
			err := node.Run(ctx, state)
			if err != nil {
				nodeName := "unknown"
				if nameable, ok := node.(ogcore.Nameable); ok {
					nodeName = nameable.Name()
				}

				return fmt.Errorf("sub node (%s) failed, err: %w", nodeName, err)
			} else {
				return nil
			}
		})
	}

	return g.Wait()
}
