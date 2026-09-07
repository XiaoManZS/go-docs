+++
date = '2026-08-10T19:00:00+08:00'
draft = false
title = 'Go 协程'
tags = ["Go", "协程", "goroutine"]
categories = ["教程"]
summary = "详解 Go 协程：用 go 启动、main 提前退出的坑、用 WaitGroup 等待多任务完成"
weight = 13
+++

# 协程

普通函数按顺序执行：前面的不结束，后面的不会开始。  
协程（goroutine）让你可以**同时跑多段逻辑**——比如一边查订单、一边发短信、一边写日志，彼此不互相堵着。

Go 的协程由运行时调度，比操作系统线程更轻。入门阶段先记住三件事：

1. 用 `go` 启动协程。
2. `main` 结束，整个程序就结束，其他协程也会被直接掐掉。
3. 多个协程要「等它们都做完」，常用 `sync.WaitGroup`。

> 协程之间怎么**传数据**，主要靠 `channel`。`channel` 内容较多，**下一篇单独讲**；本文只解决「怎么开」和「怎么等」。

## 创建协程

在函数调用前加 `go`，就会在新的协程里执行，**不会阻塞当前代码**：

```go
package main

import (
	"fmt"
	"time"
)

func say(msg string) {
	fmt.Println(msg)
}

func main() {
	go say("你好，协程")
	fmt.Println("你好，main")

	// 先凑合等一下，否则 main 太快结束，协程可能还没打印
	time.Sleep(time.Millisecond * 100)
}
```

输出顺序不一定固定，常见类似：

```text
你好，main
你好，协程
```

也可以直接 `go` 一个匿名函数：

```go
go func() {
	fmt.Println("匿名协程")
}()
```

## 坑：`main` 结束，协程也没了

下面这段**常常什么都打印不出来**（或偶尔能看到一行）：

```go
package main

import "fmt"

func main() {
	go func() {
		fmt.Println("我可能来不及打印")
	}()
	// main 立刻返回，程序退出
}
```

原因：`go` 只是启动，不保证立刻跑完。`main` 一结束，进程就退出，其他协程直接被终止。

### 为什么不能靠 `Sleep` 硬等

多个协程时，每个任务耗时往往**不一样、也不好猜**。你在 `main` 里写 `time.Sleep(xxx)`，时间是自己拍的：

- 睡短了：有的协程还没做完，程序就退出了，结果丢了。
- 睡长了：任务早做完了，还干等着，浪费时间。

所以正式写法不用「自己估延时」，而是用标准库 **`sync`** 包里的 **`WaitGroup`**：谁做完谁报到，全部报到再往下走。

## 用 WaitGroup 等待

`sync.WaitGroup` 像一个计数器：

| 方法 | 作用 |
| --- | --- |
| `Add(n)` | 计数器加 `n`（准备启动几个任务就加几） |
| `Done()` | 计数器减 `1`（一个任务结束就调一次） |
| `Wait()` | 卡住，直到计数器变为 `0` |

约定：每个协程结束时调用 `Done()`，通常写成 `defer wg.Done()`，避免漏调。

### 例子：并发处理多个订单

模拟给 3 个订单分别「发货通知」。每个通知耗时不同——如果用 `Sleep` 等，你根本不知道该睡多久；用 `WaitGroup` 就不需要猜。

先看**错误示范**（靠猜延时）：

```go
package main

import (
	"fmt"
	"time"
)

func notify(orderID string, cost time.Duration) {
	time.Sleep(cost) // 这里的 Sleep 只是「模拟业务耗时」，不是用来等协程
	fmt.Println("已通知订单:", orderID)
}

func main() {
	go notify("A1001", 200*time.Millisecond)
	go notify("A1002", 100*time.Millisecond)
	go notify("A1003", 150*time.Millisecond)

	// 自己拍一个 120ms：A1002 能出来，A1001 / A1003 可能来不及
	time.Sleep(120 * time.Millisecond)
	fmt.Println("全部通知完成？——不一定")
}
```

再看**正确写法**（`sync.WaitGroup`）：

`wg` 写在 `main` 里，外面的 `notify` **本来读不到**这个变量——这和普通局部变量一样。所以要把地址 **`&wg` 当参数传进去**，函数签名写成 `wg *sync.WaitGroup`。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// 第三个参数接收 WaitGroup 的指针，否则外面读不到 main 里的 wg
func notify(orderID string, cost time.Duration, wg *sync.WaitGroup) {
	defer wg.Done() // 这个协程结束时，计数器 -1

	time.Sleep(cost) // 仅模拟「发短信 / 调接口」耗时，与「等不等协程」无关
	fmt.Println("已通知订单:", orderID)
}

func main() {
	var wg sync.WaitGroup // 定义在 main 里：只有 main 和「被传进去的函数」能用它

	orders := []struct {
		id   string
		cost time.Duration
	}{
		{"A1001", 200 * time.Millisecond},
		{"A1002", 100 * time.Millisecond},
		{"A1003", 150 * time.Millisecond},
	}

	wg.Add(len(orders)) // 告诉 WaitGroup：一共要等 3 个
	for _, o := range orders {
		go notify(o.id, o.cost, &wg) // 传地址，不是传副本
	}

	wg.Wait() // 不猜时间：三个 Done 都调用完才继续
	fmt.Println("全部通知完成")
}
```

如果嫌传参麻烦，也可以把逻辑写在 `main` 里的匿名函数中，直接闭包用到 `wg`（效果一样）：

```go
wg.Add(1)
go func(id string, cost time.Duration) {
	defer wg.Done()
	time.Sleep(cost)
	fmt.Println("已通知订单:", id)
}(o.id, o.cost)
```

两种写法二选一即可：**独立函数 → 必须传 `*sync.WaitGroup`；匿名函数 → 可直接用外层的 `wg`。**

输出类似（顺序可能变）：

```text
已通知订单: A1002
已通知订单: A1003
已通知订单: A1001
全部通知完成
```

对比记住：

| | `time.Sleep` 硬等 | `sync.WaitGroup` |
| --- | --- | --- |
| 依据 | 自己定的毫秒数，不准 | 每个协程调用 `Done()` 报到 |
| 任务变慢时 | 容易提前退出、丢结果 | 自动多等一会儿 |
| 任务变快时 | 白白多睡 | 做完立刻往下走 |
| 适用 | 演示「协程能跑起来」可以凑合 | **多协程正式等待** |

注意三点：

1. `wg` 在 `main` 里时，外层命名函数**读不到**，必须加参数 `wg *sync.WaitGroup`，调用时传 `&wg`。
2. 一定要传**指针**。若写成 `wg sync.WaitGroup`（值传递），协程里 `Done` 的是副本，`main` 里的 `Wait` 永远等不到。
3. `Add` 在启动协程**之前**调用；`Done` 在协程内部调用。

### 例子：并发算总分

多个学生各有一科成绩，分别在协程里累加到总分（这里用局部变量返回思路，先避免多协程同时写同一个变量）：

```go
package main

import (
	"fmt"
	"sync"
)

func sumScores(name string, scores []int, wg *sync.WaitGroup) {
	defer wg.Done()

	total := 0
	for _, s := range scores {
		total += s
	}
	fmt.Printf("%s 总分: %d\n", name, total)
}

func main() {
	var wg sync.WaitGroup

	students := map[string][]int{
		"小满": {90, 85, 88},
		"张三": {70, 92, 80},
		"李四": {95, 91, 89},
	}

	wg.Add(len(students))
	for name, scores := range students {
		// 把循环变量拷一份再传进协程，避免闭包踩坑
		n, s := name, scores
		go sumScores(n, s, &wg)
	}

	wg.Wait()
	fmt.Println("统计结束")
}
```

> 循环里 `go` 时，务必把 `name`、`scores` **拷到新变量**再传入。否则多个协程可能都用到「最后一次循环」的值（Go 1.22 之前尤其常见）。习惯上写成 `n, s := name, scores` 最稳妥。

## 闭包参数再强调一次

错误示范（容易踩坑）：

```go
for _, name := range names {
	go func() {
		fmt.Println(name) // 可能全是最后一个 name
	}()
}
```

正确写法：把值作为参数传进去：

```go
for _, name := range names {
	go func(n string) {
		fmt.Println(n)
	}(name)
}
```

## 怎么选（入门版）

| 场景 | 建议 |
| --- | --- |
| 只是顺序执行业务 | 普通函数调用即可，不必上协程 |
| 多个互不依赖的任务要同时跑 | `go` + `sync.WaitGroup` |
| 想「等全部完成再继续」 | 用 `wg.Wait()`，**不要**在 `main` 里猜一个 `Sleep` |
| `Sleep` 可以用在哪 | 模拟业务耗时（发短信、调接口）；**不**用来等协程结束 |
| 协程之间要传结果、收发消息 | 用 `channel`（**下篇**） |
| 多个协程同时改同一块数据（如同一个 `map`） | 先别这么做；需要时用锁或 `channel` 串行化 |

入门阶段先练熟：**启动 → 别让 main 提前走 → 用 `sync.WaitGroup` 等齐（别靠猜延时）**。共享数据的安全写法放到后续专题。

## 作业

1. **启动协程**：写函数 `hello(name string)` 打印 `你好, xxx`。在 `main` 里用 `go` 调用它，并用 `time.Sleep` 短暂等待后结束（体会「协程能跑起来」即可；这里 Sleep 只是凑合演示）。
2. **猜延时不准**：启动 3 个协程，内部分别 `Sleep` 50ms / 200ms / 100ms 后打印。在 `main` 里只 `Sleep(80ms)` 再打印「结束」。观察哪些输出丢了，用一两句话说明为什么多协程不该靠自己定延时。
3. **WaitGroup 基础**：把上一题改成 `sync.WaitGroup`：三个任务都 `Done` 后，再打印「结束」。确认每次都能看到全部输出。
4. **并发通知**：给定订单号切片 `[]string{"O1", "O2", "O3", "O4"}`，每个订单在协程里 `Sleep` 50 毫秒后打印「处理完成: 订单号」。用 `WaitGroup` 等全部处理完再打印「收银台下班」。
5. **综合练习**：定义函数 `download(file string, wg *sync.WaitGroup)`，内部 `Sleep` 100 毫秒后打印「已下载: 文件名」。文件列表为 `[]string{"a.txt", "b.txt", "c.txt"}`。用循环 + `go` 启动下载，注意循环变量拷贝（或参数传入），最后 `Wait` 并打印「全部下载完成」。
