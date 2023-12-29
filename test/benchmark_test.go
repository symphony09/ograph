package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

func BenchmarkBase(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} }).
		RegisterFactory("BC", func() ogcore.Node { return &ograph.BaseCluster{} }).
		RegisterFactory("BW", func() ogcore.Node { return &ograph.BaseWrapper{} })

	e1 := ograph.NewElement("n1").UseFactory("BN")
	e2 := ograph.NewElement("n2").UseFactory("BN")
	e3 := ograph.NewElement("n3").UseFactory("BN").Wrap("BW")
	e4 := ograph.NewElement("c1").UseFactory("BC", e3)

	pipeline.Register(e1).
		Register(e2, ograph.DependOn(e1)).
		Register(e4, ograph.DependOn(e2))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := pipeline.Run(context.TODO(), nil); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkConcurrent_32(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} })

	for i := 0; i < 32; i++ {
		element := ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN")

		pipeline.Register(element)
	}

	// pipeline.ParallelismLimit = 1

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := pipeline.Run(context.TODO(), nil); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkConcurrent_32_Parallel(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} })

	for i := 0; i < 32; i++ {
		element := ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN")

		pipeline.Register(element)
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if err := pipeline.Run(context.TODO(), nil); err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkSerial_32(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} }).
		RegisterFactory("BC", func() ogcore.Node { return &ograph.BaseCluster{} })

	var lastElem *ograph.Element
	for i := 0; i < 32; i++ {
		elem := ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN")

		if lastElem == nil {
			pipeline.Register(elem)
		} else {
			pipeline.Register(elem, ograph.DependOn(lastElem))
		}

		lastElem = elem
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := pipeline.Run(context.TODO(), nil); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkSerial_32_Parallel(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} }).
		RegisterFactory("BC", func() ogcore.Node { return &ograph.BaseCluster{} })

	var lastElem *ograph.Element
	for i := 0; i < 32; i++ {
		elem := ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN")

		if lastElem == nil {
			pipeline.Register(elem)
		} else {
			pipeline.Register(elem, ograph.DependOn(lastElem))
		}

		lastElem = elem
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if err := pipeline.Run(context.TODO(), nil); err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkComplex_6(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} })

	var elements []*ograph.Element
	for i := 0; i < 6; i++ {
		elements = append(elements, ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN"))
	}
	a, b1, b2, c1, c2, d := elements[0], elements[1], elements[2], elements[3], elements[4], elements[5]

	pipeline.Register(a).
		Register(b1, ograph.DependOn(a)).
		Register(b2, ograph.DependOn(b1)).
		Register(c1, ograph.DependOn(a)).
		Register(c2, ograph.DependOn(c1)).
		Register(d, ograph.DependOn(b2, c2))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := pipeline.Run(context.TODO(), nil); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkComplex_6_Parallel(b *testing.B) {
	pipeline := ograph.NewPipeline()
	pipeline.Builder.
		RegisterFactory("BN", func() ogcore.Node { return &ograph.BaseNode{} })

	var elements []*ograph.Element
	for i := 0; i < 6; i++ {
		elements = append(elements, ograph.NewElement(fmt.Sprintf("n%d", i)).UseFactory("BN"))
	}
	a, b1, b2, c1, c2, d := elements[0], elements[1], elements[2], elements[3], elements[4], elements[5]

	pipeline.Register(a).
		Register(b1, ograph.DependOn(a)).
		Register(b2, ograph.DependOn(b1)).
		Register(c1, ograph.DependOn(a)).
		Register(c2, ograph.DependOn(c1)).
		Register(d, ograph.DependOn(b2, c2))

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if err := pipeline.Run(context.TODO(), nil); err != nil {
				b.Error(err)
			}
		}
	})
}
