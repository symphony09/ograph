package ogimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var AssertNodeFactory = func() ogcore.Node {
	return &AssertNode{}
}

type AssertNode struct {
	ograph.BaseNode

	AssertExpr string

	AssertFn func(ctx context.Context, state ogcore.State) (bool, error)
}

func (node *AssertNode) Init(params map[string]any) error {
	if fn, ok := params["AssertFn"].(func(ctx context.Context, state ogcore.State) (bool, error)); ok {
		node.AssertFn = fn
		return nil
	}

	if exprStr, ok := params["AssertExpr"].(string); ok {
		node.AssertExpr = exprStr

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

		node.AssertFn = func(ctx context.Context, state ogcore.State) (bool, error) {
			env := make(map[string]any)

			for _, identifier := range v.Identifiers {
				env[identifier], _ = state.Get(identifier)
			}

			output, err := expr.Run(program, env)
			if err != nil {
				return false, err
			}

			if ret, ok := output.(bool); ok {
				return ret, nil
			}

			return false, fmt.Errorf("unknown assert result: %v", output)
		}

		return nil
	}

	return errors.New("assert expr or function not set")
}

func (node *AssertNode) Run(ctx context.Context, state ogcore.State) error {
	if node.AssertFn != nil {
		pass, err := node.AssertFn(ctx, state)
		if err != nil {
			return fmt.Errorf("unable to get assert result, error:%w", err)
		}

		if !pass {
			return errors.New("assert failed")
		}
	}

	return nil
}

func NewAssertNodeFactory(
	assertFn func(ctx context.Context, state ogcore.State) (bool, error),
) func() ogcore.Node {

	return func() ogcore.Node {
		return &AssertNode{
			AssertFn: assertFn,
		}
	}
}
