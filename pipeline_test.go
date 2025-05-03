package ograph

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/symphony09/eventd"
	"github.com/symphony09/ograph/internal"
	"github.com/symphony09/ograph/ogcore"
)

func TestPipeline_Register(t *testing.T) {
	p := NewPipeline()
	t1 := NewElement("t1")
	t2 := NewElement("t2")

	p.Register(t1, Then(t2))

	if !p.graph.Edges[internal.GraphEdge[*Element]{
		From: p.graph.Vertices["t1"],
		To:   p.graph.Vertices["t2"],
	}] {
		t.Errorf("edge from t1 to t2 = false, want true")
	}
}

func TestPipeline_ForEachElem(t *testing.T) {
	p := NewPipeline()
	t1 := NewElement("t1")
	t2 := NewElement("t2")

	p.Register(t1, Then(t2))

	p.ForEachElem(func(e *Element) {
		e.Wrap("test")
	})

	for _, e := range p.elements {
		if len(e.Wrappers) != 1 || e.Wrappers[0] != "test" {
			t.Errorf("e.Wrappers = %v, want %v", e.Wrappers, []string{"test"})
		}
	}
}

func TestPipeline_Check(t *testing.T) {
	p := NewPipeline()

	if err := p.Check(); err != nil {
		t.Errorf("p.Check() = %v, want nil", err)
	}

	t1 := NewElement("t1").AsVirtual()
	p.Register(t1)

	if err := p.Check(); err != nil {
		t.Errorf("p.Check() = %v, want nil", err)
	}

	t1.Virtual = false
	wantErr := fmt.Errorf("%w, name: %s", ErrSingletonNotSet, "t1")

	if err := p.Check(); err.Error() != wantErr.Error() {
		t.Errorf("p.Check() = %v, want %v", err, wantErr)
	}

	t1.UseFactory("fake_factory")
	wantErr = fmt.Errorf("%w, name: %s", ErrFactoryNotFound, "fake_factory")

	if err := p.Check(); err.Error() != wantErr.Error() {
		t.Errorf("p.Check() = %v, want %v", err, wantErr)
	}

	p.RegisterFactory("fake_factory", func() ogcore.Node { return nil })

	if err := p.Check(); err != nil {
		t.Errorf("p.Check() = %v, want nil", err)
	}

	t2 := NewElement("sub_t").UseFactory("fake_factory2")
	t1.UseFactory("fake_factory", t2)

	wantErr = fmt.Errorf("%w, name: %s", ErrFactoryNotFound, "fake_factory2")

	if err := p.Check(); err.Error() != wantErr.Error() {
		t.Errorf("p.Check() = %v, want %v", err, wantErr)
	}

	t2.AsVirtual()
	p.Register(t1, Then(t1))
	wantErr = fmt.Errorf("found cycle between vertices: %v", []string{"t1"})

	if err := p.Check(); err.Error() != wantErr.Error() {
		t.Errorf("p.Check() = %v, want %v", err, wantErr)
	}

	p2 := NewPipeline()
	p2.Register(NewElement("sub_p").UseNode(p))

	if err := p2.Check(); err.Error() != wantErr.Error() {
		t.Errorf("p2.Check() = %v, want %v", err, wantErr)
	}
}

func TestPipeline_Run_1(t *testing.T) {
	var cnt atomic.Int32
	var n BaseNode
	n.Action = func(ctx context.Context, state ogcore.State) error {
		cnt.Add(1)
		return nil
	}

	p := NewPipeline()

	start := NewElement("start").AsVirtual()

	t1 := NewElement("t1").UseNode(&n)

	t2 := NewElement("t2").UseFactory("t")

	t3 := NewElement("t3").UseFn(func() error {
		if n := cnt.Load(); n != 2 {
			return fmt.Errorf("cnt = %d, want 2", n)
		}
		return nil
	})

	p.Register(start, Then(t1, t2)).Register(t3, Rely(t1, t2))

	var ctx context.Context
	if err := p.Run(ctx, nil); err == nil {
		t.Error("p.Run() got error = nil, want not nil")
	}

	p.RegisterFactory("t", func() ogcore.Node { return &n })

	if err := p.Run(ctx, nil); err != nil {
		t.Errorf("p.Run() got error = %v, want nil", err)
	}
}

func TestPipeline_Run_2(t *testing.T) {
	p := NewPipeline()
	p.EnableMonitor = true
	p.SlowThreshold = 50 * time.Millisecond

	t1 := NewElement("t1").UseFn(func() error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	t2 := NewElement("t2").UseFn(func() error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})

	p.Register(t1).Register(t2)

	if err := p.Run(context.Background(), nil); err != nil {
		t.Errorf("p.Run() got error = %v, want nil", err)
	}

	p.ParallelismLimit = 1

	startTime := time.Now()

	if err := p.Run(context.Background(), nil); err != nil {
		t.Errorf("p.Run() got error = %v, want nil", err)
	}

	if cost := time.Since(startTime); cost < 50*time.Millisecond {
		t.Errorf("p.Run() cost time = %v, want more than 50ms", cost)
	}
}

type stateKey string

type TxNode struct {
	BaseEventNode
}

func (tx *TxNode) Run(ctx context.Context, state ogcore.State) error {
	exportState := NewState()
	SavePrivateState[stateKey](exportState, "name", tx.Name(), true)
	tx.Emit("running", exportState)
	return nil
}

func (tx *TxNode) Commit() {
	state := NewState()
	SaveState(state, "name", tx.Name(), true)
	tx.Emit("commit", state)
}

func (tx *TxNode) Rollback() {
	state := NewState()
	SaveState(state, "name", tx.Name(), true)
	tx.Emit("rollback", state)
}

type TSilent struct {
	BaseWrapper
	ParameterX string
}

func (n *TSilent) Run(ctx context.Context, state ogcore.State) error {
	if n.ParameterX != "x" {
		return errors.New("wrong parameter")
	}
	n.Node.Run(ctx, state)
	return nil
}

type TCluster struct {
	BaseCluster
	ParameterX string
}

func (n *TCluster) Run(ctx context.Context, state ogcore.State) error {
	if n.ParameterX != "x" {
		return errors.New("wrong parameter")
	}
	return n.BaseCluster.Run(ctx, state)
}

type TNode struct {
	BaseNode
	ParameterX int
}

func (n *TNode) Init(params map[string]any) error {
	n.ParameterX, _ = strconv.Atoi(params["ParameterX"].(string))
	return nil
}

func (n *TNode) Run(ctx context.Context, state ogcore.State) error {
	if n.ParameterX != 1 {
		return errors.New("wrong parameter")
	}
	return nil
}

func TestPipeline_Run_3(t *testing.T) {
	p := NewPipeline()

	start := NewElement("start").AsVirtual()
	transactionStart := NewElement("t_start").AsVirtual()

	tx := NewElement("tx").UseNode(&TxNode{})
	tErr := NewElement("t_err").UseFn(func() error {
		return errors.New("t_err")
	})

	p.Register(transactionStart, Rely(start), Branch(tx, tErr))

	p2 := NewPipeline()
	p2.RegisterFactory("silent", func() ogcore.Node {
		return &TSilent{}
	})
	p2.RegisterFactory("cluster", func() ogcore.Node {
		return &TCluster{}
	})
	p2.RegisterFactory("t", func() ogcore.Node {
		return &TNode{}
	})

	t1 := NewElement("t1").UseFactory("t").Params("ParameterX", "1")
	t2 := NewElement("t2").UseNode(&TxNode{})
	c1 := NewElement("c1").UseFactory("cluster", t1, t2).Params("ParameterX", "x")

	t3 := NewElement("t3").UseNode(NewFuncNode(func(ctx context.Context, state ogcore.State) error {
		SavePrivateState[stateKey](state, "cnt", 1, true)
		return nil
	}))
	t4 := NewElement("t4").UseNode(NewFuncNode(func(ctx context.Context, state ogcore.State) error {
		var output int
		UpdatePrivateState[stateKey](state, "cnt", func(oldVal int) (val int) {
			val = oldVal + 1
			output = val
			return
		})
		state.Set("output", output)
		return nil
	}))

	p2.Register(NewElement("sub_p").UseNode(p).Apply(func(e *Element) {
		e.WrapByAlias("silent", "nothing").Params("nothing.ParameterX", "x")
	})).Register(c1).Register(t3, Rely(c1), Then(t4))

	var txCnt int
	events := make(map[string]string)

	p2.Subscribe(func(event string, obj ogcore.State) bool {
		if event == "running" {
			if LoadPrivateState[stateKey, string](obj, "name") == "" {
				t.Error("got empty running tx name")
			}
			txCnt++
		} else if event == "commit" {
			name := LoadState[string](obj, "name")
			events[name] = "commit"
		} else if event == "rollback" {
			name := LoadState[string](obj, "name")
			events[name] = "rollback"
		}
		return true
	}, eventd.On(".*"))

	state := NewState()

	if err := p2.Run(context.Background(), state); err != nil {
		t.Errorf("p2.Run() got error = %v, want nil", err)
	}

	if txCnt != 2 {
		t.Errorf("got txCnt = %d, want 2", txCnt)
	}

	if events["tx"] != "rollback" {
		t.Errorf("got tx event = %s, want rollback", events["tx"])
	}

	if output := LoadState[int](state, "output"); output != 2 {
		t.Errorf("got output = %d, want 2", output)
	}
}

func TestPipeline_AsyncRun(t *testing.T) {
	p := NewPipeline()

	f := func() error {
		time.Sleep(20 * time.Millisecond)
		return nil
	}

	t1 := NewElement("t1").UseFn(f)
	t2 := NewElement("t2").UseFn(f)

	p.Register(t1, Then(t2))

	startTime := time.Now()

	pause, continueRun, wait := p.AsyncRun(context.Background(), nil)

	time.Sleep(10 * time.Millisecond)

	pause()

	time.Sleep(20 * time.Millisecond)

	continueRun()

	if err := wait(); err != nil {
		t.Errorf("wait() got error = %v, want nil", err)
	}

	if cost := time.Since(startTime); cost < 50*time.Millisecond {
		t.Errorf("p.Run() cost time = %v, want more than 50ms", cost)
	}

	p2 := NewPipeline()
	p2.Register(NewElement("t1").UseFactory("fake_factory"))

	_, _, wait = p2.AsyncRun(context.Background(), nil)
	if err := wait(); err == nil {
		t.Error("wait() got error = nil, want not nil")
	}
}

func TestPipeline_Pool(t *testing.T) {
	var cnt int

	p := NewPipeline()

	p.RegisterFactory("t", func() ogcore.Node {
		cnt++
		return &BaseNode{}
	})

	p.Register(NewElement("t1").UseFactory("t"))

	p.SetPoolCache(1, true)

	for range 100 {
		p.Run(context.Background(), nil)

		if cnt != 1 {
			t.Errorf("got cnt = %d, want 1", cnt)
			return
		}
	}

	p.ResetPool()

	for range 100 {
		p.Run(context.Background(), nil)

		if cnt != 2 {
			t.Errorf("got cnt = %d, want 2", cnt)
			return
		}
	}
}

func TestPipeline_DumpAndLoadGraph(t *testing.T) {
	var cnt int

	n := &BaseNode{
		Action: func(ctx context.Context, state ogcore.State) error {
			cnt++
			return nil
		},
	}

	p := NewPipeline()

	t1 := NewElement("t1").UseFactory("t")
	t2 := NewElement("t2").UseFactory("t")

	p.Register(t1, Then(t2))

	p2 := NewPipeline()
	p2.Register(NewElement("sub_p").UseNode(p))

	d, err := p2.DumpGraph()
	if err != nil {
		t.Errorf("p2.DumpGraph() got err = %v, want nil", err)
	}

	p3 := NewPipeline()
	p3.RegisterFactory("t", func() ogcore.Node { return n })

	if err := p3.LoadGraph(d); err != nil {
		t.Errorf("p3.LoadGraph() got err = %v, want nil", err)
	}

	if err := p3.Run(context.Background(), nil); err != nil {
		t.Errorf("p3.Run() got err = %v, want nil", err)
	}

	if cnt != 2 {
		t.Errorf("got cnt = %d, want 2", cnt)
	}
}
