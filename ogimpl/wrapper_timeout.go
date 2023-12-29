package ogimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var TimeoutWrapperFactory = func() ogcore.Node {
	return &TimeoutWrapper{}
}

var ErrTimeout = errors.New("the running time exceeds the limit")

type TimeoutWrapper struct {
	ograph.BaseWrapper

	Timeout string
}

func (wrapper *TimeoutWrapper) Run(ctx context.Context, state ogcore.State) error {
	duration, err := time.ParseDuration(wrapper.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout setting, error: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	guardState := NewGuardState(state, func(key any) (flag int) {
		if ctx.Err() != nil {
			return AllowRead
		} else {
			return AllowRead | AllowWrite
		}
	})

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		err := wrapper.Node.Run(ctx, guardState)
		errCh <- err
	}(ctx)

	timer := time.NewTimer(duration)

	select {
	case err := <-errCh:
		timer.Stop()
		return err
	case <-timer.C:
		return fmt.Errorf("node failed after %s, error: %w", duration, ErrTimeout)
	}
}

func NewTimeoutWrapper(duration time.Duration) ogcore.Node {
	return &TimeoutWrapper{Timeout: duration.String()}
}
