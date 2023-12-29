package example

import (
	"context"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

type Counter struct {
}

func (counter Counter) Run(ctx context.Context, state ogcore.State) error {
	ograph.UpdateState[int](state, "count", func(oldVal int) (val int) {
		val = oldVal + 1
		return
	})

	return nil
}

func TestState(t *testing.T) {
	pipeline := ograph.NewPipeline()

	a := ograph.NewElement("a").UseNode(&Counter{})
	b := ograph.NewElement("b").UseNode(&Counter{})

	pipeline.Register(a).Register(b)

	state := ograph.NewState()
	ograph.SaveState(state, "count", 1, true)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else {
		if count := ograph.LoadState[int](state, "count"); count != 3 {
			t.Errorf("unexcept count: %d", count)
		}
	}
}

func TestPrivateState(t *testing.T) {
	pipeline := ograph.NewPipeline()

	a := ograph.NewElement("a").UseNode(&Counter{})
	b := ograph.NewElement("b").UseNode(&Counter{})

	pipeline.Register(a).Register(b)

	type pk string // private key type

	state := ograph.NewState()
	ograph.SavePrivateState[pk](state, "count", 1, true)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else {
		if count := ograph.LoadPrivateState[pk, int](state, "count"); count != 1 {
			t.Errorf("unexcept count: %d", count)
		}

		if count := ograph.LoadState[int](state, "count"); count != 2 {
			t.Errorf("unexcept count: %d", count)
		}
	}
}
