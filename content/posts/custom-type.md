+++
date = '2026-10-28T22:00:00+08:00'
draft = false
title = 'Go 自定义类型'
tags = ["Go", "自定义类型", "类型安全"]
categories = ["教程"]
summary = "详解 Go 自定义类型：挂方法、规避同类型传参错误"
weight = 19
+++

# 自定义类型

基本类型（`int`、`string`…）能用，但有时不够「说话」。比如用户名和地址都是字符串，用户 ID 和订单 ID 都是整数——混着用很容易传错参。

Go 允许用 `type` 基于已有类型定义**新类型**：

```go
type 新类型名 底层类型
```

新类型和底层类型「长得像」，但在编译器眼里是**不同的类型**。这一点带来两个很实用的能力：

1. **可以挂方法**——像结构体一样，给自己的类型加行为
2. **类型安全**——同底层类型的不同语义，编译期就能拦住传参错误

---

## 1. 自定义类型可以挂方法

定义一个基于 `string` 的 `Name`，再给它挂两个方法：加前缀、拼随机后缀。

```go
package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

// 基于 string 定义新类型 Name
type Name string

// 给 Name 挂方法：加前缀
func (n Name) PrefixName() Name {
	return "哈夫克-" + n
}

// 给 Name 挂方法：拼一个 1~100 的随机数字后缀
func (n Name) RandomName() Name {
	ran := rand.IntN(100) + 1 // IntN(100) 是 [0,100)，再 +1 变成 [1,100]
	return n + Name(strconv.Itoa(ran))
}

func main() {
	var name Name = "张三"

	fmt.Println(name.PrefixName()) // 哈夫克-张三
	fmt.Println(name.RandomName()) // 张三42（数字每次不同）
}
```

几点说明：

- `func (n Name) Xxx()` 里的 `(n Name)` 叫**接收者**，表示这个方法挂在 `Name` 上，用 `变量.方法()` 调用
- 自定义类型和底层类型不能直接相加：`"哈夫克-" + n` 可以，是因为字符串字面量和 `Name` 的底层都是 `string`，Go 会自动处理；而数字要先 `strconv.Itoa`，再显式转成 `Name(...)`
- 普通 `string` **没有**这两个方法；只有 `Name` 才能调，这正是「挂方法」的意义

> 方法也能挂在基于 `int`、`float64` 等类型上，写法一样。结构体挂方法更常见，本质相同。

---

## 2. 规避同类型传参错误

实际业务里，用户 ID、订单 ID 往往都是 `int`。如果函数写成：

```go
func CreateOrder(uid int, oid int) { ... }
```

调用时把两个参数写反，编译器**完全不管**——两个都是 `int`，语法没错，逻辑却炸了。

用自定义类型把语义拆开：

```go
package main

import "fmt"

type Uid int // 用户 ID
type Oid int // 订单 ID

// 参数用自定义类型，语义一目了然
func CreateOrder(uid Uid, oid Oid) (Uid, Oid) {
	fmt.Println("创建订单：用户", uid, "订单", oid)
	return uid, oid
}

func main() {
	uid := Uid(10)
	oid := Oid(10086)

	// 正确：类型对得上
	CreateOrder(uid, oid)

	// 错误示范（解开注释会编译失败）：
	// CreateOrder(oid, uid)
	// 报错类似：cannot use oid (variable of type Oid) as Uid value in argument to CreateOrder

	// 就算底层都是 int，也不能直接把 int 传进去：
	// var n int = 10
	// CreateOrder(n, oid) // cannot use n (variable of type int) as Uid value
}
```

对比一下：

| 写法 | 传反参数时 |
| --- | --- |
| `func CreateOrder(uid int, oid int)` | 编译通过，运行时逻辑错误 |
| `func CreateOrder(uid Uid, oid Oid)` | **编译直接报错**，问题在写代码时就被拦住 |

自定义类型和底层类型、以及其他自定义类型之间，需要**显式转换**，例如 `Uid(10)`、`int(uid)`。多写几个括号，换来的是少踩坑——很值。

---

## 小结

| 能力 | 做法 | 好处 |
| --- | --- | --- |
| 挂方法 | `type Name string` + `func (n Name) Xxx()` | 给类型绑定专属行为 |
| 类型安全 | `type Uid int` / `type Oid int` | 同底层类型不同语义，编译期防传参错误 |

记住一句话：**自定义类型 ≈ 底层类型的「带标签版本」**——能当同类数据用，但名字不同、方法不同、传参也不能乱混。
