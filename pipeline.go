package ograph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"time"

	"github.com/symphony09/ograph/global"
	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
	"github.com/symphony09/ograph/profile"
)

var ErrFactoryNotFound error = errors.New("factory not found")

type Pipeline struct {
	BaseNode
	Builder
	*slog.Logger

	graph    *PGraph
	elements map[string]*Element
	pool     internal.WorkerPool

	Interrupts       iter.Seq[string]
	ParallelismLimit int
	DisablePool      bool
	EnableMonitor    bool
	SlowThreshold    time.Duration
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
	var worker *internal.Worker

	if !pipeline.DisablePool {
		if poolWorker, ok := pipeline.pool.Get(); ok {
			worker = poolWorker
		}
	}

	if worker == nil {
		if newWorker, err := pipeline.build(pipeline.graph); err != nil {
			return err
		} else {
			worker = newWorker
		}
	}

	if !pipeline.DisablePool {
		defer func() {
			pipeline.pool.Put(worker)
		}()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if state == nil {
		state = NewState()
	}

	params := pipeline.newWorkParams()
	if pipeline.EnableMonitor {
		start := time.Now()
		params.Tracker = new(ogcore.Tracker)

		defer func() {
			if pipeline.SlowThreshold > 0 && time.Since(start) > pipeline.SlowThreshold {
				go func() {
					profiler := profile.NewProfiler(pipeline.graph, params.Tracker.TraceData)
					pipeline.Logger.Warn("monitor slow execution", "Pipeline", pipeline.Name(), "SlowHint", profiler.GetSlowHint())
				}()
			}
		}()
	}

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

func (pipeline *Pipeline) ResetPool() {
	pipeline.pool.Reset()
}

func (pipeline *Pipeline) newWorkParams() *internal.WorkParams {
	params := &internal.WorkParams{}
	if pipeline.ParallelismLimit > 0 {
		params.GorLimit = pipeline.ParallelismLimit
	} else {
		params.GorLimit = -1
	}

	if pipeline.Interrupts == nil {
		params.Interrupts = func(yield func(string) bool) {}
	} else {
		params.Interrupts = pipeline.Interrupts
	}

	return params
}

func (pipeline *Pipeline) DumpGraph() ([]byte, error) {
	marshaler := internal.NewGraphMarshaler(pipeline.graph)
	return json.Marshal(marshaler)
}

func (pipeline *Pipeline) LoadGraph(data []byte) error {
	marshaler := &internal.GraphMarshaler[*Element]{}

	if err := json.Unmarshal(data, marshaler); err != nil {
		return err
	}

	pipeline.graph = marshaler.GenerateGraph()

	pipeline.ResetPool()

	return nil
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		graph:    internal.NewGraph[*Element](),
		elements: make(map[string]*Element),

		ParallelismLimit: -1,

		Logger: slog.Default(),
	}
}
