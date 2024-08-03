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

**OGraph** 是一个用 `Go` 实现的图流程执行框架。

你可以通过构建`Pipeline`(流水线)，来控制依赖元素依次顺序执行、非依赖元素并发执行的调度功能。

此外，**OGraph** 还提供了丰富的重试，超时限制，执行追踪等开箱即用的特征。

## 同类项目对比

**OGraph** 受启发于另一个 `C++`项目 [CGraph](https://github.com/ChunelFeng/CGraph)。但 OGraph 并不等于 Go 版本的 CGraph。

### 功能对比

和 CGraph 一样，OGraph 也提供基本的构图和调度执行能力，但有以下几点关键不同：

*   用 Go 实现，使用协程而非线程进行调度，更轻量灵活

*   支持通过 Wrapper 来自定义循环、执行条件判断、错误处理等逻辑，并可以随意组合

*   支持导出图结构，再在别处导入执行（符合限制的情况下）

*   灵活的虚节点设置，用以简化依赖关系，以及延迟到运行时决定实际执行的节点，实现多态

### 性能对比

经过 Benchmark 测试，OGraph 和 CGraph 的性能在同一水平。如果在 io 密集场景下，OGraph 更有优势。

[CGraph 性能测试参考](http://www.chunel.cn/archives/cgraph-compare-taskflow-v1)

OGraph 性能测试参考

限制 8 核，三个场景（并发32节点，串行32节点，复杂情况模拟6节点）分别执行 100 w 次

```bash
cd test
go test -bench='(Concurrent_32|Serial_32|Complex_6)$' -benchtime=1000000x -benchmem -cpu=8
```

输出结果

    goos: linux
    goarch: amd64
    pkg: github.com/symphony09/ograph/test
    cpu: AMD Ryzen 5 5600G with Radeon Graphics         
    BenchmarkConcurrent_32-8         1000000              9669 ns/op            2212 B/op         64 allocs/op
    BenchmarkSerial_32-8             1000000              1761 ns/op             712 B/op         15 allocs/op
    BenchmarkComplex_6-8             1000000              3118 ns/op            1152 B/op         26 allocs/op
    PASS
    ok      github.com/symphony09/ograph/test       14.553s

## 快速开始

### 第一步：声明一个 Node 接口实现

```go
type Person struct {
	ograph.BaseNode
}

func (person *Person) Run(ctx context.Context, state ogcore.State) error {
	fmt.Printf("Hello, i am %s.\n", person.Name())
	return nil
}
```

上面代码中 Person 组合了 BaseNode，并覆写了 Node 接口方法 Run。

### 第二步：构建一个 Pipeline 并运行

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

上面代码在 pipeline 中注册了两个 Person 节点（zhangSan、liSi），并指定 liSi 依赖于 zhangSan。

输出结果

    Hello, i am ZhangSan.
    Hello, i am LiSi.

## 更多示例

更多示例代码，请参考 example 目录下代码。

| 示例文件名                | 示例说明                                 |
| :------------------------ | :--------------------------------------- |
| e01\_hello\_test.go       | 演示基本流程                             |
| e02\_state\_test.go       | 演示如何在节点间分享状态数据             |
| e03\_factory\_test.go     | 演示如何用工厂模式创建节点               |
| e04\_param\_test.go       | 演示如何设置节点参数                     |
| e05\_wrapper\_test.go     | 演示如何使用 `wrapper` 增强节点功能      |
| e06\_cluster\_test.go     | 演示如何使用 `cluster` 灵活调度多个节点  |
| e07\_global\_test.go      | 演示如何全局注册工厂函数                 |
| e08\_virtual\_test.go     | 演示如何使用虚拟节点简化依赖关系         |
| e09\_interrupter\_test.go | 演示如何在`pipeline`运行过程中插入中断   |
| e10\_compose\_test.go     | 演示怎么组合嵌套`pipeline`               |
| e11\_advance\_test.go     | 一些进阶用法，包含图校验、导出，池预热等 |

## 开箱即用

ograph 提供了一些比较通用的节点实现：

| 名称      | 类型     | 作用                         | 文档                             |
| :-------- | :------- | ---------------------------- | :------------------------------- |
| CMD       | 普通节点 | 命令行执行                   | [链接](docs/node_cmd.md)         |
| HttpReq   | 普通节点 | HTTP请求                     | [链接](docs/node_http_req.md)    |
| Choose    | 簇       | 在多个节点中选择一个执行     | 施工中                           |
| Parallel  | 簇       | 并发执行多个节点             | [链接](docs/cluster_parallel.md) |
| Queue     | 簇       | 队列顺序执行多个节点         | [链接](docs/cluster_queue.md)    |
| Race      | 簇       | 多个节点竞争执行             | 施工中                           |
| Async     | 包装器   | 异步执行被包装节点           | [链接](docs/wrapper_async.md)    |
| Condition | 包装器   | 按条件判断是否执行被包装节点 | 施工中                           |
| Debug     | 包装器   | 调试被包装节点               | [链接](docs/wrapper_debug.md)    |
| Delay     | 包装器   | 延迟执行被包装节点           | [链接](docs/wrapper_delay.md)    |
| Loop      | 包装器   | 循环执行被包装节点           | [链接](docs/wrapper_loop.md)     |
| Retry     | 包装器   | 重试失败节点                 | [链接](docs/wrapper_retry.md)    |
| Silent    | 包装器   | 抑制节点报错失败             | [链接](docs/wrapper_silent.md)   |
| Timeout   | 包装器   | 节点超时控制                 | [链接](docs/wrapper_timeout.md)  |
| Trace     | 包装器   | 追踪被包装节点的执行过程     | [链接](docs/wrapper_trace.md)    |


## Q&A

> 导出导入图的限制是什么？

所有节点需要是以工厂方式创建，导入图的 pipeline 需要已注册节点对应的工厂。

> 为什么提供多种节点创建方式（UseNode，UseFactory，UseFn）？

对于简单场景直接注册单例和运行函数比较方便，但要考虑 pipeline 并发执行问题和图导入导出时，就需要使用工厂方式。

> State 存取是并发安全的吗？

默认使用的 state 是并发安全的，但如果是使用了自定义实现则无法保证并发安全。

> 怎么达到最佳性能，有最佳实践吗？

由于协程轻量灵活，一般不用做调整优化，如果节点初始化比较慢可以考虑预热 worker 池。

