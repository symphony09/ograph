package ograph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"time"

	"github.com/symphony09/eventd"
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
	eventBus *eventd.EventBus[ogcore.State]

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
		if elem.Virtual {
			return nil
		}

		if elem.FactoryName != "" {
			if factories.Get(elem.FactoryName) == nil {
				return fmt.Errorf("%w, name: %s", ErrFactoryNotFound, elem.FactoryName)
			}

			for _, subElem := range elem.SubElements {
				if err := elemCheck(subElem); err != nil {
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
	newCtx, newState, worker, params, afterRun, err := pipeline.prepare(ctx, state)
	if err != nil {
		return err
	}

	defer afterRun()

	return worker.Work(newCtx, newState, params)
}

func (pipeline *Pipeline) AsyncRun(ctx context.Context, state ogcore.State) (pause, continueRun func(), wait func() error) {
	pause, continueRun = func() {}, func() {}

	newCtx, newState, worker, params, afterRun, err := pipeline.prepare(ctx, state)
	if err != nil {
		wait = func() error {
			return err
		}
		return
	}

	errCh := make(chan error, 1)
	go func() {
		defer afterRun()
		errCh <- worker.Work(newCtx, newState, params)
	}()

	params.ContinueCond = sync.NewCond(&sync.Mutex{})

	pause = func() {
		params.ContinueCond.L.Lock()
		params.Pause = true
		params.ContinueCond.L.Unlock()
	}

	continueRun = func() {
		params.ContinueCond.L.Lock()
		params.Pause = false
		params.ContinueCond.L.Unlock()
		params.ContinueCond.Broadcast()
	}

	wait = func() error {
		return <-errCh
	}

	return
}

func (pipeline *Pipeline) prepare(ctx context.Context, state ogcore.State) (context.Context, ogcore.State,
	*internal.Worker, *internal.WorkParams, func(), error) {

	if ctx == nil {
		ctx = context.Background()
	}

	if state == nil {
		state = NewState()
	}

	pool := &pipeline.pool

	var worker *internal.Worker

	if !pipeline.DisablePool {
		if poolWorker, ok := pool.Get(); ok {
			worker = poolWorker
		}
	}

	if worker == nil {
		if newWorker, err := pipeline.build(pipeline.graph, pipeline.eventBus); err != nil {
			return ctx, state, nil, nil, nil, err
		} else {
			worker = newWorker
		}
	}

	params := &internal.WorkParams{}
	if pipeline.ParallelismLimit > 0 {
		params.GorLimit = pipeline.ParallelismLimit
	} else {
		params.GorLimit = -1
	}
	if pipeline.EnableMonitor {
		params.Tracker = new(ogcore.Tracker)
		params.Tracker.StartTime = time.Now()
	}
	params.Interrupts = pipeline.Interrupts

	afterRun := func() {
		if !pipeline.DisablePool {
			pool.Put(worker)
		}

		if pipeline.EnableMonitor {
			if pipeline.SlowThreshold > 0 && time.Since(params.Tracker.StartTime) > pipeline.SlowThreshold {
				go func() {
					profiler := profile.NewProfiler(pipeline.graph, params.Tracker.TraceData)
					pipeline.Logger.Warn("monitor slow execution", "Pipeline", pipeline.Name(), "SlowHint", profiler.GetSlowHint())
				}()
			}
		}
	}

	return ctx, state, worker, params, afterRun, nil
}

func (pipeline *Pipeline) SetPoolCache(size int, warmup bool) error {
	pipeline.pool = *internal.NewWorkerPool(size)

	if warmup {
		for i := 0; i < size; i++ {
			if worker, err := pipeline.build(pipeline.graph, pipeline.eventBus); err != nil {
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

func (pipeline *Pipeline) Subscribe(callback eventd.CallBack[ogcore.State], ops ...eventd.Op) (cancel func(), err error) {
	return pipeline.eventBus.Subscribe(callback, ops...)
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		graph:    internal.NewGraph[*Element](),
		elements: make(map[string]*Element),
		eventBus: new(eventd.EventBus[ogcore.State]),

		ParallelismLimit: -1,

		Logger: slog.Default(),
	}
}
