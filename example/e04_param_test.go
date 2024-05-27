package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

type Speaker struct {
	ograph.BaseNode

	Words string
}

func (speaker Speaker) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("%s say: %s\n", speaker.Name(), speaker.Words)
	return nil
}

func TestParam(t *testing.T) {
	pipeline := ograph.NewPipeline()

	// for auto set params, node must be a pointer to a map or struct.
	pipeline.Builder.RegisterFactory("Speaker", func() ogcore.Node { return &Speaker{} })

	zhangSan := ograph.NewElement("ZhangSan").UseFactory("Speaker").Params("Words", "Don't be angry.")
	liSi := ograph.NewElement("LiSi").UseFactory("Speaker").Params("Words", "Be happy!")

	pipeline.Register(zhangSan).Register(liSi)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
