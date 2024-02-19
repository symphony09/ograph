package ogcore

import (
	"context"
)

type ImplQueryKey string

type OGContext struct {
	context.Context

	VirtualImpl func(key string) string
	MatchedImpl map[string]string
}

func (ctx *OGContext) Value(key any) any {
	if ctx.VirtualImpl != nil {
		if k, ok := key.(ImplQueryKey); ok {
			val := ctx.VirtualImpl(string(k))

			if ctx.MatchedImpl == nil {
				ctx.MatchedImpl = make(map[string]string)
			}

			ctx.MatchedImpl[string(k)] = val

			return val
		}
	}

	if ctx.Context != nil {
		return ctx.Context.Value(key)
	}

	return nil
}

type CtxOp func(ctx *OGContext)

var WithImplMap = func(impls map[string]string) CtxOp {
	return func(ctx *OGContext) {
		ctx.VirtualImpl = func(key string) string {
			return impls[key]
		}
	}
}

func NewOGCtx(parent context.Context, ops ...CtxOp) *OGContext {
	ctx := &OGContext{
		Context: parent,
	}

	for _, op := range ops {
		op(ctx)
	}

	return ctx
}
