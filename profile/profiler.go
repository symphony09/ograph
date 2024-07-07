package profile

import (
	"fmt"
	"time"

	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Profiler struct {
	CostGraph *internal.Graph[SlowPathNode]
	TraceData []ogcore.EventTrace
}

func NewProfiler[E any](graphData *internal.Graph[E], traceData []ogcore.EventTrace) *Profiler {
	costGraph, _ := internal.MapToNewGraph(graphData, func(e E) (SlowPathNode, error) {
		return SlowPathNode{}, nil
	})

	nodeCostMap := GetNodeCostMap(traceData)

	for _, v := range costGraph.Vertices {
		v.Elem.Name = v.Name
		v.Elem.Cost = nodeCostMap[v.Name].TotalCost
	}

	profiler := &Profiler{
		CostGraph: costGraph,
		TraceData: traceData,
	}

	return profiler
}

func (profiler *Profiler) GetSlowHint() string {
	steps, _ := profiler.CostGraph.Steps()
	if len(steps) == 0 {
		return ""
	}

	for _, stepNodes := range steps {
		for _, nodeName := range stepNodes {
			var mostCostDepNode *SlowPathNode

			v := profiler.CostGraph.Vertices[nodeName]

			for _, dep := range v.Dependencies {
				if mostCostDepNode == nil || mostCostDepNode.CostSum < dep.Elem.CostSum {
					mostCostDepNode = &dep.Elem
				}
			}

			if mostCostDepNode == nil {
				v.Elem.CostSum = v.Elem.Cost
			} else {
				v.Elem.Prev = mostCostDepNode
				v.Elem.CostSum = v.Elem.Cost + mostCostDepNode.CostSum
			}
		}
	}

	lastStepNodes := steps[len(steps)-1]
	var slowPathTail *SlowPathNode

	for _, nodeName := range lastStepNodes {
		v := profiler.CostGraph.Vertices[nodeName]

		if slowPathTail == nil || slowPathTail.CostSum < v.Elem.CostSum {
			slowPathTail = &v.Elem
		}
	}

	return slowPathTail.PrintPath()
}

type SlowPathNode struct {
	Name    string
	Cost    time.Duration
	CostSum time.Duration
	Prev    *SlowPathNode
}

func (n *SlowPathNode) PrintPath() string {
	if n == nil {
		return ""
	}

	if n.Prev == nil {
		return fmt.Sprintf("%s(%s)", n.Name, n.Cost)
	}

	return fmt.Sprintf("%s->%s(%s)", n.Prev.PrintPath(), n.Name, n.Cost)
}
