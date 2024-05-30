# Parallel Cluster 并行簇

> 用于并行执行多个节点
> 
> For parallel execution of multiple nodes.

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

	e := ograph.NewElement("busy_boy").
		UseFactory(ogimpl.Parallel, eat, work)

	// The eat and work nodes will be executed in parallel.
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example) |
| :----------- | :------------- | :------------ | ---------- | :------------ |
| -            | -              | -             | -          |
