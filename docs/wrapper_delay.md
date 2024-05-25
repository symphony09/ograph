# Delay Wrapper 延迟执行

> 用于延迟执行被包装的节点
>
> Used for delaying the execution of wrapped nodes.

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

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type)              | 示例(Example)                                      |
| :----------- | :------------- | :------------ | ----------------------- | :------------------------------------------------- |
| Wait         | ✗              | wait duration | string<br>time.Duration | "1h59m59s"<br>time.Second                          |
| Until        | ✗              | wait until    | string<br>time.Time     | "2024-05-25T10:44:57.6061908+08:00"<br>time.Time{} |

Wait 和 Until 支持两种类型的参数，使用 string 类型作为 Until 参数时，需要符合 time.RFC3339Nano 格式。

"Wait" and "Until" support two types of parameters. When using the "Until" parameter with a string type, it must adhere to the time.RFC3339Nano format.


## 小贴示 | Tips

可以同时设置 Wait 和 Until 参数，都满足时执行被包装节点

You can set both Wait and Until parameters simultaneously, and execute the wrapped node when both conditions are met.