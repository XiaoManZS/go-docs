+++
date = '2026-11-15T20:40:00+08:00'
draft = false
title = 'Go Http标准库'
tags = ["Go", "http", "网络", "network"]
categories = ["教程"]
summary = "Go 自带 net/http：ServeMux、HandleFunc、Write 返回字符串、返回 JSON、Query 参数查询"
weight = 22
+++

# http

Go 自带 `net/http`，不用额外框架，就能写 HTTP 服务。底层每个请求一个 goroutine，默认并发很强。

本篇按动手顺序来：

1. `http.NewServeMux` + `HandleFunc` 注册路由（重点：第一个参数 = 请求方式 + 路径）  
2. 启动服务  
3. `w.Write` 返回字符串（以及 `[]byte` 转换）  
4. 返回 JSON：`json.NewEncoder(w).Encode(...)`  
5. 带查询参数的用户列表搜索  

---

## 1. 创建路由：NewServeMux + HandleFunc

`ServeMux` 是路由器：根据**请求方式**和**路径**，把请求交给对应处理函数。

```go
mux := http.NewServeMux()
mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
	// 处理 /hello
})
```

| API | 作用 |
| --- | --- |
| `http.NewServeMux()` | 新建一个路由器 |
| `mux.HandleFunc(pattern, 处理函数)` | 用第一个参数 `pattern` 匹配请求，命中则执行处理函数 |

处理函数签名固定：

```go
func(w http.ResponseWriter, r *http.Request)
```

- `w`：用来**写响应**（状态码、Header、Body）  
- `r`：用来**读请求**（方法、路径、Query、Body）

> 也可以用包级函数 `http.HandleFunc(...)`，它挂在默认的 `DefaultServeMux` 上。自己 `NewServeMux()` 更清晰，推荐项目里用独立 mux。

### HandleFunc 的第一个参数：pattern

签名可以看成：

```go
mux.HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
```

**第一个参数 `pattern` 不是随便写的字符串**，它描述「什么样的请求才进这个处理函数」。

Go 1.22 起，`pattern` 的常见写法是：

```text
[请求方式 ]路径
```

中间有一个空格：前面是 HTTP 方法（可省略），后面是路径。

#### 只写路径（不限制方法）

```go
mux.HandleFunc("/hello", handler)
```

含义：只要路径是 `/hello`，**不管**是 GET、POST 还是 DELETE，都会进这个函数。

适合入门演示；正式接口更建议把方法写上，避免误匹配。

#### 请求方式 + 路径（推荐）

```go
mux.HandleFunc("GET /hello", handler)   // 只匹配 GET /hello
mux.HandleFunc("POST /hello", handler)  // 只匹配 POST /hello
mux.HandleFunc("PUT /user", handler)
mux.HandleFunc("DELETE /user", handler)
```

| pattern | 匹配什么 |
| --- | --- |
| `"/hello"` | 任意方法 + 路径 `/hello` |
| `"GET /hello"` | 仅 GET（以及 HEAD）+ `/hello` |
| `"POST /hello"` | 仅 POST + `/hello` |
| `"GET /user/{id}"` | GET，路径形如 `/user/1`，可用 `r.PathValue("id")` 取 `"1"` |

常见请求方式：

| 方法 | 一般用途 |
| --- | --- |
| `GET` | 查数据（浏览器地址栏、`<a>` 默认就是 GET） |
| `POST` | 新建 / 提交表单 |
| `PUT` / `PATCH` | 更新 |
| `DELETE` | 删除 |

同一个路径、不同方法，可以挂不同处理函数：

```go
mux.HandleFunc("GET /user", listUser)    // 查列表
mux.HandleFunc("POST /user", createUser) // 新建
```

若只注册了 `GET /user`，却用 POST 访问，ServeMux 会自动回 **405 Method Not Allowed**（方法不对），不用你自己先判断。

#### 路径里的通配（可选）

```go
mux.HandleFunc("GET /user/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // /user/1 → "1"
	fmt.Fprint(w, "用户 id = ", id)
})
```

注意和 Query 的区别：

| 写法 | 例子 | 怎么取 |
| --- | --- | --- |
| 路径参数 | `/user/1` | `r.PathValue("id")`（pattern 里要有 `{id}`） |
| 查询参数 | `/user/list?id=1` | `r.URL.Query().Get("id")` |

后面搜索案例用的是 **Query**；路径参数知道有这回事即可。

#### 一眼对照

```go
mux := http.NewServeMux()

// 第一个参数 = 请求方式 + 空格 + 路径
mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello get"))
})
mux.HandleFunc("POST /hello", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello post"))
})
```

浏览器打开 `http://localhost:8080/hello` 走的是 **GET**；用 Postman / curl 发 POST 才会进第二个处理函数。

> 需要 Go **1.22+**。`go.mod` 里 `go 1.22`（或更高）才支持 `"GET /path"` 这种写法；更老的版本把 `"GET /path"` 会理解错，路由对不上。

---

## 2. 创建并启动服务

把 mux 交给 `http.ListenAndServe`：

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})

	fmt.Println("服务启动：http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
```

- `":8080"`：监听本机 8080 端口  
- 第二个参数传 `mux`；若传 `nil`，则用默认的 `DefaultServeMux`  
- 浏览器访问：`http://localhost:8080/hello`

---

## 3. 返回字符串：Write 与 []byte 转换

`ResponseWriter.Write` 的签名是：

```go
Write([]byte) (int, error)
```

它只收 **`[]byte`**，不直接收 `string`。所以要先转换：

```go
mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
	msg := "你好，小满"
	w.Write([]byte(msg)) // string → []byte
})
```

`[]byte(字符串)` 是类型转换（有人叫它「语法糖」）：把字符串变成字节切片，内存里就是那段 UTF-8 字节。

反过来：

```go
b := []byte("abc")
s := string(b) // []byte → string
```

更省事的写法（内部一样会写到 `w`）：

```go
fmt.Fprint(w, "你好，小满")
io.WriteString(w, "你好，小满")
```

入门先记住：`Write` 要 `[]byte`，字符串用 `[]byte("...")` 包一下。

---

## 4. 返回 JSON

接口常见做法：设 `Content-Type`，再用 `json.NewEncoder(w).Encode(...)` 直接往响应里写 JSON。

注意大小写（你草稿里的写法对照）：

| 你写的 | 正确写法 |
| --- | --- |
| `json.newEncode(w).encode()` | `json.NewEncoder(w).Encode(...)` |
| `type User stauct` | `type User struct` |

完整例子：

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		u := User{Id: 1, Name: "小满"}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("服务启动：http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

访问 `http://localhost:8080/user`，响应大致是：

```json
{"id":1,"name":"小满"}
```

拆开看：

1. `` `json:"id"` ``：JSON 字段名用小写 `id`（字段本身仍要大写才能导出）  
2. `w.Header().Set("Content-Type", ...)`：告诉客户端这是 JSON  
3. `json.NewEncoder(w)`：创建一个往 `w` 写的编码器  
4. `.Encode(u)`：序列化并写入响应（末尾会自动加换行）

也可以先 `json.Marshal` 再 `w.Write`，效果类似；`NewEncoder(w).Encode` 少一次中间变量，写接口更常见。

---

## 5. 案例：按 id 搜索用户列表

目标地址：

```text
http://localhost:8080/user/list?id=1
```

你草稿里的几处要改：

| 你写的 | 问题 | 正确写法 |
| --- | --- | --- |
| `r.URL.Query.Get("id")` | `Query` 是**方法**，要加 `()` | `r.URL.Query().Get("id")` |
| `item.Id == targetId` | `Get` 得到的是 **string**，`Id` 是 **int**，不能直接比 | 先 `strconv.Atoi` |
| `stauct` | 拼写错误 | `struct` |

完整可运行示例：

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var userList = []User{
	{Id: 1, Name: "小满"},
	{Id: 2, Name: "张三"},
	{Id: 3, Name: "李四"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/user/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		idStr := r.URL.Query().Get("id") // ?id=1 → "1"

		// 没传 id：返回全部
		if idStr == "" {
			json.NewEncoder(w).Encode(userList)
			return
		}

		targetId, err := strconv.Atoi(idStr) // string → int
		if err != nil {
			http.Error(w, "id 必须是数字", http.StatusBadRequest)
			return
		}

		result := make([]User, 0)
		for _, item := range userList {
			if item.Id == targetId {
				result = append(result, item)
			}
		}

		json.NewEncoder(w).Encode(result)
	})

	fmt.Println("服务启动：http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

试一下：

| 请求 | 结果 |
| --- | --- |
| `/user/list` | 全部用户 |
| `/user/list?id=1` | `[{"id":1,"name":"小满"}]` |
| `/user/list?id=99` | `[]`（没有匹配） |
| `/user/list?id=abc` | 400，提示 id 必须是数字 |

`r.URL.Query()` 返回 `url.Values`（本质是 `map[string][]string`），`.Get("id")` 取第一个值；没有该参数时返回空字符串 `""`。

---

## 小结

| 主题 | 记住什么 |
| --- | --- |
| 路由 | `NewServeMux` + `HandleFunc`；第一个参数是 `"[METHOD ]路径"`，如 `"GET /hello"` |
| 启动 | `ListenAndServe(":8080", mux)` |
| 写文本 | `w.Write([]byte("..."))` |
| 写 JSON | `Header` 设 Content-Type + `json.NewEncoder(w).Encode(v)` |
| 结构体标签 | `` Id int `json:"id"` ``，字段要导出 |
| Query | `r.URL.Query().Get("id")`，和 int 比较前先 `Atoi` |

最小流程：

> **mux 注册路径 → ListenAndServe 启动 → 处理函数里读 `r`、写 `w`。**
