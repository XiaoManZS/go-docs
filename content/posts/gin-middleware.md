+++
date = '2026-11-15T20:40:00+08:00'
draft = false
title = 'Gin 中间件'
tags = ["Go", "Gin", "中间件", "cors"]
categories = ["教程"]
summary = "Gin 中间件：Next、Abort、洋葱模型，手写 CORS 看预检请求，项目里改用 gin-contrib/cors"
weight = 25
+++

# Gin 中间件

什么是中间件?

中间件属于，[真正的业务代码]，之前，以及之后一段的通用逻辑，对请求做统一的加工，拦截，包装等，基本所有的框架都有这个思想。

例如：

- 跨域 CORS
- 登录鉴权、权限校验
- 请求日志、耗时统计
- panic 捕获 & 统一异常返回
- 请求限流、ip 黑名单
- 统一加请求 ID（traceId）

### 注册中间件

中间件，是通过`Use`方法去进行注册的，接受一个函数，并且注入参数`gin.Context`上下文，
>需注意：中间件是放在顶部的，而不能放在路由下面，否则不生效


```go
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors(ctx *gin.Context) {
	//请求之前做点什么
	fmt.Println("==================before")
	ctx.Next()
    //请求之后做点什么
	fmt.Println("==================after")
}

func main() {
	router := gin.New()
	router.Use(Cors)

	router.GET("/ping", func(ctx *gin.Context) {
		fmt.Println("result")
		ctx.String(http.StatusOK, "请求成功:%v", ctx.Request.URL)
	})

    //router.Use(Cors) 如放在路由下面则不生效

	router.Run()
}
```

输出结果

```txt
==================before
result
==================after
```

### Next & Abort

#### Next函数
请求进来时，Gin 把中间件和路由处理函数排成一条链，从头开始调用。遇到 Next()，当前函数先停住，去执行链上还没跑的部分：后面的中间件，以及最后的 GET /ping。这些全部 return 之后，Next() 才返回，after 才会打印

```go
func cors(c *gin.Context) {
    // 👉【前置逻辑】先跑这里
    fmt.Println("before next")

    c.Next() // ✅ 关键点：放行，去跑后面的链条，这里代码会卡住，等待后面全部跑完

    // 👉【后置逻辑】等后面所有中间件、handler 跑完之后，才回来执行这里
    fmt.Println("after next")
}
```

#### Abort

`Abort` 把这条链标记为已中止。还没开始的后续中间件，以及最后的业务函数，都不会再被调用。`/ping` 里的 `result` 不会打印。

它不会让当前函数停在这一行。`Abort()` 下面的语句照常往下走。想跳过当前函数剩下的代码，要自己 `return`。`return` 只离开当前函数，拦住后面链条的是 `Abort`。

已经在外层执行、并且调用过 `Next()` 的中间件也不受影响。内层中止之后，外层的 `Next()` 会返回，写在它下面的后置逻辑仍会执行。

`Abort` 本身不写响应。只调用它时，客户端拿到的是默认的 **200** 和空正文。拦截时先把状态码和内容写出去，或者直接用 `AbortWithStatusJSON`（内部就是先 `Abort`，再 `JSON`）。

```go
func logCost(ctx *gin.Context) {
	fmt.Println("Log before")
	ctx.Next()
	fmt.Println("Log after")
}

func auth(ctx *gin.Context) {
	fmt.Println("before Abort")

	if ctx.GetHeader("token") == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "未登录",
		})
		fmt.Println("after Abort")
		return
	}

	ctx.Next()
}
```

注册顺序是先 `logCost` 再 `auth`。没带 `token` 访问 `/ping` 时：

```txt
Log before
before Abort
after Abort
Log after
```

响应是 `401`，正文 `{"error":"未登录"}`。`Log after` 仍会打印，因为 `logCost` 已经停在自己的 `Next()` 上，等的就是内层返回。`Abort` 拦的是还没轮到的函数。

### 多个中间件执行例子

`Use` 可以连续注册。先注册的包在外面，后注册的在里面。同一次请求先按注册顺序走进去，业务函数返回后，再按相反的顺序走出来。

下面先挂 `Cors`，再挂 `Cors2`。两个路由都写在 `Use` 后面，所以 `/ping` 和 `/test` 都会经过这两个中间件。下面这份输出是访问 `/ping` 的：只有它在处理函数里打印了 `result`。

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors(ctx *gin.Context) {
	//请求之前做点什么
	fmt.Println("==================before")
	ctx.Next()
	fmt.Println("==================after")
}
func Cors2(ctx *gin.Context) {
	//请求之前做点什么
	fmt.Println("==================before2")
	ctx.Next()
	fmt.Println("==================after2")
}

func main() {
	router := gin.New()

	router.Use(Cors)
	router.Use(Cors2)

	router.GET("/ping", func(ctx *gin.Context) {
		fmt.Println("result")
		ctx.String(http.StatusOK, "请求成功:%v", ctx.Request.URL)
	})

	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "请求成功:%v", ctx.Request.URL)
	})

	router.Run()
}
```

访问 `GET /ping` 的输出：

```txt
==================before
==================before2
result
==================after2
==================after
```

套在一起就是：

```txt
Cors before
  Cors2 before
    result
  Cors2 after
Cors after
```

`Cors` 先打印 `before`，调用 `Next()` 后停住，把控制权交给后注册的 `Cors2`。`Cors2` 同样先打印 `before2`，再 `Next()` 进入 `/ping`。`/ping` 返回之后，才轮到 `after2`；接着 `Cors` 的 `Next()` 返回，最后打印 `after`。

谁先 `Use`，谁的前置最早执行，谁的后置最晚执行。把 `Use(Cors)` 和 `Use(Cors2)` 对调，`before` / `after` 的先后也会对调。

访问 `/test` 时，两层中间件的顺序相同，中间没有 `result` 这一行。

### 案例：跨域 CORS

浏览器默认只让页面读**同源**接口的响应。源看三样：协议、主机、端口，三样全相同才算同一个源。

| 页面地址 | 接口地址 | 是否同源 |
| --- | --- | --- |
| `http://127.0.0.1:5500` | `http://127.0.0.1:5500/ping` | 同源 |
| `http://127.0.0.1:5500` | `http://127.0.0.1:8080/ping` | 跨域，端口不同 |
| `http://localhost:8080` | `http://127.0.0.1:8080` | 跨域，`localhost` 和 `127.0.0.1` 是两台「主机」 |
| `https://example.com` | `http://example.com` | 跨域，协议不同 |

上面这段脚本如果跑在 `http://127.0.0.1:5500` 的页面里，去请求 `http://localhost:8080/ping`，就是跨域。服务端不给 CORS 响应头时，接口其实可能已经执行了，但浏览器会把结果拦下，前端 `fetch` 直接报错。

```html
<script>
    fetch("http://localhost:8080/ping", {
        method: "DELETE",
        headers: {
            "Content-Type": "application/json; charset=utf-8"
        }
    })
</script>
```

#### 请求方法

`method` 就是这次请求的方法，和 Gin 里 `router.GET`、`router.DELETE` 是同一组名字：

| 方法 | 一般用途 |
| --- | --- |
| `GET` | 查询 |
| `POST` | 新建、提交 |
| `PUT` / `PATCH` | 更新 |
| `DELETE` | 删除 |
| `OPTIONS` | 浏览器用来问「这个跨域请求允不允许」，不是业务接口 |

这里写的是 `DELETE`，所以服务端要有 `router.DELETE("/ping", ...)`。写成 `router.GET` 时，真正的删除请求对不上，会得到 404。

#### 预检请求

有的跨域请求浏览器会直接发，有的会先发一次 **预检**。预检是浏览器自动补的 `OPTIONS`，页面里的 `fetch` 不用自己写。

这次会预检，因为同时占了两条：

- 方法是 `DELETE`（不是 `GET` / `HEAD` / `POST`）
- 请求头 `Content-Type` 是 `application/json`（不是表单或纯文本）

浏览器实际会发两次：

1. `OPTIONS /ping`，头里带上 `Origin`（页面的源，例如 `http://127.0.0.1:5500`）、`Access-Control-Request-Method: DELETE`、`Access-Control-Request-Headers: content-type`。问服务器：这个源、这个方法、这个头，能不能发。
2. 预检通过之后，才发真正的 `DELETE /ping`。

预检没有对应的业务路由。中间件不接住 `OPTIONS` 的话，Gin 回 404，浏览器认为不允许跨域，第二次 `DELETE` 也不会发，终端里看不到 `result`。

#### 三个响应头

手写的中间件在进业务之前，把允许范围写到响应头上。真正的 `DELETE` 也要带上这些头，所以写在 `OPTIONS` 判断的前面，两条请求都能带上。

| 响应头 | 例子里的值 | 作用 |
| --- | --- | --- |
| `Access-Control-Allow-Origin` | `*` | 允许哪个源读取响应。`*` 表示任意源。页面要带 Cookie 时不能用 `*`，得写成具体源 |
| `Access-Control-Allow-Methods` | `GET,POST,DELETE,PUT,PATCH` | 预检时告诉浏览器，这些方法可以跨域调用。要包含这次的 `DELETE` |
| `Access-Control-Allow-Headers` | `Content-Type` | 预检时告诉浏览器，请求可以带哪些头。前端写了 `Content-Type`，这里就要允许它 |

```go
func Cors(ctx *gin.Context) {
	fmt.Println("==================before")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,PATCH")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type")

	// 预检不是业务请求，回 204 后结束，不再进入 /ping
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
	fmt.Println("==================after")
}
```

`AbortWithStatus(204)` 把预检标成成功（没有正文），并中止后面的链条。这里要 `return`：否则还会继续执行 `Next()`。`Abort` 之后 `Next()` 不会进 `/ping`，但当前函数剩下的代码仍会走，和上一节一样。

完整例子：

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors(ctx *gin.Context) {
	fmt.Println("==================before")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,PATCH")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type")

	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
	fmt.Println("==================after")
}

func main() {
	router := gin.New()
	router.Use(Cors)

	router.DELETE("/ping", func(ctx *gin.Context) {
		fmt.Println("result")
		ctx.String(http.StatusOK, "请求成功:%v", ctx.Request.URL)
	})

	router.Run()
}
```

用另一个源的页面点这次 `fetch`，顺序是：

| 顺序 | 请求 | 终端 | 响应 |
| --- | --- | --- | --- |
| 1 | `OPTIONS /ping` | 只有 `before` | `204`，带上面三个头 |
| 2 | `DELETE /ping` | `before`、`result`、`after` | `200`，正文是「请求成功」，响应里同样有 `Access-Control-Allow-Origin` |

第二次如果缺 `Allow-Origin`，浏览器还是不会把正文交给 `fetch`。

#### 项目里用现成的

上面这份是为了看清预检和三个头，项目里不用自己维护。Gin 官方配套的是 [`github.com/gin-contrib/cors`](https://github.com/gin-contrib/cors)，它会处理 `OPTIONS`、补响应头。

```bash
go get github.com/gin-contrib/cors
go mod tidy
```

和手写 `*` 同样宽松的写法：

```go
import "github.com/gin-contrib/cors"

func main() {
	router := gin.New()
	router.Use(cors.Default())

	router.DELETE("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "请求成功:%v", ctx.Request.URL)
	})

	router.Run()
}
```

`cors.Default()` 允许任意源，效果接近 `Access-Control-Allow-Origin: *`，也不带 Cookie。要收成某一个前端地址时，改用配置。`AllowOrigins` 写页面的源，方法、请求头和手写的那三行是同一件事：

```go
config := cors.DefaultConfig()
config.AllowOrigins = []string{"http://127.0.0.1:5500"}
config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
config.AllowHeaders = []string{"Content-Type"}

router.Use(cors.New(config))
```

`DefaultConfig()` 里其实已经带了这些方法和 `Content-Type`。源必须自己填，漏了的话 `cors.New` 会直接报配置冲突。页面地址是 `localhost` 还是 `127.0.0.1`，要和 `AllowOrigins` 里的字符串一致。