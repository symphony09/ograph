package example

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

type ProductionX struct {
	Date time.Time
}

func (p ProductionX) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("produced at %s\n", p.Date)
	return nil
}

func NewProductionX() ogcore.Node {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	time.Sleep(time.Duration(r.Intn(10)) * time.Millisecond)
	return ProductionX{Date: time.Now()}
}

func TestFactory(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.RegisterFactory("X", NewProductionX)

	x1 := ograph.NewElement("x1").UseFactory("X")
	x2 := ograph.NewElement("x2").UseFactory("X")

	pipeline.Register(x1).Register(x2)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

type Fakes struct{}

func (p Fakes) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("produced at %s\n", time.Now().Add(time.Hour*24))
	return nil
}

func TestPrivateFactory(t *testing.T) {
	pipeline := ograph.NewPipeline()

	x1 := ograph.NewElement("x1").UsePrivateFactory(
		func() ogcore.Node {
			return &Fakes{}
		})

	pipeline.Register(x1)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
