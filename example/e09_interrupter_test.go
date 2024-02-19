package example

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

func TestInterrupter(t *testing.T) {
	pipeline := ograph.NewPipeline()

	begin := ograph.NewElement("Begin").AsVirtual()
	end := ograph.NewElement("End").AsVirtual()

	zhangShan := ograph.NewElement("ZhangSan").UseNode(&Sloth{})
	flash := ograph.NewElement("Flash").UseNode(&Sloth{})

	pipeline.Register(begin).
		Register(zhangShan, ograph.DependOn(begin)).
		Register(flash, ograph.DependOn(begin)).
		Register(end, ograph.DependOn(zhangShan, flash))

	pipeline.ParallelismLimit = 1

	handler := func(in ogcore.InterruptCtx, yield func(error) ogcore.InterruptCtx) error {
		start := time.Now()

		yield(nil)

		fmt.Printf("[TimeCounter] Total time cost: %s\n", time.Since(start))

		return nil
	}

	pipeline.RegisterInterrupt(handler, "Begin:before", "End:after")

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
