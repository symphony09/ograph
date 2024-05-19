# Async Wrapper 异步执行

> 用于异步执行被包装的节点

## 基本使用方式 | Basic Usage

```go
	stopCh := make(chan struct{})

	p := ograph.NewPipeline()

	e := ograph.NewElement("node_to_async_run").
		UseFn(func() error {
			time.Sleep(time.Second)
			fmt.Println("node end")
			stopCh <- struct{}{}
			return nil
		}).
		Wrap(ogimpl.Async)

	// async node wont't block pipeline
	p.Register(e).Run(context.TODO(), nil)

	// The pipeline will end immediately after completion,
	// while the node will end one second after startup.
	fmt.Println("pipeline end")
	<-stopCh
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 示例(Example) |
| :----------- | :------------- | :------------ | :------------ |
| -            | -              | -             | -             |



## 小贴示 | Tips

在没有额外同步机制的情况下，无法保证能获取异步逻辑的执行结果。

为了避免不一致，异步节点对 `state`的读写对其他节点来说是透明的。也就是说，其他节点始终无法观察到异步节点对`state`的修改，而不是某些情况下能观察到，某些情况下观察不到。

如果某一个或者多个节点依赖异步节点的执行结果，应该考虑改为普通节点，或者使用额外的同步机制来传递结果。

Without additional synchronization mechanisms, it is not guaranteed to obtain the result of the execution of asynchronous logic.

To avoid inconsistencies, the read-write operations on `state` by asynchronous nodes are transparent to other nodes. This means that other nodes cannot observe the modifications made by asynchronous nodes to `state`, not in some cases, but in all cases.

If a node or multiple nodes depend on the execution result of an asynchronous node, it is recommended to consider changing to a regular node or using an additional synchronization mechanism to transmit the result.