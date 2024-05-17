package example

import (
	"context"
	"errors"
	"fmt"
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

	pipeline.Register(zhangSan, ograph.DependOn(liSi)).
		Register(liSi, ograph.DependOn(zhangSan))

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
		Register(liSi, ograph.DependOn(zhangSan))

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
		Register(liSi, ograph.DependOn(zhangSan))

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
	end1 := ograph.NewElement("End_1").UseFactory("Node")
	end2 := ograph.NewElement("End_2").UseFactory("Node")

	end1.Implement(end, true)
	end2.Implement(end, false)

	pipeline.Register(begin, ograph.Then(c1, n3)).
		Register(end, ograph.DependOn(c1, n3))

	if data, err := pipeline.DumpDOT(); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(data))
	}
}

func init() {
	global.Factories.Add("Person", func() ogcore.Node { return &Person{} })
}
