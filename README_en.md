                    ________________                     ______  
                    __  __ \_  ____/____________ ___________  /_ 
                    _  / / /  / __ __  ___/  __ `/__  __ \_  __ \
                    / /_/ // /_/ / _  /   / /_/ /__  /_/ /  / / /
                    \____/ \____/  /_/    \__,_/ _  .___//_/ /_/ 
    			                                 /_/             

# OGraph: A simple way to build a pipeline with Go

<p align="left">
  <a href="https://github.com/symphony09/ograph"><img src="https://badgen.net/badge/langs/Golang?list=1" alt="languages"></a>
  <a href="https://github.com/symphony09/ograph"><img src="https://badgen.net/badge/os/MacOS,Linux,Windows/cyan?list=1" alt="os"></a>
</p>

[中文](README.md) | [English](README_en.md)

**OGraph** is a graph execution framework implemented in Go.

You can control the scheduling of sequential execution of dependent elements and concurrent execution of non-dependent elements by building a `Pipeline`.

In addition, OGraph also provides a range of features out-of-the-box, including retry limits, timeout settings, and execution tracking.

## Comparison with similar projects

**OGraph** was inspired by another C++ project, [CGraph](https://github.com/ChunelFeng/CGraph). However, OGraph is not the Go version of CGraph.

### Feature Comparison

Like CGraph, OGraph also provides basic graph construction and scheduling execution capabilities. However, there are several key differences:

*   Implemented in Go, using coroutines instead of threads for scheduling, making it lighter and more flexible.

*   Supports customizing loop, condition, error handling, and other logic through `Wrapper`, which can be combined freely.

*   Supports exporting graph structure and importing it for execution (within the constraints).

*    Flexible virtual node settings to simplify dependencies.

### Performance Comparison

After benchmarking, the performance of OGraph and CGraph are on the same level. However, OGraph has an advantage in performance in io-intensive scenarios.

[CGraph Performance test reference](http://www.chunel.cn/archives/cgraph-compare-taskflow-v1)

OGraph Performance test reference

Limit: 8 cores, three scenarios (concurrent 32 nodes, sequential 32 nodes, complex scenario simulating 6 nodes) each executed 1 million times.

```bash
cd test
go test -bench='(Concurrent_32|Serial_32|Complex_6)$' -benchtime=1000000x -benchmem -cpu=8
```

outputs

    goos: linux
    goarch: amd64
    pkg: github.com/symphony09/ograph/test
    cpu: AMD Ryzen 5 5600G with Radeon Graphics         
    BenchmarkConcurrent_32-8         1000000              9669 ns/op            2212 B/op         64 allocs/op
    BenchmarkSerial_32-8             1000000              1761 ns/op             712 B/op         15 allocs/op
    BenchmarkComplex_6-8             1000000              3118 ns/op            1152 B/op         26 allocs/op
    PASS
    ok      github.com/symphony09/ograph/test       14.553s

## Quick Start

### Step 1: Declare a Node interface implementation.

```go
type Person struct {
	ograph.BaseNode
}

func (person *Person) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("Hello, i am %s.\n", person.Name())
	return nil
}
```

In the code above, the Person struct combines the BaseNode and overrides the Node interface method Run.

### Step 2: Build a Pipeline and run it.

```go
func TestHello(t *testing.T) {
	pipeline := ograph.NewPipeline()

	zhangSan := ograph.NewElement("ZhangSan").UseNode(&Person{})
	liSi := ograph.NewElement("LiSi").UseNode(&Person{})

	pipeline.Register(zhangSan).
		Register(liSi, ograph.Rely(zhangSan))

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
```

In the code above, the two Person nodes (zhangSan and liSi) in the pipeline are registered, and liSi is specified to depend on zhangSan.

outputs

    Hello, i am ZhangSan.
    Hello, i am LiSi.

## More Documents

Please follow the documentation at [https://symphony09.github.io/ograph-docs](https://symphony09.github.io/ograph-docs/docs/quick-start/)!