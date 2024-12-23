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

## More Examples

More examples can be found in the code under the "example" directory.

| file                      | introduce                                                                                              |
| :------------------------ | :----------------------------------------------------------------------------------------------------- |
| e01\_hello\_test.go       | Demonstrate the basic flow.                                                                            |
| e02\_state\_test.go       | Demonstrate how to share state data between nodes.                                                     |
| e03\_factory\_test.go     | Demonstrate how to create nodes using the factory pattern.                                             |
| e04\_param\_test.go       | Demonstrate how to set node parameters.                                                                |
| e05\_wrapper\_test.go     | Demonstrate how to use the `wrapper` to enhance node functionality.                                    |
| e06\_cluster\_test.go     | Demonstrate how to use the `cluster` to flexibly schedule multiple nodes.                              |
| e07\_global\_test.go      | Demonstrate how to globalize the factory function.                                                     |
| e08\_virtual\_test.go     | Demonstrate how to use virtual nodes to simplify dependency relationships.                             |
| e09\_interrupter\_test.go | Demonstrate how to add interruptions during the execution of `pipeline`.                               |
| e10\_compose\_test.go     | Demonstrate how to combine nested `pipelines`.                                                         |
| e11\_advance\_test.go     | Demonstrate some advanced usage, including graph verification, exporting, and preheating of pipelines. |

## Ready-to-use

The ograph provides some common node implementations:

| Name      | Type         | Function                                                    | Documentation                                  |
| :-------- | :----------- | ----------------------------------------------------------- | :--------------------------------------------- |
| CMD       | General Node | Command execution                                           | [Documentation Link](docs/node_cmd.md)         |
| HttpReq   | General Node | HTTP request                                                | [Documentation Link](docs/node_http_req.md)    |
| Choose    | Cluster      | Select one node to execute from multiple nodes              | Work in progress                               |
| Parallel  | Cluster      | Concurrent execution of multiple nodes                      | [Documentation Link](docs/cluster_parallel.md) |
| Queue     | Cluster      | Sequential execution of multiple nodes in a queue           | [Documentation Link](docs/cluster_queue.md)    |
| Race      | Cluster      | Concurrent nodes competing to execute                       | Work in progress                               |
| Async     | Wrapper      | Asynchronous execution of the wrapped node                  | [Documentation Link](docs/wrapper_async.md)    |
| Condition | Wrapper      | Conditionally determine whether to execute the wrapped node | Work in progress                               |
| Debug     | Wrapper      | Debugging the wrapped node                                  | [Documentation Link](docs/wrapper_debug.md)    |
| Delay     | Wrapper      | Delay the execution of the wrapped node                     | [Documentation Link](docs/wrapper_delay.md)    |
| Loop      | Wrapper      | Loop the execution of the wrapped node                      | [Documentation Link](docs/wrapper_loop.md)     |
| Retry     | Wrapper      | Retry failed nodes                                          | [Documentation Link](docs/wrapper_retry.md)    |
| Silent    | Wrapper      | Suppress errors and failures of the node                    | [Documentation Link](docs/wrapper_silent.md)   |
| Timeout   | Wrapper      | Node timeout control                                        | [Documentation Link](docs/wrapper_timeout.md)  |
| Trace     | Wrapper      | Trace the execution process of the wrapped node             | [Documentation Link](docs/wrapper_trace.md)    |

## Q&A

> What are the limitations of exporting and importing graphs?

All nodes need to be created using a factory method, and the import graph pipeline must be registered with the factory associated with the imported node.

> Why do we provide multiple node create methods (UseNode, UseFactory, UseFn)?

For simple scenarios, it's convenient to register a singleton and run functions directly. However, when considering pipeline concurrency issues and graph import/export, we need to use the factory method.

> Is the access to State safe for concurrent use?

By default, the State access is safe for concurrent use, but if a custom implementation is used, concurrency safety cannot be guaranteed.

> How to achieve optimal performance? Are there any best practices?

Since coroutines are lightweight and flexible, they usually don't require adjustments or optimizations. If node initialization is slow, you can consider preheating the worker pool.
