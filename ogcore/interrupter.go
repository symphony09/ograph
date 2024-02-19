package ogcore

import (
	"context"
	"strings"

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

func GenActionsByIntr(interrupters []Interrupter) (map[string]Action, map[string]Action) {
	actionsBeforeRun := make(map[string]Action)
	actionsAfterRun := make(map[string]Action)

	for _, interrupter := range interrupters {
		interrupt := NewInterrupt(interrupter.Handler)

		for _, point := range interrupter.Points {
			nodeName, at, _ := strings.Cut(point, ":")
			if at == "" {
				at = "before"
			}

			if at == "before" {
				if action := actionsBeforeRun[nodeName]; action == nil {
					actionsBeforeRun[nodeName] = func(ctx context.Context, state State) error {
						err, _ := interrupt(NewInterruptCtx(ctx, nodeName, state))
						return err
					}
				} else {
					actionsBeforeRun[nodeName] = func(ctx context.Context, state State) error {
						if err, _ := interrupt(NewInterruptCtx(ctx, nodeName, state)); err != nil {
							return err
						}

						return action(ctx, state)
					}
				}
			} else if at == "after" {
				if action := actionsAfterRun[nodeName]; action == nil {
					actionsAfterRun[nodeName] = func(ctx context.Context, state State) error {
						err, _ := interrupt(NewInterruptCtx(ctx, nodeName, state))
						return err
					}
				} else {
					actionsAfterRun[nodeName] = func(ctx context.Context, state State) error {
						if err, _ := interrupt(NewInterruptCtx(ctx, nodeName, state)); err != nil {
							return err
						}

						return action(ctx, state)
					}
				}
			}
		}
	}

	return actionsBeforeRun, actionsAfterRun
}
