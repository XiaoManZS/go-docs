+++
date = '2026-07-01T19:00:00+08:00'
draft = false
title = 'Go 条件判断'
tags = ["Go", "条件判断"]
categories = ["教程"]
summary = "详解Go条件判断：if、else、else if、switch、case、default"
weight = 5
+++

# 北冥神功(太玄经)

```go
package main

import "fmt"

type Person struct {
	Name string
	Age  int
	Sex  string
}

func main() {
	person := Person{
		Name: "张三",
		Age:  18,
		Sex:  "男",
	}
	if person.Age >= 18 {
		fmt.Println("成年")
	} else {
		fmt.Println("未成年")
	}
}
```

我们定义了一个 `Person` 结构体，包含 `Name`、`Age`、`Sex` 三个字段。然后创建了一个实例，判断其年龄是否满 18 岁：满 18 岁输出「成年」，否则输出「未成年」。这就是最基本的 `if` / `else` 二选一判断。


## 多条件判断

```go
package main

import "fmt"

type Person struct {
	Status string
}

func main() {
	person := Person{
		Status: "在线",
	}
	if person.Status == "在线" {
		fmt.Println("在线")
	} else if person.Status == "离线" {
		fmt.Println("离线")
	} else if person.Status == "忙碌" {
		fmt.Println("忙碌")
	} else {
		fmt.Println("未知状态")
	}
}
```

当需要处理**三种及以上**的分支时，可以在 `if` 后面连续使用 `else if`，从上到下依次判断，**命中第一个为真的分支后就不再往下执行**。最后的 `else` 用于兜底，处理所有未匹配的情况。

上面示例模拟了用户在线状态：`Status` 为「在线」「离线」「忙碌」时分别输出对应文字，其他值则输出「未知状态」。

同样的逻辑，用 `switch` 写会更简洁——把要判断的值写在 `switch` 后面，每个 `case` 对应一种取值：

```go
package main

import "fmt"

type Person struct {
	Status string
}

func main() {
	person := Person{
		Status: "在线",
	}
	switch person.Status {
	case "在线":
		fmt.Println("在线")
	case "离线":
		fmt.Println("离线")
	case "忙碌":
		fmt.Println("忙碌")
	default:
		fmt.Println("未知状态")
	}
}
```

`switch` 会拿 `person.Status` 与每个 `case` 逐一比较，匹配到就执行对应分支。Go 的 `switch` **默认不会穿透**到下一个 `case`（不需要写 `break`）。所有 `case` 都不匹配时，执行 `default` 分支。

| 写法 | 适用场景 |
| --- | --- |
| `if` / `else if` | 条件各不相同，不好抽成一个表达式 |
| `switch 表达式` | 同一个值匹配多种固定取值（如状态、等级、星期） |

## 逻辑嵌套

```go
package main

import "fmt"

type Person struct {
	Name string
	Age  int
	Sex  string
}

func main() {
	person := Person{
		Name: "张三",
		Age:  18,
		Sex:  "男",
	}
	if person.Age >= 18 {
		if person.Sex == "男" {
			fmt.Println("成年男性")
		} else {
			fmt.Println("成年女性")
		}
	} else {
		fmt.Println("未成年")
		if person.Sex == "男" {
			fmt.Println("未成年男性")
		} else {
			fmt.Println("未成年女性")
		}
	}
}
```

多个if条件判断可以嵌套使用，我们首先判断年龄是否等于18，如果等于18，则再判断性别是否等于"男"，如果等于"男"，则输出"成年男性"，否则输出"成年女性"。如果不等于18，则输出"未成年"，再判断性别是否等于"男"，如果等于"男"，则输出"未成年男性"，否则输出"未成年女性"。这就是一个基本的逻辑嵌套。当然也不是只能嵌套两层，可以嵌套多层，但是要注意逻辑的正确性。


## 逻辑运算符

判断条件经常需要组合多个表达式，常用的逻辑运算符如下：

| 运算符 | 含义 | 示例 | 为真条件 |
| --- | --- | --- | --- |
| `&&` | 并且 | `a && b` | `a`、`b` 都为真 |
| `\|\|` | 或者 | `a \|\| b` | `a`、`b` 至少一个为真 |
| `!` | 取反 | `!a` | `a` 为假 |

### 并且（`&&`）

```go
age := 18
sex := "男"

if age >= 18 && sex == "男" {
	fmt.Println("成年男性")
}
```

`&&` 表示**并且**：左右两边条件都为真，整个表达式才为真。上面只有「年满 18 岁」且「性别为男」时，才会打印「成年男性」。

### 或者（`||`）

```go
age := 33

if age == 18 || age == 19 {
	fmt.Println("18 岁或 19 岁")
} else {
	fmt.Println("其他年龄")
}
```

`||` 表示**或者**：左右两边只要有一个为真，整个表达式就为真。上面当年龄是 18 **或** 19 时走 `if` 分支，否则走 `else`。

### 取反（`!`）

```go
isLogin := false

if !isLogin {
	fmt.Println("请先登录")
}
```

`!` 把布尔值**反过来**：`!true` 为 `false`，`!false` 为 `true`。常用于「不满足某条件时」的分支判断。

## 作业

1. **基础判断**：声明一个 `int` 类型的分数变量，用 `if` / `else if` / `else` 判断等级：90 分及以上为「优秀」，60～89 为「及格」，60 分以下为「不及格」，并打印结果。
2. **switch 练习**：定义一个 `string` 变量表示星期几（如 `"周一"`），用 `switch` 匹配并打印对应提示；未匹配时走 `default`，输出「无效的星期」。
3. **状态判断**：参照文中的 `Person` 结构体，将 `Status` 分别设为「在线」「离线」「忙碌」和一个自定义值，分别运行 `if` / `else if` 和 `switch` 两个版本，观察输出是否一致。
4. **逻辑运算符**：声明 `age` 和 `hasTicket` 两个变量，用 `&&` 判断「年满 18 岁且已购票」方可入场；再改用 `||` 判断「会员或年龄大于 60 岁」可享受优惠，分别打印结果。
5. **综合练习**：定义包含 `Name`、`Age`、`Score` 的结构体，结合 `if` 嵌套判断：年满 18 且分数不低于 60 才输出「合格」，否则输出「不合格」，并带上姓名。