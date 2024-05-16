package ogimpl

import (
	"context"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var DelayWrapperFactory = func() ogcore.Node {
	return &DelayWrapper{}
}

type DelayWrapper struct {
	ograph.BaseWrapper

	Wait  time.Duration
	Until time.Time
}

func (wrapper *DelayWrapper) Run(ctx context.Context, state ogcore.State) error {
	var timeToRun time.Time

	if t := time.Now().Add(wrapper.Wait); t.After(wrapper.Until) {
		timeToRun = t
	} else {
		timeToRun = wrapper.Until
	}

	time.Sleep(time.Until(timeToRun))

	return wrapper.Node.Run(ctx, state)
}
