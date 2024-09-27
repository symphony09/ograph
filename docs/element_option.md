# Element Option

## 使用方式

```go
	flash := ograph.NewElement("Flash").UseNode(&Sloth{}).
		Apply(
			ogimpl.DelayOp(time.Second),
			ogimpl.LoopOp(3),
			ogimpl.TimeoutOp(time.Second*3),
			ogimpl.RetryOp(2),
		)
```

以上代码等同于：

```go
	flash := ograph.NewElement("Flash").UseNode(&Sloth{}).
		Wrap(ogimpl.Delay).Params("Wait", time.Second).
		Wrap(ogimpl.Loop).Params("LoopTimes", 3).
		Wrap(ogimpl.Timeout).Params("Timeout", time.Second*3).
		Wrap(ogimpl.Retry).Params("MaxRetryTimes", 2)
```

都是为元素添加多个功能，并设置相应参数

但是使用 option 更加简单，也避免了错误设置参数名