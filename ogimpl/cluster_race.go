package ogimpl

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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

	RaceStateKeys string
}

func (cluster *RaceCluster) Run(ctx context.Context, state ogcore.State) error {
	if cluster.Logger == nil {
		cluster.Logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	raceCh := make(chan struct{})
	raceKV := make(map[string]any)

	var failures atomic.Uint32
	var once sync.Once
	var winnerState ogcore.State

	var raceStateKeys []string
	if cluster.RaceStateKeys != "" {
		raceStateKeys = strings.Split(cluster.RaceStateKeys, ",")
	}

	for _, k := range raceStateKeys {
		raceKV[k], _ = state.Get(k)

		go func(key string) {
			state.Update(key, func(val any) any {
				<-raceCh

				if winnerState != nil {
					if newVal, ok := winnerState.Get(key); ok {
						return newVal
					}
				}

				return val
			})
		}(k)
	}

	for _, node := range cluster.Group {
		go func(node ogcore.Node) {
			overlayState := NewOverlayState(state)
			for k, v := range raceKV {
				overlayState.Set(k, v)
			}

			if err := node.Run(ctx, overlayState); err != nil {
				nodeName := "unknown"

				if nameable, ok := node.(ogcore.Nameable); ok {
					nodeName = nameable.Name()
				}

				cluster.Warn("race node failed",
					"RaceCluster", cluster.Name(), "RaceNode", nodeName, "Error", err)

				if int(failures.Add(1)) == len(cluster.Group) {
					close(raceCh)
				}
			} else {
				once.Do(func() {
					winnerState = overlayState
					close(raceCh)
				})
			}
		}(node)
	}

	<-raceCh
	if winnerState == nil {
		return errors.New("all race nodes failed")
	}

	return nil
}
