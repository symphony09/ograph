# Http Req Node HTTP请求节点

> 用于发起 http 请求的节点，用于 webhook 等场景
> 
> Node for initiating HTTP requests, used for scenarios such as webhooks.

## 基本使用方式 | Basic Usage

```go
	state := ograph.NewState()
	state.Set("time", time.Now())

	p := ograph.NewPipeline()

	e := ograph.NewElement("req").UseFactory(ogimpl.HttpReq).
		Params("Method", "POST").
		Params("Url", "http://localhost:8080/ping").
		// Render body as:
		// Send-Time: 2024-08-03 10:10:20.0873372 +0800 CST m=+0.002574001
		Params("BodyTpl", `Send-Time: {{GetState "time"}}`)

	p.Register(e).Run(context.Background(), state)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning) | 类型(Type) | 示例(Example)                   |
| :----------- | :------------- | :------------ | ---------- | :------------------------------ |
| Method       | ✔              | http method   | string     | POST                            |
| Url          | ✔              | http url      | string     | http://localhost:8080/ping      |
| ContentType  | ✗              | content-type  | string     | application/json                |
| Body         | ✗              | request body  | string     | hello                           |
| BodyTpl      | ✗              | body template | string     | hello, i am {{GetState "name"}} |

## 小贴示 | Tips

目前仅支持 GET, POST 两种请求方法。

Currently supports only GET and POST request methods.

使用模板渲染请求体时，可以使用 GetState 函数提取执行状态值。要求这个状态值是可以公开访问的（即 key 需要为 string 类型，而非私有类型）。

When rendering the request body with a template, the GetState function can be used to extract the execution state values. The state value must be publicly accessible (i.e., the key needs to be of string type, not a private type).