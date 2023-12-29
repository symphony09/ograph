package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

type Person struct {
	ograph.BaseNode
}

func (person *Person) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("Hello, i am %s.\n", person.Name())
	return nil
}

func TestHello(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Person{})
	liSi := ograph.NewElement("LiSi").UseNode(&Person{})

	pipeline.Register(zhangSan).
		Register(liSi, ograph.DependOn(zhangSan))

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
