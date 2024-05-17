package ograph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/symphony09/ograph/global"
	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

var ErrFactoryNotFound error = errors.New("factory not found")

type Pipeline struct {
	BaseNode
	Builder

	graph    *PGraph
	elements map[string]*Element
	pool     internal.WorkerPool

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
	factories := pipeline.Builder.Factories
	if factories == nil {
		factories = global.Factories.Clone()
	}

	for _, vertex := range pipeline.graph.Vertices {
		for factory := range vertex.Elem.GetRequiredFactories() {
			if factories.Get(factory) == nil {
				return fmt.Errorf("%w, name: %s", ErrFactoryNotFound, factory)
			}
		}
	}

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

	params := pipeline.newWorkParams()

	return worker.Work(ctx, state, params)
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

func (pipeline *Pipeline) newWorkParams() *internal.WorkParams {
	params := &internal.WorkParams{}
	if pipeline.ParallelismLimit > 0 {
		params.GorLimit = pipeline.ParallelismLimit
	} else {
		params.GorLimit = -1
	}

	params.ActionsBeforeRun, params.ActionsAfterRun = ogcore.GenActionsByIntr(pipeline.Interrupters)

	return params
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
	}
}
