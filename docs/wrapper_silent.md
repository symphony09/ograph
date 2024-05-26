# Silent Wrapper 静默执行

> 用于忽略被包装的节点的运行错误
>
>  Ignore running errors of wrapped nodes.

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("problem_node").
		UseFn(func() error {
			return errors.New("something going wrong")
		}).
		Wrap(ogimpl.Silent)

	// Pipeline will not fail due to failed nodes in the problem.
	err := p.Register(e).Run(context.TODO(), nil)
	fmt.Println(err == nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example) |
| :----------- | :------------- | :------------ | ---------- | :------------ |
| -            | -              | -             | -          | -             |

## 小贴示 | Tips

节点报错会直接导致 pipeline 停止运行并返回错误，所以有时需要忽略非关键节点报错，报错信息仍然可以通过日志追踪到。

Nodes reporting errors will directly cause the pipeline to stop running and return an error. Therefore, sometimes it is necessary to ignore non-critical node errors, and error information can still be tracked through logs.