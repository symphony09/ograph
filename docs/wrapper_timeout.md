# Timeout Wrapper 超时限制执行

> 用于超时取消被包装的节点
>
> Cancel wrapped nodes that exceed a timeout.

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("slow_node").
		UseFn(func() error {
			time.Sleep(time.Minute)
			return nil
		}).
		Wrap(ogimpl.Timeout).
		Params("Timeout", "10ms")

	// After a node times out, the pipeline stops running and returns a timeout error.
	err := p.Register(e).Run(context.TODO(), nil)
	fmt.Println(errors.Is(err, ogimpl.ErrTimeout))
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type)              | 示例(Example)       |
| :----------- | :------------- | :------------ | ----------------------- | :------------------ |
| Timeout      | ✔              | 超时时间      | string<br>time.Duration | "1s"<br>time.Second |


Timeout 支持两种类型的参数

## 小贴示 | Tips

Timeout Wrapper 并不能控制节点停止运行和释放资源，只是通过 ctx 传递取消信号。同时向上报告超时错误，使 pipeline 可以不再等待节点运行结束。

The Timeout Wrapper does not control node termination and resource release, but instead passes a cancel signal through ctx. It also reports a timeout error to allow the pipeline to proceed without waiting for the node.

如果不希望超时错误影响 pipeline 继续执行，可以配合 Silent Wrapper 一起使用。

To avoid allowing timeout errors to affect the pipeline's continued execution, you can use the Silent Wrapper in conjunction with the Timeout Wrapper.