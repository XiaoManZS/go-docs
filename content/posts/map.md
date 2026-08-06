+++
date = '2026-08-01T19:00:00+08:00'
draft = false
title = 'Go map映射'
tags = ["Go", "map"]
categories = ["教程"]
summary = "详解Go map：定义、初始化、操作、删除"
weight = 7
+++

# Map

Map 是一种**无序**的键值对集合，适合按「键」快速查找「值」。日常开发里常用它存配置、统计次数、缓存结果等。

Map 是引用类型，零值是 `nil`。**键必须是可比较的类型**（如 `string`、`int`、数组等），不能是切片、map、函数；值可以是任意类型。

## 语法

```go
var 变量名 map[键类型]值类型
变量名 := map[键类型]值类型{键: 值, ...}
变量名 := make(map[键类型]值类型)
```

## 定义与初始化

有三种常见写法：字面量、`make`、以及先声明再赋值。

```go
package main

import "fmt"

func main() {
	// 1. 字面量初始化（推荐，一眼能看出内容）
	scores := map[string]int{
		"语文": 90,
		"数学": 85,
		"英语": 92,
	}

	// 2. make 创建空 map，再往里塞
	person := make(map[string]string)
	person["name"] = "张三"
	person["city"] = "杭州"

	// 3. 先声明（此时是 nil），再赋值一个已有 map
	var ages map[string]int
	ages = map[string]int{"张三": 18, "李四": 20}

	fmt.Println(scores)
	fmt.Println(person)
	fmt.Println(ages)
}
```

注意：只写 `var m map[string]int` 而不初始化时，`m` 是 `nil`。**向 nil map 写入会 panic**，读却可以（得到零值）。需要写入时，务必先用 `make` 或字面量初始化。

```go
package main

func main() {
	var m map[string]int
	// m["a"] = 1  // panic: assignment to entry in nil map
	_ = m["a"]     // 读可以，得到 0
}
```

也可以给 `make` 传第二个参数，预估容量，减少扩容次数（不是长度上限）：

```go
m := make(map[string]int, 10)
```

## 读写元素

用 `map[键]` 读写，和切片按下标访问很像：

```go
package main

import "fmt"

func main() {
	scores := map[string]int{
		"语文": 90,
		"数学": 85,
	}

	fmt.Println(scores["语文"]) // 90

	scores["数学"] = 95        // 修改
	scores["英语"] = 88        // 新增
	fmt.Println(scores)       // map[语文:90 数学:95 英语:88]
}
```

键不存在时，读出来的是**值类型的零值**，不会报错：

```go
package main

import "fmt"

func main() {
	scores := map[string]int{"语文": 90}
	fmt.Println(scores["体育"]) // 0（int 的零值）
}
```

所以单靠读到的值，分不清「键不存在」和「值刚好是零值」。这时要用「逗号 ok」写法。

## 判断键是否存在

`value, ok := map[键]`：`ok` 为 `true` 表示键存在。

```go
package main

import "fmt"

func main() {
	scores := map[string]int{
		"语文": 90,
		"体育": 0, // 值就是 0
	}

	if score, ok := scores["语文"]; ok {
		fmt.Println("语文成绩：", score) // 语文成绩： 90
	}

	if score, ok := scores["体育"]; ok {
		fmt.Println("体育成绩：", score) // 体育成绩： 0（键存在，值就是 0）
	} else {
		fmt.Println("没有体育这门课")
	}

	if _, ok := scores["音乐"]; !ok {
		fmt.Println("没有音乐这门课")
	}
}
```

只要判断存在与否、不关心值时，用 `_` 丢掉第一个返回值即可：`_, ok := scores["音乐"]`。

## 删除元素

用内置函数 `delete`。键不存在时调用也是安全的，不会报错：

```go
package main

import "fmt"

func main() {
	scores := map[string]int{
		"语文": 90,
		"数学": 85,
		"英语": 92,
	}

	delete(scores, "数学")
	fmt.Println(scores) // map[语文:90 英语:92]

	delete(scores, "体育") // 键不存在，什么都不发生
	fmt.Println(scores)
}
```

`delete` 没有返回值，也不会返回「是否删成功」。若要确认，删除后再用逗号 ok 查一次即可。

## 获取长度

用 `len` 获取键值对个数：

```go
package main

import "fmt"

func main() {
	scores := map[string]int{
		"语文": 90,
		"数学": 85,
	}
	fmt.Println(len(scores)) // 2

	delete(scores, "数学")
	fmt.Println(len(scores)) // 1
}
```

Map 没有「容量」概念，不像切片那样有 `cap`。

## 遍历 map

用 `range` 遍历。每次迭代返回**键**和**值**：

```go
package main

import "fmt"

func main() {
	person := map[string]string{
		"name": "张三",
		"city": "杭州",
		"job":  "程序员",
	}

	for key, value := range person {
		fmt.Println(key, value)
	}
}
```

只要键或只要值时：

```go
for key := range person {
	fmt.Println(key)
}

for _, value := range person {
	fmt.Println(value)
}
```

注意：map 的遍历**顺序不固定**，每次运行打印顺序可能不同。若需要稳定顺序，可以把键收集到切片里再排序后遍历。

## 作为函数参数

Map 是引用类型，传给函数后，函数里对它的增删改会反映到外面：

```go
package main

import "fmt"

func bump(scores map[string]int, subject string, delta int) {
	scores[subject] += delta
}

func main() {
	scores := map[string]int{"语文": 90}
	bump(scores, "语文", 5)
	fmt.Println(scores["语文"]) // 95
}
```

但若在函数里把参数重新指向另一个 map（`scores = make(...)`），外面的变量不会跟着变——改的是局部变量本身，不是底层数据。

## 常见注意点

| 点 | 说明 |
| --- | --- |
| 零值 | `nil`，写入会 panic，读取得到值类型零值 |
| 键类型 | 必须可比较，不能是 slice / map / function |
| 无序 | `range` 顺序不固定 |
| 键不存在 | 读到零值；用 `value, ok := m[k]` 区分 |
| 并发 | 默认**不是**并发安全的，多 goroutine 同时写会出问题 |

## 作业

1. **声明与读写**：创建一个 `map[string]string`，存自己的姓名、城市、爱好；修改城市，再新增一个「职业」键，打印整个 map。
2. **逗号 ok**：给定 `scores := map[string]int{"语文": 90, "数学": 0}`，分别判断「语文」「数学」「英语」是否存在，并正确打印成绩或「不存在」。
3. **删除与长度**：在上一题的 map 上删除「数学」，用 `len` 打印剩余键值对数量。
4. **遍历统计**：用 `map[string]int` 统计字符串切片 `[]string{"苹果", "香蕉", "苹果", "橘子", "香蕉", "苹果"}` 中每种水果出现的次数，并遍历打印结果。
5. **综合练习**：写一个函数 `addScore(scores map[string]int, name string, score int)`，把某人的分数累加进去（没有则新建）。在 `main` 里多次调用后，打印总分最高的姓名和分数。
