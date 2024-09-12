package example

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/ograph"
)

func TestInterrupter(t *testing.T) {
	pipeline := ograph.NewPipeline()

	begin := ograph.NewElement("Begin").AsVirtual()
	end := ograph.NewElement("End").AsVirtual()

	zhangShan := ograph.NewElement("ZhangSan").UseNode(&Sloth{})
	flash := ograph.NewElement("Flash").UseNode(&Sloth{})

	pipeline.Register(begin).
		Register(zhangShan, ograph.Rely(begin)).
		Register(flash, ograph.Rely(begin)).
		Register(end, ograph.Rely(zhangShan, flash))

	pipeline.ParallelismLimit = 1

	pipeline.Interrupts = func(yield func(string) bool) {
		start := time.Now()
		yield("Begin:start")
		yield("End:end")
		fmt.Printf("[TimeCounter] Total time cost: %s\n", time.Since(start))
	}

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
