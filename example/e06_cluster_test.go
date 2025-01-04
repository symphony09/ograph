package example

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
	"github.com/symphony09/ograph/ogimpl"
)

func TestCluster_Choose(t *testing.T) {
	pipeline := ograph.NewPipeline()

	var chosen string

	a := ograph.NewElement("A").UseFn(func() error {
		chosen = "A"
		return nil
	})
	b := ograph.NewElement("B").UseFn(func() error {
		chosen = "B"
		return nil
	})

	race := ograph.NewElement("Cond").UseFactory(ogimpl.Choose, a, b).Params("ChooseExpr", "index")

	pipeline.Register(race)

	state := ograph.NewState()
	state.Set("index", 1)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else if chosen != "A" {
		t.Error(errors.New("node A not ran when index equals 1"))
	}

	state.Set("index", 2)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else if chosen != "B" {
		t.Error(errors.New("node B not ran when index equals 2"))
	}
}

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
