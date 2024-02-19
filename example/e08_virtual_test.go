package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

func TestVirtual(t *testing.T) {
	pipeline := ograph.NewPipeline()

	middle := ograph.NewElement("Middle").AsVirtual()

	zhangShan := ograph.NewElement("ZhangSan").UseNode(&Person{})
	liSi := ograph.NewElement("LiSi").UseNode(&Person{})
	WangWu := ograph.NewElement("WangWu").UseNode(&Person{})
	ZhaoLiu := ograph.NewElement("ZhaoLiu").UseNode(&Person{})

	// A->C, A->D, B->C, B->D => (A, B)->V->(C, D)
	pipeline.Register(middle,
		ograph.DependOn(zhangShan, liSi),
		ograph.Then(WangWu, ZhaoLiu))

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

func TestImplVirtual(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.Builder.RegisterFactory("Speaker", func() ogcore.Node { return &Speaker{} })

	vSpeaker := ograph.NewElement("V").AsVirtual()

	ograph.NewElement("ZhangSan").
		UseFactory("Speaker").Params("Words", "Dont't be angry.").
		Implement(vSpeaker, true)

	ograph.NewElement("LiSi").
		UseFactory("Speaker").Params("Words", "Be angry.").
		Implement(vSpeaker, false)

	pipeline.Register(vSpeaker)

	fmt.Print("[Default] ")

	if err := pipeline.Run(context.Background(), nil); err != nil { // use default
		t.Error(err)
	}

	fmt.Print("[Another] ")

	ctx := ogcore.NewOGCtx(context.Background(), ogcore.WithImplMap(
		map[string]string{"V": "LiSi"},
	))

	if err := pipeline.Run(ctx, nil); err != nil {
		t.Error(err)
	}
}
