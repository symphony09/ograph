package ogimpl

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var RaceClusterFactory = func() ogcore.Node {
	return &RaceCluster{}
}

type RaceCluster struct {
	ograph.BaseCluster
	*slog.Logger

	StateIsolation bool
}

func (cluster *RaceCluster) Run(ctx context.Context, state ogcore.State) error {
	if cluster.Logger == nil {
		cluster.Logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	raceCh := make(chan struct{})

	var failures atomic.Uint32
	var once sync.Once
	var winner string

	for _, node := range cluster.Group {
		go func(node ogcore.Node) {
			var clusterState ogcore.State

			if cluster.StateIsolation {
				clusterState = NewOverlayState(state)
			} else {
				clusterState = NewGuardState(state, func(key any) (flag int) {
					if ctx.Err() != nil {
						return AllowRead
					} else {
						return AllowRead | AllowWrite
					}
				})
			}

			nodeName := "unknown"

			if nameable, ok := node.(ogcore.Nameable); ok {
				nodeName = nameable.Name()
			}

			if err := node.Run(ctx, clusterState); err != nil {
				cluster.Warn("race node failed",
					"RaceCluster", cluster.Name(), "RaceNode", nodeName, "Error", err)

				if int(failures.Add(1)) == len(cluster.Group) {
					close(raceCh)
				}
			} else {
				once.Do(func() {
					winner = nodeName
					close(raceCh)
				})
			}
		}(node)
	}

	<-raceCh

	if int(failures.Load()) == len(cluster.Group) {
		return errors.New("all race nodes failed")
	}

	cluster.Info("race cluster finish", "Winner", winner)

	return nil
}
