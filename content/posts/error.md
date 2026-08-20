+++
date = '2026-08-10T19:00:00+08:00'
draft = false
title = 'Go 错误处理'
tags = ["Go", "错误处理", "defer"]
categories = ["教程"]
summary = "详解Go错误处理：业务错误、程序错误、defer、panic、recover"
weight = 11
+++

# 错误处理

Go **没有** `try / catch`。日常约定是：

- **能预期的失败**（参数不对、查无此人、除数为 0）→ 返回 `error`，由调用方决定怎么处理。
- **不该发生、程序状态已不可信**（下标越界、断言失败、严重逻辑崩坏）→ 用 `panic`，必要时再在上层用 `defer` + `recover` 兜底。

本文用「九九乘法表」同一个业务场景，对比这两种写法的差别。

## 业务错误 vs 程序错误

| 类型 | 常见手段 | 出错后同函数里后面的代码 | 典型场景 |
| --- | --- | --- | --- |
| 业务错误 | 返回 `(结果, error)` | **还能继续走**（只要你没 `return`） | 用户输入非法、业务校验失败 |
| 程序错误 | `panic` | **不再执行**（即使 `recover` 了也是如此） | 严重异常、不可恢复状态 |

记住口诀：**业务错误是「报告给调用方」；程序错误是「当前这条执行路径断了」。**

## 业务错误：返回 `error`

`error` 是内置接口，只有一个方法 `Error() string`。没出错返回 `nil`，有错返回非 `nil`。

下面要求参数必须是 `9` 才打印九九表；传 `8` 时返回错误。注意：处理完错误之后，`main` 里的「其余逻辑」**仍然会执行**。

```go
package main

import (
	"errors"
	"fmt"
)

// 业务校验失败：返回 error，不中断整个程序
func nn(a int) (bool, error) {
	if a != 9 {
		return false, errors.New("请输入9")
	}

	for i := 1; i <= a; i++ {
		for k := 1; k <= i; k++ {
			fmt.Printf("%d * %d = %d  ", k, i, i*k)
		}
		fmt.Println()
	}
	return true, nil
}

func main() {
	success, err := nn(8)
	if err != nil {
		fmt.Println(err, "执行失败")
	} else {
		fmt.Println(success, "执行成功")
	}

	// 业务错误只是「这次调用失败」，后面的代码照常走
	fmt.Println("其余逻辑")
}
```

输出类似：

```text
请输入9 执行失败
其余逻辑
```

要点：

1. 用 `errors.New("说明")` 或 `fmt.Errorf("a=%d 非法", a)` 创建错误。
2. 调用方标准姿势：`if err != nil { ... }`。
3. 返回错误后，**函数正常结束**；调用方可以打日志、换参数重试，也可以继续干别的事。

## 程序错误：`panic`

把「请输入 9」改成 `panic`。一旦触发，从 `panic` 那一行开始，**同函数里后面的代码不会再跑**。

先看**没有** `recover` 的情况——进程会直接崩溃：

```go
package main

import "fmt"

func nn(a int) (bool, error) {
	if a != 9 {
		panic("请输入9") // 程序错误：直接中断当前执行路径
	}

	for i := 1; i <= a; i++ {
		for k := 1; k <= i; k++ {
			fmt.Printf("%d * %d = %d  ", k, i, i*k)
		}
		fmt.Println()
	}
	return true, nil
}

func main() {
	success, err := nn(8)
	if err != nil {
		fmt.Println(err, "执行失败")
	} else {
		fmt.Println(success, "执行成功")
	}

	// 上面已经 panic，这行永远走不到
	fmt.Println("其余逻辑")
}
```

运行后会看到 panic 栈，且没有「其余逻辑」。这就是和业务错误最大的差别。

## defer 用法

`defer` 会把一次函数调用推迟到 **当前函数即将返回之前** 执行，常用于关文件、解锁、以及配合 `recover` 捕获 `panic`。

基本顺序：先写正常逻辑，`defer` 的内容最后跑。

```go
package main

import "fmt"

func main() {
	fmt.Println("开始")
	defer fmt.Println("defer：收尾")
	fmt.Println("业务逻辑")
}
```

输出：

```text
开始
业务逻辑
defer：收尾
```

多个 `defer` 按 **后进先出（LIFO）** 执行——最后注册的最先跑：

```go
package main

import "fmt"

func main() {
	defer fmt.Println("第一")
	defer fmt.Println("第二")
	defer fmt.Println("第三")
	fmt.Println("函数体")
}
```

输出：

```text
函数体
第三
第二
第一
```

和错误处理相关的关键点：

1. **`recover` 必须写在 `defer` 里**才有效；普通代码里调用 `recover()` 拿不到 panic。
2. `defer` 常用来保证「出错也要收尾」：关文件、释放锁、打印现场。
3. 参数在写 `defer` 那一刻就算好（不是真正执行时才算）。更多细节见「Go 函数」一文的 defer 小节。

## 用 `defer` + `recover` 兜住 panic

加上 `recover` 后，进程**不会崩溃**，可以打印 panic 信息；但 `panic` 发生点之后的代码**仍然不会执行**。

```go
package main

import "fmt"

func nn(a int) (bool, error) {
	if a != 9 {
		panic("请输入9")
	}

	for i := 1; i <= a; i++ {
		for k := 1; k <= i; k++ {
			fmt.Printf("%d * %d = %d  ", k, i, i*k)
		}
		fmt.Println()
	}
	return true, nil
}

func main() {
	defer func() {
		if info := recover(); info != nil {
			fmt.Println("捕获到程序错误:", info)
		}
	}()

	success, err := nn(8) // 这里 panic
	if err != nil {
		fmt.Println(err, "执行失败")
	} else {
		fmt.Println(success, "执行成功")
	}

	// panic 之后：同函数里这些代码依然不走
	fmt.Println("其余逻辑")
}
```

输出类似：

```text
捕获到程序错误: 请输入9
```

没有「执行失败 / 执行成功」，也没有「其余逻辑」。  
`recover` 的作用是：**拦住崩溃**，不是**从 panic 下一行接着跑**。

## 怎么选

| 情况 | 建议 |
| --- | --- |
| 用户输错、查不到数据、权限不够 | 返回 `error` |
| 库函数约定用 `error`（如 `os.Open`） | 跟着返回 / 包装 `error` |
| 明显的编程错误、不可继续的状态 | 可以 `panic` |
| 服务入口想避免整个进程挂掉 | 边界处 `defer` + `recover`，打日志后返回错误响应 |

入门阶段优先练好 **`if err != nil`**；`panic` / `recover` 留给真正的异常路径，不要拿来代替普通业务校验。

## 作业

1. **返回 error**：写 `divide(a, b int) (int, error)`，`b == 0` 时返回错误，否则返回商。在 `main` 里分别用 `(10, 2)` 和 `(10, 0)` 调用并处理。
2. **业务错误对比**：改写文中的 `nn`，允许传入 `1`～`9` 任意整数打印对应乘法表；小于 `1` 或大于 `9` 时返回 `error`。调用失败后打印一句「其余逻辑」，确认它仍会执行。
3. **defer 顺序**：在 `main` 里注册三个 `defer`，打印 `A`、`B`、`C`，再打印一句「函数体」，观察输出顺序是否为「函数体 → C → B → A」。
4. **panic 与 recover**：写一个函数 `mustPositive(n int)`，`n <= 0` 时 `panic`。在 `main` 里用 `defer` + `recover` 捕获，打印 panic 信息；并在调用后面再写一句 `fmt.Println("其余逻辑")`，验证它不会执行。
5. **综合练习**：写 `printTable(n int) error`（业务错误版）和 `mustPrintTable(n int)`（`n != 9` 就 `panic`）。在 `main` 里先演示返回 `error` 后后续代码仍运行；再用 `recover` 演示 `panic` 被兜住但后续代码不运行。
