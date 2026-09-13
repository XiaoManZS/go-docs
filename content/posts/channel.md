+++
date = '2026-08-11T19:00:00+08:00'
draft = false
title = 'Go 通道'
tags = ["Go", "通道", "channel"]
categories = ["教程"]
summary = "详解 Go 通道：无缓冲与有缓冲的区别、关闭与 range 接收、用 select 多路监听"
weight = 14
+++

# 通道

上一篇讲了怎么用 `go` 启动协程、用 `WaitGroup` 等待。  
协程之间要**传数据**，Go 的推荐方式是 **channel（通道）**：一边往里送，一边从里取，比多协程同时改同一个变量更安全、更好懂。

入门先记住四件事：

1. 用 `make(chan T)` 创建通道，`T` 是收发的数据类型。
2. `ch <- v` 发送，`v := <-ch` 接收。
3. **无缓冲**和**有缓冲**行为不一样，卡不卡在这一步。
4. 多个通道同时等，用 `select`。

## 创建与收发

```go
ch := make(chan int) // 无缓冲 int 通道
ch <- 1              // 发送 1
x := <-ch            // 接收到 x
```

也可以只声明发送或接收方向（函数参数里常见）：

```go
func producer(out chan<- int) { // 只能发送
	out <- 1
}

func consumer(in <-chan int) { // 只能接收
	fmt.Println(<-in)
}
```

## 无缓冲通道

`make(chan T)` 不写容量，就是**无缓冲**：

- 发送方 `ch <- v` 会**卡住**，直到有人来接收。
- 接收方 `<-ch` 也会**卡住**，直到有人来发送。

两边必须「对上」才能继续——像当面递东西：你递、对方接，动作同时完成。

```go
package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() {
		ch <- "订单已支付" // 没人接时，这里会一直等
	}()

	msg := <-ch // main 在这里接住
	fmt.Println(msg)
}
```

输出：

```text
订单已支付
```

注意：如果**只有发送、永远没有接收**（或反过来），程序会**死锁**，运行时报 `fatal error: all goroutines are asleep - deadlock!`。

### 例子：协程算完，通过通道把结果交回 main

```go
package main

import "fmt"

func sum(nums []int, ch chan int) {
	total := 0
	for _, n := range nums {
		total += n
	}
	ch <- total // 把结果送出去
}

func main() {
	ch := make(chan int)
	go sum([]int{1, 2, 3, 4}, ch)

	result := <-ch // 等协程送来结果
	fmt.Println("总和:", result)
}
```

无缓冲通道本身就能「等对方」：**接收会阻塞到有发送**，所以这里不必再写 `WaitGroup`。

## 有缓冲通道

`make(chan T, n)` 的 `n` 是缓冲区容量：

- 缓冲区**没满**时，发送可以立刻成功，不必等接收方。
- 缓冲区**空**时，接收会卡住，等有人发送。
- 缓冲区**满了**再发，才会卡住等接收腾出空位。

像信箱：最多放 `n` 封信；没满就能投，空了再取就得等。

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 2) // 容量 2

	ch <- 10 // 不阻塞
	ch <- 20 // 不阻塞
	// ch <- 30 // 再写就会卡住（缓冲区已满），除非另有协程在收

	fmt.Println(<-ch) // 10
	fmt.Println(<-ch) // 20
}
```

对比记住：

| | 无缓冲 `make(chan T)` | 有缓冲 `make(chan T, n)` |
| --- | --- | --- |
| 发送何时不卡 | 必须立刻有人接收 | 缓冲区还有空位 |
| 接收何时不卡 | 必须立刻有人发送 | 缓冲区里已有数据 |
| 典型用途 | 同步、交接结果 | 削峰、批量暂存 |

入门选型：

- 要「算完立刻交回 / 两边对齐」→ 无缓冲够用。
- 生产者偶尔比消费者快一点，想先攒几条 → 有缓冲。

## 关闭通道与 range

发送方不再发时，应 `close(ch)`。接收方可以用 `range` 一直取，直到通道关闭：

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 3)

	ch <- 1
	ch <- 2
	ch <- 3
	close(ch) // 关闭后不能再发送；还能继续接收剩余数据

	for v := range ch {
		fmt.Println(v)
	}
}
```

约定：

1. **谁发送，谁关闭**；接收方一般不要 `close`。
2. 对已关闭的通道再发送会 **panic**。
3. 从已关闭且空的通道接收，会立刻拿到类型零值；可用 `v, ok := <-ch`，`ok == false` 表示已关闭且没有数据了。

```go
v, ok := <-ch
if !ok {
	fmt.Println("通道已关闭")
}
```

## select：同时等多个通道

`select` 像通道版的 `switch`：哪个 case 能立刻收发，就走哪个；多个都能走时**随机选一个**。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "来自 ch1"
	}()
	go func() {
		time.Sleep(30 * time.Millisecond)
		ch2 <- "来自 ch2"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		}
	}
}
```

常见输出（`ch2` 先就绪）：

```text
来自 ch2
来自 ch1
```

### default：非阻塞尝试

所有 case 都暂时走不了时，如果写了 `default`，就立刻执行 `default`，**不会卡死**：

```go
select {
case msg := <-ch:
	fmt.Println("收到:", msg)
default:
	fmt.Println("暂时没有数据")
}
```

适合「试一下有没有消息，没有就干别的」。

### 例子：带超时

用 `time.After` 做一个超时分支，避免一直等某个通道：

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch <- "慢请求的结果"
	}()

	select {
	case msg := <-ch:
		fmt.Println(msg)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("超时了")
	}
}
```

输出：

```text
超时了
```

（把超时改成 `300ms`，就能收到「慢请求的结果」。）

## 怎么选（入门版）

| 场景 | 建议 |
| --- | --- |
| 协程算完把结果交回 | 无缓冲通道，`main` 里 `<-ch` |
| 多个结果依次送出 | 发送完 `close`，接收端 `for range` |
| 想先攒几条再处理 | 有缓冲 `make(chan T, n)` |
| 同时等好几个通道 / 要超时 | `select` |
| 试一下有没有数据，没有就继续 | `select` + `default` |
| 多协程只是「等全部做完」、不传数据 | 继续用上一篇的 `WaitGroup` |

入门阶段先练熟：**创建 → 无缓冲/有缓冲区别 → 关闭与 range → select 多路与超时**。

## 作业

1. **无缓冲交接**：写函数 `double(n int, ch chan int)`，把 `n * 2` 发送到通道。在 `main` 里 `go double(21, ch)`，再接收并打印结果（应是 `42`）。
2. **体会阻塞**：无缓冲通道上，先在 `main` 里发送一个值、再启动协程接收——观察是否死锁。改成「先 `go` 接收（或先 `go` 发送），再在另一边操作」，说明无缓冲为什么要两边配合。
3. **有缓冲**：创建容量为 `2` 的 `chan string`，连续发送 `"A"`、`"B"`（不启动其他协程），再依次接收打印。思考：若再发送第三个而不先接收，会发生什么？
4. **关闭与 range**：启动一个协程，向通道依次发送 `1..5` 后 `close`；`main` 里用 `for range` 打印所有数字。
5. **select 二选一**：两个协程分别向 `ch1`、`ch2` 发送字符串（可加不同的 `Sleep`）。用 `select` 接收两次并打印，观察谁先到。
6. **综合：超时查询**：模拟查询：协程 `Sleep` 150ms 后往通道送 `"查到了"`。`main` 用 `select`：100ms 内收到就打印结果，否则打印「查询超时」。再把超时改成 200ms，确认能收到结果。
