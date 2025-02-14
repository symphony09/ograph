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

func TestWrapper_Condition(t *testing.T) {
	pipeline := ograph.NewPipeline()

	ran := false

	zhangSan := ograph.NewElement("ZhangSan").
		UseFn(func() error {
			ran = true
			return nil
		}).
		Apply(ogimpl.ConditionOp("k == 2"))

	pipeline.Register(zhangSan)

	state := ograph.NewState()
	state.Set("k", 1)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else if ran == true {
		t.Error(errors.New("node ran when k equals 1"))
		return
	}

	state.Set("k", 2)

	if err := pipeline.Run(context.TODO(), state); err != nil {
		t.Error(err)
	} else if ran != true {
		t.Error(errors.New("node not ran when k equals 2"))
	}
}

type Loser struct {
	ograph.BaseNode
}

func (loser *Loser) Run(ctx context.Context, state ogcore.State) error {
	failureCount := ograph.LoadState[int](state, "failureCount")

	if failureCount == 0 {
		fmt.Printf("%s failed.\n", loser.Name())
	} else if failureCount < 3 {
		fmt.Printf("%s failed again.\n", loser.Name())
	} else {
		fmt.Printf("%s succeed!!!.\n", loser.Name())
	}

	if failureCount < 3 {
		ograph.UpdateState[int](state, "failureCount", func(oldVal int) (val int) {
			val = oldVal + 1
			return
		})

		return errors.New("too difficult")
	} else {
		return nil
	}
}

func TestWrapper_Retry(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Loser{}).
		Wrap(ogimpl.Retry).Params("MaxRetryTimes", 99)

	pipeline.Register(zhangSan)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

type Sloth struct {
	ograph.BaseNode
}

func (sloth *Sloth) Run(ctx context.Context, state ogcore.State) error {
	time.Sleep(time.Millisecond * 50)
	if ctx.Err() == nil {
		fmt.Println("Hi, i am Flash")
	}
	return nil
}

func TestWrapper_Timeout(t *testing.T) {
	pipeline := ograph.NewPipeline()

	flash := ograph.NewElement("Flash").UseNode(&Sloth{}).
		Wrap(ogimpl.Timeout).Params("Timeout", "10ms")

	pipeline.Register(flash)

	start := time.Now()

	err := pipeline.Run(context.TODO(), nil)
	if errors.Is(err, ogimpl.ErrTimeout) {
		fmt.Printf("time cost: %s\n", time.Since(start))
	} else {
		t.Error(err)
	}
}

func TestWrapper_Compose(t *testing.T) {
	pipeline := ograph.NewPipeline()

	flash := ograph.NewElement("Flash").UseNode(&Sloth{}).
		Wrap(ogimpl.Timeout).Params("Timeout", "10ms").
		Wrap(ogimpl.Silent).
		Wrap(ogimpl.Trace)

	pipeline.Register(flash)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

type CustomWrapper struct {
	ograph.BaseWrapper
}

func (wrapper *CustomWrapper) Run(ctx context.Context, state ogcore.State) error {
	fmt.Println("Before node start")

	wrapper.Node.Run(ctx, state)

	fmt.Println("After node finish")

	return nil
}

func NewCustomWrapper() ogcore.Node {
	return &CustomWrapper{}
}

func TestWrapper_Customize(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.RegisterFactory("MyWrapper", NewCustomWrapper)

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Loser{}).
		Wrap("MyWrapper")

	pipeline.Register(zhangSan)

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
