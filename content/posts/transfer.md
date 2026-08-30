+++
date = '2026-08-10T19:00:00+08:00'
draft = false
title = 'Go 值传递与引用类型'
tags = ["Go", "值传递", "引用类型", "指针"]
categories = ["教程"]
summary = "详解 Go 的值传递：哪些类型传参后修改会影响原数据，结构体与指针的区别"
weight = 12
+++

# 值传递与引用类型

调用函数时，实参怎么传给形参？改形参会不会影响外面的变量吗？  
Go 的答案比「值传递 / 引用传递」二分法更精确：**Go 里永远是值传递**——传参时会复制一份。但有些类型的「那一份」只是描述符（header），底层数据仍然共享，所以看起来就像「引用传递」。

## 先记住一句话

> **Go 只有值传递，没有 C++ 那种引用传递（`&`）。**  
> 区别不在于语法，而在于**被复制的到底是什么**：复制整个 `int`，还是复制一个指向底层数据的「头」。

| 分类 | 常见类型 | 传参时复制什么 | 函数内改形参，外面会变吗 |
| --- | --- | --- | --- |
| 值类型 | `int`、`bool`、`float`、`array`、`struct`（普通字段） | 完整拷贝一份数据 | **不会**（改的是副本） |
| 引用类型 | `slice`、`map`、`func`、`interface` | 复制 header，共享底层 | **会**（改的是同一份底层数据） |
| 指针 | `*T` | 复制指针值（地址） | **会**（通过地址改原数据） |

下面分开看。

## 值传递

值传递时，函数拿到的是实参的**完整副本**。形参和实参是两个变量，互不影响。

```go
package main

import "fmt"

func change(x int) {
	x = 20 // 改的是副本
}

func main() {
	x := 10
	change(x)
	fmt.Println(x) // 10，外面的 x 没变
}
```

数组也是值类型——传参会复制**整个数组**，开销比切片大，日常更常用切片：

```go
package main

import "fmt"

func change(arr [3]int) {
	arr[0] = 100
}

func main() {
	arr := [3]int{1, 2, 3}
	change(arr)
	fmt.Println(arr) // [1 2 3]，原数组没变
}
```

## 引用类型有哪些

Go 没有官方术语「引用类型」，但社区通常把下面几类归为一组——它们的变量里藏着**指向底层数据的指针**：

| 类型 | 说明 |
| --- | --- |
| `slice` | 切片，可变长序列 |
| `map` | 映射，键值对 |
| `func` | 函数值，指向同一段代码 |
| `interface` | 接口值，内部含类型与数据指针 |
| `chan` | 通道，用于 goroutine 通信（**本节仅了解名字，详细用法见后续专题**） |

下面用 `slice` 和 `map` 举例（入门阶段重点掌握这两个）；`func`、`interface` 会在函数、接口相关章节深入，这里知道「也属于引用类型」即可。

### 切片 slice

传参时复制的是 slice header（指针、长度、容量），**底层数组共享**。所以在函数里改元素，外面看得见：

```go
package main

import "fmt"

func change(s []string) {
	s[0] = "猫来"
}

func main() {
	s := []string{"牛来", "羊来", "码来"}
	change(s)
	fmt.Println(s) // [猫来 羊来 码来]
}
```

用 `append` 追加元素时，只要**没触发扩容**（未超过 `cap`），同样共享底层数组：

```go
package main

import "fmt"

func appendOne(s []int) {
	s = append(s, 100) // 在共享的底层数组上追加
}

func main() {
	s := make([]int, 0, 4) // len=0, cap=4
	appendOne(s)
	fmt.Println(s, len(s)) // [] 0 —— 注意：形参 s 的 len 变了，但 main 里的 s 仍是原来的 header
}
```

上面这个例子要仔细看：`appendOne` 里 `s` 的 `len` 变了，但 `main` 里的 `s` 还是原来的 header（`len` 仍为 0）。  
**改已有下标的元素**会生效；**`append` 可能只改形参自己的 header**，外面不一定变长。若 `append` 触发扩容、分配了新数组，则完全与外面无关。

```go
package main

import "fmt"

func appendMany(s []int) {
	s = append(s, 1, 2, 3) // cap 不够时会新分配数组
}

func main() {
	s := make([]int, 0, 1)
	appendMany(s)
	fmt.Println(s, len(s)) // [] 0 —— 扩容后外面完全不受影响
}
```

要点：**改 `s[i]` 共享；`append` 是否影响外面，要看有没有扩容、以及是否把新 slice 返回给调用方。**

### 映射 map

`map` 传参后，内外指向同一张哈希表。改值、删键、增键，外面都看得见：

```go
package main

import "fmt"

func change(m map[string]int) {
	m["语文"] = 800
}

func main() {
	m := map[string]int{
		"语文": 100,
		"数学": 99,
		"英语": 98,
	}
	change(m)
	fmt.Println(m) // map[语文:800 数学:99 英语:98]
}
```

调用 `delete` 等 API，同样共享底层：

```go
package main

import "fmt"

func change(m map[string]int) {
	m["语文"] = 800
	delete(m, "数学")
}

func main() {
	m := map[string]int{
		"语文": 100,
		"数学": 99,
		"英语": 98,
	}
	change(m)
	fmt.Println(m) // map[语文:800 英语:98]
}
```

### 通道 chan（了解即可）

`chan` 和 `slice`、`map` 一样，传参后内外指向同一个通道。  
通道用于 goroutine 之间收发数据，用法较多，**本教程不展开**；后续会有单独一篇。这里只要记住：**它也是引用类型之一**。

## 结构体：值传递 vs 指针传递

结构体本身是**值类型**。传 `Person` 会复制整个结构体，函数里改字段不影响外面：

```go
package main

import "fmt"

type Person struct {
	Name  string
	Age   int
	Email string
}

func change(person Person) {
	person.Name = "张满月"
	person.Age = 91
	person.Email = "zmy@qq.com"
}

func main() {
	person := Person{
		Name:  "小满",
		Age:   25,
		Email: "1195566313@qq.com",
	}
	change(person)
	fmt.Println(person) // {小满 25 1195566313@qq.com}，原结构体未变
}
```

若希望函数内修改能反映到外面，传**指针** `*Person`：

```go
package main

import "fmt"

type Person struct {
	Name  string
	Age   int
	Email string
}

func change(person *Person) {
	person.Name = "张满月" // 语法糖，等价于 (*person).Name
	person.Age = 91
	person.Email = "zmy@qq.com"
}

func main() {
	person := Person{
		Name:  "小满",
		Age:   25,
		Email: "1195566313@qq.com",
	}
	change(&person)
	fmt.Println(person) // {张满月 91 zmy@qq.com}
}
```

补充说明：

- `*Person` 是**指针类型**，不是「引用类型」表格里的那一类；传指针时复制的是**地址**，通过地址仍能改到原结构体。
- 结构体里某字段写成 `*string`，表示该字段是指针，和「把整个人当成 `*Person` 传递」是不同层面的概念。入门阶段优先掌握 **`Person` 与 `*Person` 的区别** 即可。

## 怎么选

| 场景 | 建议 |
| --- | --- |
| 基本类型、小结构体，函数只读不改 | 直接传值 |
| 需要在函数里改调用方的结构体 | 传 `*Struct` |
| 一组可变长数据 | 用 `slice`，注意 `append` 与共享的规则 |
| 按键查找、统计 | 用 `map` |
| 不确定要不要改外面 | 先想清楚「要不要共享」；要共享就指针或引用类型，不要共享就传值或返回新值 |

## 作业

1. **值传递**：写 `setZero(n int)`，在函数内把 `n` 设为 `0`。在 `main` 里传入 `42`，调用前后各打印一次，确认外面仍是 `42`。
2. **切片改元素**：写 `firstToX(s []string)`，把第一个元素改成 `"X"`。在 `main` 里传入 `[]string{"A", "B", "C"}`，调用后打印切片，确认已改变。
3. **map 增删改**：写 `updateScore(scores map[string]int, subject string, delta int)`，给某科加分；若科目不存在则新建。再写 `removeSubject(scores map[string]int, subject string)` 删除一科。在 `main` 里验证增、改、删后 map 的变化。
4. **结构体值 vs 指针**：定义 `Counter struct { Value int }`。分别写 `addByValue(c Counter)` 和 `addByPointer(c *Counter)`，都让 `Value + 1`。在 `main` 里各调用一次，对比哪种方式能让外面的 `Counter` 真的加 1。
5. **综合练习**：用 `map[string][]int` 表示「学生姓名 → 各科成绩列表」。写函数 `addScore(scores map[string][]int, name string, score int)` 追加一门成绩；写 `fixFirst(scores map[string][]int, name string, newScore int)` 把该生第一门课分数改成 `newScore`（若存在）。在 `main` 里多次调用后打印，体会 map 与 slice 嵌套时的共享行为。
