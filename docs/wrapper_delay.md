# Delay Wrapper 延迟执行

> 用于延迟执行被包装的节点

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("node_to_delay").
		UseFn(func() error {
			fmt.Println("node start running")
			return nil
		}).
		Wrap(ogimpl.Delay).
		Params("Wait", "3s")

	// node_to_delay will run after 3 seconds
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 示例(Example) |
| :----------- | :------------- | :------------ | :------------ |
| Wait         | ✗              | wait duration | "1h59m59s"    |
| Until        | ✗              | wait until    | time.Time{}   |


## 小贴示 | Tips

可以同时设置 Wait 和 Until 参数，都满足时执行被包装节点

You can set both Wait and Until parameters simultaneously, and execute the wrapped node when both conditions are met.