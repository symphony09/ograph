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
var ErrSingletonNotSet error = errors.New("single node not set")

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

	var elemCheck func(elem *Element) error
	elemCheck = func(elem *Element) error {
		if elem.FactoryName != "" {
			if factories.Get(elem.FactoryName) == nil {
				return fmt.Errorf("%w, name: %s", ErrFactoryNotFound, elem.FactoryName)
			}

			for _, subElem := range elem.SubElements {
				if err := elemCheck(subElem); err != nil {
					return err
				}
			}
		} else if elem.Virtual {
			for _, implElem := range elem.ImplElements {
				if err := elemCheck(implElem); err != nil {
					return err
				}
			}
		} else {
			if elem.Singleton == nil {
				return fmt.Errorf("%w, name: %s", ErrSingletonNotSet, elem.Name)
			} else if subPipeline, ok := elem.Singleton.(*Pipeline); ok {
				if err := subPipeline.Check(); err != nil {
					return err
				}
			}
		}

		return nil
	}

	for _, vertex := range pipeline.graph.Vertices {
		if err := elemCheck(vertex.Elem); err != nil {
			return err
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

	for _, v := range pipeline.graph.Vertices {
		if v.Elem.Singleton != nil {
			if subPipeline, ok := v.Elem.Singleton.(*Pipeline); ok {
				if marshaler.SubGraphs == nil {
					marshaler.SubGraphs = make(map[string]*internal.GraphMarshaler[*Element])
				}

				marshaler.SubGraphs[v.Name] = internal.NewGraphMarshaler(subPipeline.graph)
			}
		}
	}

	return json.Marshal(marshaler)
}

func (pipeline *Pipeline) LoadGraph(data []byte) error {
	marshaler := &internal.GraphMarshaler[*Element]{}

	if err := json.Unmarshal(data, marshaler); err != nil {
		return err
	}

	pipeline.graph = marshaler.GenerateGraph()

	for name, subMarshaler := range marshaler.SubGraphs {
		if v, ok := pipeline.graph.Vertices[name]; ok {
			subPipeline := NewPipeline()
			subPipeline.graph = subMarshaler.GenerateGraph()
			v.Elem.Singleton = subPipeline
		}
	}

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
