# Cmd Node 命令行节点

> 用于执行命令行的普通节点，可以跨平台使用

## 基本使用方式 | Basic Usage

```go
	p := ograph.NewPipeline()

	e := ograph.NewElement("cmd").UseFactory(ogimpl.CMD).
		Params("Cmd", []string{"go", "version"})

    // register and run the node which exec go version
	p.Register(e).Run(context.TODO(), nil)
```

## 参数 | Parameter

| 参数名(Name) | 必需(Required) | 含义(Meaning)     | 示例(Example)            |
| :----------- | :------------- | :---------------- | :----------------------- |
| Cmd          | ✔              | cmd to exec       | []string{"go","version"} |
| Env          | ✗              | exec env          | []string{"key=value"}    |
| Dir          | ✗              | working directory | "/root"                  |

## 最佳实践 | Best practice

使用 OGRAPH_ALLOW_CMD_LIST 环境变量限制可执行命令

Use the OGRAPH_ALLOW_CMD_LIST environment variable to limit executable commands

**Liunx**

```bash 
export OGRAPH_ALLOW_CMD_LIST=ls,cat
```