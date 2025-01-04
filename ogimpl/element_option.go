package ogimpl

import (
	"time"

	"github.com/symphony09/ograph"
)

func DelayOp(wait time.Duration) ograph.ElementOption {
	return func(e *ograph.Element) {
		e.Wrap(Delay).Params("Wait", wait)
	}
}

func LoopOp(n int) ograph.ElementOption {
	return func(e *ograph.Element) {
		e.Wrap(Loop).Params("LoopTimes", n)
	}
}

func TimeoutOp(dur time.Duration) ograph.ElementOption {
	return func(e *ograph.Element) {
		e.Wrap(Timeout).Params("Timeout", dur)
	}
}

func RetryOp(n int) ograph.ElementOption {
	return func(e *ograph.Element) {
		e.Wrap(Retry).Params("MaxRetryTimes", n)
	}
}

func ConditionOp(expr string) ograph.ElementOption {
	return func(e *ograph.Element) {
		e.Wrap(Condition).Params("ConditionExpr", expr)
	}
}
