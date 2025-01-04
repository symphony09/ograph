package ogimpl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

var ChooseClusterFactory = func() ogcore.Node {
	return &ChooseCluster{}
}

type ChooseCluster struct {
	ograph.BaseCluster
	*slog.Logger

	ChooseExpr string

	ChooseFn func(ctx context.Context, state ogcore.State) int
}

func (cluster *ChooseCluster) Init(params map[string]any) error {
	if fn, ok := params["ChooseFn"].(func(ctx context.Context, state ogcore.State) int); ok {
		cluster.ChooseFn = fn
		return nil
	}

	if exprStr, ok := params["ChooseExpr"].(string); ok {
		cluster.ChooseExpr = exprStr

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

		cluster.ChooseFn = func(ctx context.Context, state ogcore.State) int {
			env := make(map[string]any)

			for _, identifier := range v.Identifiers {
				env[identifier], _ = state.Get(identifier)
			}

			output, err := expr.Run(program, env)
			if err != nil {
				panic(err)
			}

			if ret, ok := output.(int); ok {
				return ret
			}

			panic(fmt.Errorf("unknown result: %v", output))
		}

		return nil
	}

	return errors.New("choose expr or function not set")
}

func (cluster *ChooseCluster) Run(ctx context.Context, state ogcore.State) error {
	if cluster.Logger == nil {
		cluster.Logger = slog.Default()
	}

	var chosenNode ogcore.Node

	if cluster.ChooseFn != nil {
		n := cluster.ChooseFn(ctx, state)
		if n > 0 && n <= len(cluster.Group) {
			chosenNode = cluster.Group[n-1]
		}
	}

	if chosenNode == nil {
		return nil
	}

	nodeName := "unknown"
	if nameable, ok := chosenNode.(ogcore.Nameable); ok {
		nodeName = nameable.Name()
	}

	if err := chosenNode.Run(ctx, state); err != nil {
		return fmt.Errorf("chosen node (%s) failed, err: %w", nodeName, err)
	} else {
		cluster.Info("choose cluster finish", "Chosen", nodeName)
		return nil
	}
}

func NewChooseClusterFactory(
	chooseFn func(ctx context.Context, state ogcore.State) int,
) func() ogcore.Node {

	return func() ogcore.Node {
		return &ChooseCluster{
			ChooseFn: chooseFn,
		}
	}
}
