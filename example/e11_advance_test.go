package example

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/global"
	"github.com/symphony09/ograph/ogcore"
	"github.com/symphony09/ograph/ogimpl"
)

func TestAdvance_Check(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Person{})
	liSi := ograph.NewElement("LiSi").UseNode(&Person{})

	pipeline.Register(zhangSan, ograph.Rely(liSi)).
		Register(liSi, ograph.Rely(zhangSan))

	if err := pipeline.Check(); err == nil {
		t.Error("unexpect nil err")
	} else {
		fmt.Println(err)
	}

	pipeline2 := ograph.NewPipeline()

	wangWu := ograph.NewElement("WangWu").UseFactory("404")

	pipeline2.Register(wangWu)

	if err := pipeline2.Check(); err == nil {
		t.Error("unexpect nil err")
	} else if !errors.Is(err, ograph.ErrFactoryNotFound) {
		t.Errorf("unexpect err: %#v", err)
	} else {
		fmt.Println(err)
	}
}

func TestAdvance_BatchOp(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Person{})
	liSi := ograph.NewElement("LiSi").UseNode(&Person{})

	pipeline.Register(zhangSan).
		Register(liSi, ograph.Rely(zhangSan))

	pipeline.ForEachElem(func(e *ograph.Element) { e.Wrap(ogimpl.Trace) })

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

func TestAdvance_WarmUp(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.RegisterFactory("SlowReady", func() ogcore.Node {
		time.Sleep(time.Millisecond * 50)
		return &Person{}
	})

	zhangSan := ograph.NewElement("ZhangSan").UseFactory("SlowReady")

	pipeline.Register(zhangSan).SetPoolCache(1, true)

	start := time.Now()

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("time cost: %s\n", time.Since(start))
	}
}

func TestAdvance_DumpAndLoad(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseFactory("Person")
	liSi := ograph.NewElement("LiSi").UseFactory("Person")

	pipeline.Register(zhangSan).
		Register(liSi, ograph.Rely(zhangSan))

	if graphData, err := pipeline.DumpGraph(); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(graphData))

		newPipeline := ograph.NewPipeline()
		newPipeline.LoadGraph(graphData)

		if err := newPipeline.Run(context.TODO(), nil); err != nil {
			t.Error(err)
		}
	}
}

func TestAdvance_DumpDOT(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.RegisterFactory("Cluster", func() ogcore.Node { return &ograph.BaseCluster{} })
	pipeline.RegisterFactory("Node", func() ogcore.Node { return &ograph.BaseNode{} })

	begin := ograph.NewElement("Begin").AsVirtual()

	n1 := ograph.NewElement("Node_1").UseFn(func() error { return nil })
	n2 := ograph.NewElement("Node_2").UseFactory("Node")
	c1 := ograph.NewElement("Cluster_1").UseFactory("Cluster", n1, n2)
	n3 := ograph.NewElement("Node_3").UseNode(&ograph.BaseNode{})

	end := ograph.NewElement("End").AsVirtual()

	pipeline.Register(begin, ograph.Then(c1, n3)).
		Register(end, ograph.Rely(c1, n3))

	if data, err := pipeline.DumpDOT(); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(data))
	}
}

type CustomInitPerson struct {
	FullName string
}

func (node *CustomInitPerson) Init(params map[string]any) error {
	node.FullName = fmt.Sprintf("%v %v", params["FirstName"], params["LastName"])
	return nil
}

func (node *CustomInitPerson) Run(ctx context.Context, state ogcore.State) error {
	fmt.Println("Full name is:", node.FullName)
	return nil
}

func TestAdvance_CustomInit(t *testing.T) {
	pipeline := ograph.NewPipeline()
	e := ograph.NewElement("Person").
		UsePrivateFactory(func() ogcore.Node {
			return &CustomInitPerson{}
		}).
		Params("FirstName", "Jack").Params("LastName", "Chen")

	if err := pipeline.Register(e).Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

func TestAdvance_CustomLog(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	pipeline := ograph.NewPipeline()
	e := ograph.NewElement("ZhangSan").
		UseNode(&Person{}).
		Wrap(ogimpl.Trace)

	if err := pipeline.Register(e).Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

func init() {
	global.Factories.Add("Person", func() ogcore.Node { return &Person{} })
}

type TxNode struct {
	ograph.BaseNode
}

func (tx *TxNode) Run(ctx context.Context, state ogcore.State) error {
	fmt.Println("run tx", tx.Name())
	return nil
}

func (tx *TxNode) Commit() {
	fmt.Println("commit tx", tx.Name())
}

func (tx *TxNode) Rollback() {
	fmt.Println("rollback tx", tx.Name())
}

func TestAdvance_Transaction(t *testing.T) {
	pipeline := ograph.NewPipeline()
	exceptErr := errors.New("except error")

	n1 := ograph.NewElement("n1").UseFn(func() error {
		return nil
	})
	n2 := ograph.NewElement("n2").UseFn(func() error {
		fmt.Println("an error occurred")
		return exceptErr
	})
	t1 := ograph.NewElement("t1").UseNode(&TxNode{})
	t2 := ograph.NewElement("t2").UseNode(&TxNode{})

	pipeline.Register(n1, ograph.Branch(t1, t2))

	fmt.Println("[pipeline without error]")
	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}

	pipeline = ograph.NewPipeline()
	pipeline.Register(n1, ograph.Branch(t1, n2, t2))

	fmt.Println("[pipeline with error]")
	if err := pipeline.Run(context.TODO(), nil); errors.Unwrap(err).Error() != "except error" {
		t.Error(err)
	}
}
