package example

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
	"github.com/symphony09/ograph/ogimpl"
)

func TestCluster_Parallel(t *testing.T) {
	pipeline := ograph.NewPipeline()

	var counters []*ograph.Element

	for i := 0; i < 3; i++ {
		counters = append(counters, ograph.NewElement("").UseFn(func() error {
			for i := 0; i < 10; i++ {
				fmt.Print(i)
				time.Sleep(time.Millisecond)
			}
			return nil
		}))
	}

	parallelCounters := ograph.NewElement("ParallelCounters").
		UseFactory(ogimpl.Parallel, counters...)

	pipeline.Register(parallelCounters)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

type Player struct {
	ograph.BaseNode

	Cost  time.Duration
	Delay time.Duration
}

func (player *Player) Run(ctx context.Context, state ogcore.State) error {
	time.Sleep(player.Cost + player.Delay)
	state.Set("Winner", player.Name())
	return nil
}

func TestCluster_Race(t *testing.T) {
	pipeline := ograph.NewPipeline()

	turtle := ograph.NewElement("Turtle").UseNode(&Player{Cost: 5 * time.Millisecond})
	rabbit := ograph.NewElement("Rabbit").UseNode(&Player{Cost: 1 * time.Millisecond, Delay: 50 * time.Millisecond})

	race := ograph.NewElement("Race").UseFactory(ogimpl.Race, turtle, rabbit).Wrap(ogimpl.Trace)

	pipeline.Register(race)

	state := ograph.NewState()

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else {
		winner, _ := state.Get("Winner")
		fmt.Println("Winner is", winner)
	}
}

type CustomCluster struct {
	ograph.BaseCluster
}

func (cluster *CustomCluster) Run(ctx context.Context, state ogcore.State) error {
	for i := len(cluster.Group) - 1; i >= 0; i-- {
		if err := cluster.Group[i].Run(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

func NewCustomCluster() ogcore.Node {
	return &CustomCluster{}
}

func TestCluster_Customize(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.RegisterFactory("ReverseOrder", NewCustomCluster)

	first := ograph.NewElement("first").UseFn(func() error {
		fmt.Println("first")
		return nil
	})

	second := ograph.NewElement("second").UseFn(func() error {
		fmt.Println("second")
		return nil
	})

	myCluster := ograph.NewElement("MyCluster").
		UseFactory("ReverseOrder", first, second)

	pipeline.Register(myCluster)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
