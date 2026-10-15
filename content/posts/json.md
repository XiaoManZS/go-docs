+++
date = '2026-10-15T23:52:00+08:00'
draft = false
title = 'Go JSON 序列化与反序列化'
tags = ["Go", "JSON", "序列化", "encoding/json"]
categories = ["教程"]
summary = "详解 Go encoding/json：Marshal / Unmarshal、结构体标签、omitempty，以及常见坑"
weight = 18
+++

# JSON 序列化与反序列化

写接口、存配置、跟前端或别的服务交换数据，几乎都会碰到 **JSON**。  
Go 标准库 **`encoding/json`** 负责两件事：

- **序列化（Marshal）**：把 Go 的值变成 JSON 字节（`[]byte`）
- **反序列化（Unmarshal）**：把 JSON 字节解析回 Go 的值

入门先记住五件事：

1. 序列化用 `json.Marshal`，反序列化用 `json.Unmarshal`。
2. 结构体字段必须**导出**（首字母大写），否则 JSON 里看不到。
3. 用 `` `json:"字段名"` `` 标签控制 JSON 里的名字。
4. `omitempty` 可以在「空值」时省略该字段。
5. 反序列化时第二个参数必须是**指针**。

> 口诀：出站 `Marshal`，进站 `Unmarshal`；字段要大写，名字靠 tag。

## 序列化：Marshal

把结构体、map、切片等变成 JSON：

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	p := Person{Name: "小满", Age: 18, City: "杭州"}

	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// {"Name":"小满","Age":18,"City":"杭州"}
}
```

| 写法 | 含义 |
| --- | --- |
| `json.Marshal(v)` | 返回 `[]byte` 和 `error` |
| `string(data)` | 把字节转成可读字符串 |
| 字段名 | 默认用结构体字段名（如 `Name`） |

想格式化（缩进、好看一点）用 `MarshalIndent`：

```go
data, err := json.MarshalIndent(p, "", "  ")
if err != nil {
	panic(err)
}
fmt.Println(string(data))
```

输出类似：

```json
{
  "Name": "小满",
  "Age": 18,
  "City": "杭州"
}
```

## 反序列化：Unmarshal

把 JSON 填进结构体。**第二个参数必须是指针**：

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name string
	Age  int
	City string
}

func main() {
	raw := []byte(`{"Name":"小满","Age":18,"City":"杭州"}`)

	var p Person
	err := json.Unmarshal(raw, &p) // 注意 &p
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", p)
	// {Name:小满 Age:18 City:杭州}
}
```

| 写法 | 含义 |
| --- | --- |
| `json.Unmarshal(data, &v)` | 把 JSON 写入 `v` |
| 必须传指针 | 传值无法修改外部变量 |
| JSON 多了字段 | 默认忽略，不报错 |
| JSON 少了字段 | 对应字段保持零值 |

> 忘了写 `&` 会报错：`json: Unmarshal(non-pointer ...)`。

## 结构体标签（json tag）

接口里常见小写字段名（`name`、`age`），而 Go 导出字段必须大写。用 tag 对齐两边：

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func main() {
	p := Person{Name: "小满", Age: 18, City: "杭州"}

	data, _ := json.Marshal(p)
	fmt.Println(string(data))
	// {"name":"小满","age":18,"city":"杭州"}

	var p2 Person
	_ = json.Unmarshal([]byte(`{"name":"奶茶","age":20,"city":"上海"}`), &p2)
	fmt.Printf("%+v\n", p2)
	// {Name:奶茶 Age:20 City:上海}
}
```

常用 tag 选项：

| tag | 作用 |
| --- | --- |
| `` `json:"name"` `` | JSON 字段名用 `name` |
| `` `json:"name,omitempty"` `` | 空值时不输出该字段 |
| `` `json:"-"` `` | 序列化/反序列化都忽略 |
| `` `json:",omitempty"` `` | 保留字段名，只加 omitempty |

### omitempty

「空」的含义：`0`、`""`、`false`、`nil`、空切片/map 等零值。

```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name   string   `json:"name"`
	Age    int      `json:"age,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Secret string   `json:"-"` // 永远不出现在 JSON 里
}

func main() {
	u := User{Name: "小满", Secret: "不要对外暴露"}
	data, _ := json.Marshal(u)
	fmt.Println(string(data))
	// {"name":"小满"}
	// Age=0、Tags=nil 被 omitempty 省略；Secret 被 - 忽略
}
```

## 嵌套结构体

JSON 里经常有对象套对象，对应 Go 的嵌套结构体：

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Address struct {
	City    string `json:"city"`
	Street  string `json:"street"`
}

type Order struct {
	ID      int     `json:"id"`
	Product string  `json:"product"`
	Addr    Address `json:"address"`
}

func main() {
	o := Order{
		ID:      1001,
		Product: "芋泥波波",
		Addr:    Address{City: "杭州", Street: "西湖路"},
	}

	data, _ := json.Marshal(o)
	fmt.Println(string(data))
	// {"id":1001,"product":"芋泥波波","address":{"city":"杭州","street":"西湖路"}}

	raw := []byte(`{"id":1002,"product":"绿茶","address":{"city":"上海","street":"南京路"}}`)
	var o2 Order
	_ = json.Unmarshal(raw, &o2)
	fmt.Printf("%+v\n", o2)
}
```

## map 与切片

不一定非要结构体，map 和切片也能直接编解码：

```go
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// map → JSON
	m := map[string]any{
		"name": "小满",
		"age":  18,
		"ok":   true,
	}
	data, _ := json.Marshal(m)
	fmt.Println(string(data))

	// JSON → map
	var m2 map[string]any
	_ = json.Unmarshal(data, &m2)
	fmt.Println(m2["name"], m2["age"])
	// 注意：JSON 数字默认解成 float64

	// 切片
	list := []string{"芋泥波波", "绿茶", "红茶"}
	b, _ := json.Marshal(list)
	fmt.Println(string(b)) // ["芋泥波波","绿茶","红茶"]
}
```

> `map[string]any` 很灵活，但类型要自己断言。业务数据优先用结构体，类型更安全。

数字反序列化进 `any` 时是 **`float64`**，不是 `int`：

```go
var m map[string]any
_ = json.Unmarshal([]byte(`{"age":18}`), &m)
age := m["age"].(float64) // 对
// age := m["age"].(int)  // panic：类型不对
```

## Encoder / Decoder（流式）

跟文件、网络连接打交道时，不必先读成整段 `[]byte`，可以用流式接口：

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Encoder：往 Writer 写 JSON
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(Person{Name: "小满", Age: 18})
	// 输出一行 JSON，末尾自带换行

	// Decoder：从 Reader 读 JSON
	r := strings.NewReader(`{"name":"奶茶","age":20}`)
	dec := json.NewDecoder(r)

	var p Person
	_ = dec.Decode(&p)
	fmt.Printf("%+v\n", p)
}
```

| API | 场景 |
| --- | --- |
| `Marshal` / `Unmarshal` | 内存里一整段 JSON |
| `NewEncoder` / `NewDecoder` | 文件、HTTP Body、长连接等流 |

## 常见坑

### 1. 字段未导出

```go
type Bad struct {
	name string // 小写：Marshal 后是 {}
}
```

必须写成 `Name`，再用 tag 映射成 `name`。

### 2. 时间字段

`time.Time` 默认按 RFC3339 字符串编解码：

```go
type Event struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
}
```

JSON 侧一般是 `"2026-10-15T23:52:00+08:00"` 这种格式。

### 3. 指针字段与 omitempty

想用「没传」和「传了零值」区分时，常用指针：

```go
type Patch struct {
	Age *int `json:"age,omitempty"`
}
// Age == nil → 没传；Age 指向 0 → 明确传了 0
```

### 4. 严格校验多余字段

默认会忽略 JSON 里多出来的键。若要拒绝未知字段：

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
err := dec.Decode(&p)
```

## 速查表

| 你想做的事 | 优先用 |
| --- | --- |
| 值 → JSON 字节 | `json.Marshal` |
| 美化输出 | `json.MarshalIndent` |
| JSON 字节 → 值 | `json.Unmarshal(data, &v)` |
| 改 JSON 字段名 | `` `json:"xxx"` `` |
| 空值不输出 | `` `json:"xxx,omitempty"` `` |
| 永久隐藏字段 | `` `json:"-"` `` |
| 写到文件/响应流 | `json.NewEncoder(w).Encode(v)` |
| 从请求体读 | `json.NewDecoder(r).Decode(&v)` |

入门阶段先把 **`Marshal` / `Unmarshal` + 结构体 tag** 练熟；需要读写 HTTP Body 或文件时再上 `Encoder` / `Decoder`。

## 作业

1. 定义 `User`（`name`、`age`、`email`），`Marshal` 成 JSON 并打印。  
2. 把一段 JSON 字符串 `Unmarshal` 进同一个结构体，打印各字段。  
3. 给 `email` 加 `omitempty`，不赋值时确认输出里没有该字段。  
4. 用 `` `json:"-"` `` 隐藏一个 `Password` 字段，确认序列化结果不含密码。  
5. 写嵌套结构体 `Product` + `Price`，完成一次往返编解码。
