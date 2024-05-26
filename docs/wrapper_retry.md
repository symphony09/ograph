# Retry Wrapper 错误重试

> 用于重试运行错误的节点
>
> For retrying failed nodes.

## 基本使用方式 | Basic Usage

```go
	var once sync.Once

	p := ograph.NewPipeline()

	e := ograph.NewElement("problem_node").
		UseFn(func() error {
			var err error

			fmt.Println("node start running")

			once.Do(func() { // Returns error only once.
				err = errors.New("something going wrong")
			})

			return err
		}).
		Wrap(ogimpl.Retry)

	// The pipeline completes normally after the retry fails.
	err := p.Register(e).Run(context.TODO(), nil)
	fmt.Println(err == nil)
```

## 参数 | Parameter

| 参数名(Name)  | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example) |
| :------------ | :------------- | :------------ | ---------- | :------------ |
| MaxRetryTimes | ✗              | 最大重试次数  | int        | 3             |

如果 MaxRetryTimes 小于或等于 0，则使用默认值 1。

If MaxRetryTimes is less than or equal to 0, use the default value of 1.
