+++
date = '2026-11-15T20:40:00+08:00'
draft = false
title = 'Gin Web 框架'
tags = ["Go", "Gin", "http", "网络"]
categories = ["教程"]
summary = "安装 Gin、把 Go 升到最新（或按版本另装），用 Air 热更新，以及路径参数、Query、表单和 JSON 传参"
weight = 23
+++

# Gin

Gin 是 Go 里常用的 HTTP 框架。路由按请求方法分开注册，取参数、回 JSON 都比 `net/http` 少写几行。它底层仍是标准库，上一篇里「读请求、写响应」的概念这里照样用。

本篇按动手顺序来：

1. 把 Go 升到最新稳定版（旧版和 SSE 依赖有兼容问题），或按版本另装一份  
2. 安装 Gin  
3. 安装热更新工具 Air，并看懂那条 `go install`  
4. 起一个最小服务  
5. 传参：路径参数、Query、表单、JSON  

中间件、路由分组这些先不讲。

---

## 1. Go 版本：建议升到最新

先看本机版本：

```bash
go version
```

Gin 会依赖 [`github.com/gin-contrib/sse`](https://github.com/gin-contrib/sse)（Server-Sent Events，服务端按事件往客户端推数据）。这里有两层问题：

- 新版 `sse` 提高了 Go 语言版本要求。Go 太旧时，`go get` 最新 Gin 会在拉依赖阶段失败。  
- 旧版 `sse` 跟较新的 `net/http` 在刷新响应（Flush）上行为不一致，真正用到推流时容易出兼容问题。  

当前最新稳定版是 **Go 1.27.4**。直接覆盖安装最省事：打开 [https://go.dev/dl/](https://go.dev/dl/) 下载对应系统的安装包。Windows 也可以：

```powershell
winget upgrade GoLang.Go
```

装完再执行 `go version`，应看到 `go1.27.4`。

### 不想替换系统里的 go：按版本另装

机器上已经有一个旧的 `go`，又不想动它，可以再装一份指定版本，两条命令：

```bash
go install golang.org/dl/go1.27.4@latest
go1.27.4 download
```

| 命令 | 做了什么 |
| --- | --- |
| `go install golang.org/dl/go1.27.4@latest` | 只装一个名叫 `go1.27.4` 的小命令，放到 `GOPATH/bin`，这时还没有完整工具链 |
| `go1.27.4 download` | 用上面那个命令，把 Go 1.27.4 本体下载下来 |

之后在这个项目里把 `go` 换成 `go1.27.4`，例如 `go1.27.4 mod init`、`go1.27.4 get ...`、`go1.27.4 run .`。原来的 `go version` 不变。

---

## 2. 安装 Gin

在模块目录里执行（还没有模块就先 `go mod init example.com/gin-demo`）：

```bash
go get github.com/gin-gonic/gin@latest
```

`@latest` 拉当前最新稳定版。写作时是 **v1.12.0**，`go.mod` 里要求 **Go 1.25+**，所以前面才先升级。

代码里加上导入，再整理一次依赖：

```go
import "github.com/gin-gonic/gin"
```

```bash
go mod tidy
```

国内拉模块超时的话，先配代理（Modules 那篇写过）：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

用了上一节的并存版本时，这几条里的 `go` 同样换成 `go1.27.4`。

---

## 3. 热更新：Air

开发时每次改完都要自己停进程、再 `go run`，比较烦。Air 就是干这个的：盯着项目里的源文件，一保存就重新编译并重启进程。

安装命令：

```bash
go install github.com/air-verse/air@latest
```

把它拆开看：

| 片段 | 含义 |
| --- | --- |
| `go install` | 编译这个程序，把可执行文件装进 `GOPATH/bin`。它**不会**写进当前项目的 `go.mod`，因为 Air 是你本机的开发工具，不是业务依赖 |
| `github.com/air-verse/air` | 模块路径。早期仓库在 `github.com/cosmtrek/air`，已经迁到 `air-verse`，安装用新路径 |
| `@latest` | 装最新版。Air 本身要求 Go 1.25+，所以同样用刚升级的那套 Go |

装好后进入项目目录：

```bash
air
```

它会按默认配置监视 `.go` 文件。想改监视规则再执行 `air init`，会在当前目录生成 `.air.toml`。

记住它的边界：Air 只在**你自己电脑上写代码**时用。上线还是正常 `go build` 出一个二进制，不会在生产环境里靠它热重启。

Windows 上如果提示找不到 `air`，把 `go env GOPATH` 下面的 `bin` 加进 PATH（常见是 `C:\Users\你的用户名\go\bin`），新开一个终端再执行。并存安装时，用 `go1.27.4 install github.com/air-verse/air@latest`，装出来的命令同样在那个 `bin` 里。

---

## 4. 最小服务

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好")
	})

	if err := router.Run(); err != nil {
		panic(err)
	}
}
```

| API | 作用 |
| --- | --- |
| `gin.Default()` | 创建一个引擎，自带访问日志和 panic 恢复 |
| `router.GET(路径, 处理函数)` | 只接收 GET。POST 就写成 `router.POST` |
| `ctx.String(状态码, 格式, 参数...)` | 往响应里写一段文本，格式和 `fmt.Sprintf` 一样 |
| `router.Run()` | 监听 `:8080`。写成 `router.Run(":8080")` 是一回事 |

`*gin.Context` 同时负责读这次请求、写这次响应，对应标准库里的 `r` 和 `w`。

`go run .` 或 `air` 启动后，浏览器打开 `http://localhost:8080`，能看到「你好」。端口被占用时 `Run` 会返回错误，这里直接 `panic` 让终端把原因打出来。

---

## 5. 传参

接口里数据通常从四个地方来：

| 方式 | 长什么样 | 怎么取 |
| --- | --- | --- |
| 路径参数 | `/user/12` | `ctx.Param("id")` |
| Query | `/user?id=1&name=小满` | `ctx.DefaultQuery` / `ctx.Query` |
| 表单 | POST，正文是 `id=1&name=小满` | `ctx.DefaultPostForm` |
| JSON | POST，正文是 `{"name":"小满","age":18}` | `ctx.ShouldBindJSON` |

下面每种先看一小段，最后合成一个能跑的 `main.go`。

### 路径参数

路径里用 `:名字` 占一位：

```go
router.GET("/user/:id", func(ctx *gin.Context) {
	id := ctx.Param("id") // /user/12 → "12"
	ctx.String(http.StatusOK, "id=%s", id)
})
```

`Param` 拿到的永远是字符串。请求是 `/user/12`，`id` 就是 `"12"`，不会自动变成数字。

### Query

问号后面的键值对，用在「筛选、搜索」这种可选条件上：

```go
router.GET("/user", func(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "none")
	name := ctx.DefaultQuery("name", "none")
	ctx.String(http.StatusOK, "id=%s name=%s", id, name)
})
```

`DefaultQuery` 的第二个参数是**没带这个键**时的默认值。`/user` 什么都不带，两个都是 `none`；`/user?id=1&name=小满` 则取到实际的值。

只要键出现了，哪怕值是空的（`?name=`），也不会再用默认值。想要「空字符串就是没有」，用 `ctx.Query("name")`，缺省时它返回 `""`。

`/user/:id` 和 `/user` 可以同时注册：一个要路径上多一段，一个就是 `/user` 本身，Gin 能分开。

### 表单

POST 的正文是表单（`application/x-www-form-urlencoded` 或 `multipart/form-data`）时，用 `DefaultPostForm`。第二个参数同样是「没带这个字段」时的默认值：

```go
router.POST("/user/formdata", func(ctx *gin.Context) {
	id := ctx.DefaultPostForm("id", "none")
	name := ctx.DefaultPostForm("name", "none")
	ctx.String(http.StatusOK, "id=%s name=%s", id, name)
})
```

它读的是 **Body**，不是 URL 上的 `?id=1`。用 Query 那套函数去读表单，会一直拿到默认值。

### JSON

定义一个结构体，字段要大写才能被填上。`` `json` `` 决定 JSON 里的名字，`` `binding` `` 决定要不要校验：

```go
type User struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age"`
}
```

标签是 `binding`，少写一个字母就不会校验，请求照样能进来。

```go
router.POST("/user/create", func(ctx *gin.Context) {
	var user User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    1,
		"data":    user,
		"success": true,
	})
})
```

`ShouldBindJSON` 做两件事：按 `json` 标签把正文填进结构体，再按 `binding` 标签检查。`Name` 标了 `required`，缺 `name` 或是空字符串都会返回错误。`Age` 没标，不传就是 `0`。

失败时要 `return`。否则函数还会继续往下走，同一次响应里先写 400、再写 201。

`gin.H` 是 `map[string]any` 的别名，用来随手组一个 JSON 对象。`ctx.JSON` 会设置 `Content-Type: application/json` 并序列化。这里成功用 `201 Created`，参数不合法用 `400 Bad Request`。

请求需要：

- 方法：`POST`  
- 头：`Content-Type: application/json`  
- 正文：`{"name":"小满","age":18}`  

### 合在一起

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age"`
}

func main() {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好")
	})

	// 路径参数：/user/12
	router.GET("/user/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		ctx.String(http.StatusOK, "id=%s", id)
	})

	// Query：/user?id=1&name=小满
	router.GET("/user", func(ctx *gin.Context) {
		id := ctx.DefaultQuery("id", "none")
		name := ctx.DefaultQuery("name", "none")
		ctx.String(http.StatusOK, "id=%s name=%s", id, name)
	})

	// 表单正文：id=1&name=小满
	router.POST("/user/formdata", func(ctx *gin.Context) {
		id := ctx.DefaultPostForm("id", "none")
		name := ctx.DefaultPostForm("name", "none")
		ctx.String(http.StatusOK, "id=%s name=%s", id, name)
	})

	// JSON 正文：{"name":"小满","age":18}
	router.POST("/user/create", func(ctx *gin.Context) {
		var user User
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{
			"code":    1,
			"data":    user,
			"success": true,
		})
	})

	if err := router.Run(); err != nil {
		panic(err)
	}
}
```

用 `curl` 对一下（服务已在 8080）：

```bash
curl http://localhost:8080/
curl http://localhost:8080/user/12
curl "http://localhost:8080/user?id=1&name=小满"
curl -X POST http://localhost:8080/user/formdata -d "id=1&name=小满"
curl -X POST http://localhost:8080/user/create \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"小满\",\"age\":18}"
```

| 请求 | 结果 |
| --- | --- |
| `GET /` | `你好` |
| `GET /user/12` | `id=12` |
| `GET /user?id=1&name=小满` | `id=1 name=小满` |
| `GET /user` | `id=none name=none` |
| `POST /user/formdata`，正文 `id=1&name=小满` | `id=1 name=小满` |
| `POST /user/create`，正文 `{"name":"小满","age":18}` | `201`，`data` 里是这个用户 |
| `POST /user/create`，正文 `{}` | `400`，提示 `name` 为必填 |

---

## 小结

| 主题 | 记住什么 |
| --- | --- |
| Go 版本 | 升到最新稳定版（目前 1.27.4）。旧 Go 和新版 `gin-contrib/sse` 对不上；不想覆盖安装就 `go install golang.org/dl/go1.27.4@latest` 再 `download` |
| 装 Gin | `go get github.com/gin-gonic/gin@latest`，然后 `go mod tidy` |
| Air | `go install github.com/air-verse/air@latest`，项目目录执行 `air`。它监视文件并重启进程，只用于本地开发 |
| 启动 | `gin.Default()` 注册路由，`router.Run()` 监听 `:8080` |
| 路径参数 | `/user/:id`，用 `ctx.Param("id")`，结果是字符串 |
| Query | `ctx.DefaultQuery("id", "none")`，第二个参数是没带这个键时的默认值 |
| 表单 | `ctx.DefaultPostForm`，读的是 POST 正文，不是 URL |
| JSON | 结构体加 `json` / `binding` 标签，`ShouldBindJSON` 绑定；失败要 `return`，再用 `ctx.JSON` 写响应 |
