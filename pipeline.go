package ograph

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

type Pipeline struct {
	BaseNode
	Builder

	graph    *PGraph
	elements map[string]*Element
	pool     internal.WorkerPool

	VirtualSlots     map[string]ogcore.Action
	Interrupters     []ogcore.Interrupter
	ParallelismLimit int
}

func (pipeline *Pipeline) Register(e *Element, ops ...Op) *Pipeline {
	if pipeline.elements[e.Name] == nil {
		pipeline.elements[e.Name] = e
		pipeline.graph.AddVertex(e.Name, e)
	}

	for _, op := range ops {
		op(pipeline, e)
	}

	return pipeline
}

func (pipeline *Pipeline) ForEachElem(op func(e *Element)) *Pipeline {
	for _, e := range pipeline.elements {
		op(e)
	}

	return pipeline
}

func (pipeline *Pipeline) RegisterInterrupt(handler ogcore.InterruptHandler, on ...string) *Pipeline {
	pipeline.Interrupters = append(pipeline.Interrupters, ogcore.Interrupter{
		Handler: handler,
		Points:  on,
	})
	return pipeline
}

func (pipeline *Pipeline) Check() error {
	return pipeline.graph.Check()
}

func (pipeline *Pipeline) Run(ctx context.Context, state ogcore.State) error {
	worker, ok := pipeline.pool.Get()

	if !ok {
		var buildErr error

		worker, buildErr = pipeline.build(pipeline.graph)
		if buildErr != nil {
			return buildErr
		}
	}

	defer func() {
		pipeline.pool.Put(worker)
	}()

	if ctx == nil {
		ctx = context.Background()
	}

	if state == nil {
		state = NewState()
	}

	pipeline.prepare(worker)

	return worker.Work(ctx, state)
}

func (pipeline *Pipeline) SetPoolCache(size int, warmup bool) error {
	pipeline.pool = *internal.NewWorkerPool(size)

	if warmup {
		for i := 0; i < size; i++ {
			if worker, err := pipeline.build(pipeline.graph); err != nil {
				return err
			} else {
				pipeline.pool.Put(worker)
			}
		}
	}

	return nil
}

func (pipeline *Pipeline) prepare(worker *internal.Worker) {
	if pipeline.ParallelismLimit > 0 {
		worker.GorLimit = pipeline.ParallelismLimit
	}

	actions := make(map[string]ogcore.Action)
	maps.Copy(actions, pipeline.VirtualSlots)

	for _, interrupter := range pipeline.Interrupters {
		interrupt := ogcore.NewInterrupt(interrupter.Handler)

		for _, point := range interrupter.Points {
			if action := actions[point]; action == nil {
				actions[point] = func(ctx context.Context, state ogcore.State) error {
					err, _ := interrupt(ogcore.NewInterruptCtx(ctx, point, state))
					return err
				}
			} else {
				actions[point] = func(ctx context.Context, state ogcore.State) error {
					if err, _ := interrupt(ogcore.NewInterruptCtx(ctx, point, state)); err != nil {
						return err
					}

					return action(ctx, state)
				}
			}
		}
	}

	worker.DynamicActions = actions
}

func (pipeline *Pipeline) DumpGraph() ([]byte, error) {
	marsher := internal.NewGraphMarshaler(pipeline.graph)
	return json.Marshal(marsher)
}

func (pipeline *Pipeline) LoadGraph(data []byte) error {
	marshaler := &internal.GraphMarshaler[*Element]{}

	if err := json.Unmarshal(data, marshaler); err != nil {
		return err
	}

	pipeline.graph = marshaler.GenerateGraph()
	return nil
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		graph:    internal.NewGraph[*Element](),
		elements: make(map[string]*Element),

		ParallelismLimit: -1,

		VirtualSlots: make(map[string]ogcore.Action),
	}
}
