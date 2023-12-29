package example

import (
	"context"
	"testing"

	"github.com/symphony09/ograph"
)

func TestVirtual(t *testing.T) {
	pipeline := ograph.NewPipeline()

	middle := ograph.NewElement("Middle").SetVirtual(true)

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
