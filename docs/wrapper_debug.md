# Debug Wrapper 调试执行

> 用于把被包装节点的状态调试记录在日志中
>
> The state of the wrapped node can be debugged and recorded in the log.

## 基本使用方式 | Basic Usage

```go
type ReadWriteStateNode struct {
	ograph.BaseNode
}

func (node *ReadWriteStateNode) Run(ctx context.Context, state ogcore.State) error {
	state.Set("i", 1)
	state.Update("i", func(val any) any {
		return "1"
	})
	state.Get("i")
	return nil
}

```

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("node_to_debug").
		UseNode(&ReadWriteStateNode{}).
		Wrap(ogimpl.Debug)

	// The state of a node before and after reading and writing will be recorded in the log.
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
