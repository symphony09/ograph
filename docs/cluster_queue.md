# Queue Cluster 串行队列簇

> 用于顺序执行多个节点
> 
> For sequentially executing multiple nodes.

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	eat := ograph.NewElement("eat").UseFn(func() error {
		fmt.Println("start eat")
		time.Sleep(time.Second)
		fmt.Println("end eat")
		return nil
	})

	work := ograph.NewElement("work").UseFn(func() error {
		fmt.Println("start work")
		time.Sleep(time.Second)
		fmt.Println("end work")
		return nil
	})

	e := ograph.NewElement("boy").
		UseFactory(ogimpl.Queue, eat, work)

	// The eat and work nodes will be executed one by one.
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example) |
| :----------- | :------------- | :------------ | ---------- | :------------ |
| -            | -              | -             | -          |
