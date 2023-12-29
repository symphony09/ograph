package ogcore

import (
	"context"

	"github.com/symphony09/ograph/coro"
)

type Interrupter struct {
	Handler InterruptHandler
	Points  []string
}

type InterruptCtx struct {
	context.Context

	From  string
	State State
}

type InterruptHandler func(in InterruptCtx, yield func(error) InterruptCtx) error

type Interrupt func(in InterruptCtx) (error, bool)

func NewInterrupt(handler InterruptHandler) Interrupt {
	return coro.New(handler)
}

func NewInterruptCtx(ctx context.Context, from string, state State) InterruptCtx {
	return InterruptCtx{
		Context: ctx,
		From:    from,
		State:   state,
	}
}
