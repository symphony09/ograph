package ogimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

var ConditionWrapperFactory = func() ogcore.Node {
	return &ConditionWrapper{}
}

type ConditionWrapper struct {
	ograph.BaseWrapper

	ConditionExpr string

	Condition func(ctx context.Context, state ogcore.State) bool
}

func (wrapper *ConditionWrapper) Init(params map[string]any) error {
	if fn, ok := params["Condition"].(func(ctx context.Context, state ogcore.State) bool); ok {
		wrapper.Condition = fn
		return nil
	}

	if exprStr, ok := params["ConditionExpr"].(string); ok {
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

	return errors.New("condition not set")
}

func (wrapper *ConditionWrapper) Run(ctx context.Context, state ogcore.State) error {
	if wrapper.Condition != nil {
		if ok := wrapper.Condition(ctx, state); ok {
			return wrapper.Node.Run(ctx, state)
		} else {
			return nil
		}
	}

	return nil
}

type Visitor struct {
	Identifiers []string
}

func (v *Visitor) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.Identifiers = append(v.Identifiers, n.Value)
	}
}

func NewConditionWrapperFactory(
	cond func(ctx context.Context, state ogcore.State) bool,
) func() ogcore.Node {

	return func() ogcore.Node {
		return &ConditionWrapper{
			Condition: cond,
		}
	}
}
