# Loop Wrapper 循环执行

> 用于循环执行被包装的节点
>
>  Used for looping execution of wrapped nodes.

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("node_to_loop").
		UseFn(func() error {
			fmt.Println("node start running")
			return nil
		}).
		Wrap(ogimpl.Loop).
		Params("LoopTimes", 3)

	// node_to_loop will loop run 3 times
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type)              | 示例(Example)       |
| :----------- | :------------- | :------------ | ----------------------- | :------------------ |
| LoopTimes    | ✔              | 循环次数      | int                     | 3                   |
| LoopInterval | ✗              | 循环间隔      | string<br>time.Duration | "1s"<br>time.Second |


LoopInterval 支持两种类型的参数
