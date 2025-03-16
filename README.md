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

*   灵活的虚节点设置，用以简化依赖关系

### 性能对比

经过 Benchmark 测试，OGraph 性能优于 CGraph。

[CGraph 性能测试参考](http://www.chunel.cn/archives/cgraph-compare-taskflow-v1)

[OGraph 性能测试参考](docs/benchmark_report.md)


|                          | CGraph（基准） | OGraph（本项目）     |
| :----------------------- | :------------- | :------------------- |
| 场景一（无连接32节点）   | 8204 ns/op     | 4308 ns/op（+90.4%） |
| 场景二（串行连接32节点） | 572 ns/op      | 281.7 ns/op（+103%） |
| 场景三（简单DAG 6节点）  | 4042 ns/op     | 2762 ns/op（+46.3%） |
| 场景四（8x8全连接）      | 13450 ns/op    | 8333 ns/op（+61.4%） |


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

## 更多文档

请前往 [https://symphony09.github.io/ograph-docs](https://symphony09.github.io/ograph-docs/zh/docs/quick-start/) 查看更多文档!

