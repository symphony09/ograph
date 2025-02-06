package ogimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var LoopWrapperFactory = func() ogcore.Node {
	return &LoopWrapper{LoopTimes: 1}
}

type LoopWrapper struct {
	ograph.BaseWrapper

	LoopTimes    int
	LoopInterval time.Duration

	ConditionExpr string
	Condition     func(ctx context.Context, state ogcore.State) bool
}

func (wrapper *LoopWrapper) Init(params map[string]any) error {
	if loopTimes, ok := params["LoopTimes"].(int); ok {
		wrapper.LoopTimes = loopTimes
	}

	if loopInterval, ok := params["LoopInterval"].(time.Duration); ok {
		wrapper.LoopInterval = loopInterval
	} else if loopIntervalStr, ok := params["LoopInterval"].(string); ok {
		loopInterval, err := time.ParseDuration(loopIntervalStr)
		if err != nil {
			return err
		}
		wrapper.LoopInterval = loopInterval
	}

	if fn, ok := params["Condition"].(func(ctx context.Context, state ogcore.State) bool); ok {
		wrapper.Condition = fn
	} else if exprStr, ok := params["ConditionExpr"].(string); ok {
		wrapper.ConditionExpr = exprStr

		program, err := expr.Compile(exprStr)
		if err != nil {
			return err
		}

		tree, err := parser.Parse(exprStr)
		if err != nil {
			return err
		}

		v := &Visitor{}
		ast.Walk(&tree.Node, v)

		wrapper.Condition = func(ctx context.Context, state ogcore.State) bool {
			env := make(map[string]any)

			for _, identifier := range v.Identifiers {
				env[identifier], _ = state.Get(identifier)
			}

			output, err := expr.Run(program, env)
			if err != nil {
				panic(err)
			}

			if ret, ok := output.(bool); ok {
				return ret
			}

			panic(fmt.Errorf("unknown result: %v", output))
		}

		return nil
	}

	return nil
}

func (wrapper *LoopWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Condition == nil {
		if wrapper.LoopTimes < 0 {
			wrapper.LoopTimes = 1
		}

		for i := 0; i < wrapper.LoopTimes; i++ {
			if err := wrapper.Node.Run(ctx, state); err != nil {
				return err
			}

			time.Sleep(wrapper.LoopInterval)
		}

		return nil
	} else {
		for wrapper.Condition(ctx, state) {
			if err := wrapper.Node.Run(ctx, state); err != nil {
				return err
			}

			time.Sleep(wrapper.LoopInterval)
		}

		return nil
	}
}
