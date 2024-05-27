# Trace Wrapper 追踪执行

> 用于把被包装节点的执行过程记录在日志中
>
> It is used to record the execution process of the wrapped node in the log.

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("node_to_be_trace").
		UseNode(&ograph.BaseNode{}).
		Wrap(ogimpl.Trace)

	// The execution process of the node will be recorded in the log.
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example) |
| :----------- | :------------- | :------------ | ---------- | :------------ |
| -            | -              | -             | -          | -             |


## 小贴示 | Tips

可以在 state 中传入 id 用于关联某次执行过程。

You can pass the `id` into the `state` to associate a specific execution process.

```go
	// ...  Declare pipeline and element

	state := ograph.NewState()
	ogimpl.SetTraceId(state, "your_trace_id")

	p.Register(e).Run(context.TODO(), state)
```

