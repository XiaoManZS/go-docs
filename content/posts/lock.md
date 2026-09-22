+++
date = '2026-08-12T19:00:00+08:00'
draft = false
title = 'Go 互斥锁与读写锁'
tags = ["Go", "锁", "Mutex", "RWMutex"]
categories = ["教程"]
summary = "详解 Go 锁：用 Mutex 保护共享切片、不加锁会丢数据，以及读多写少时用 RWMutex"
weight = 15
+++

# 互斥锁与读写锁

上一篇用 **channel** 在协程之间**传数据**。  
有时协程要**共改同一块内存**——比如几个协程一起往同一个切片里 `append`。这时通道不是唯一答案，标准库 `sync` 里的**锁**也很常用。

入门先记住三件事：

1. **互斥锁 `Mutex`**：同一时刻只允许一个协程进「临界区」改共享数据。
2. 不加锁并发改切片 / map，可能丢数据、甚至崩溃（数据竞争）。
3. **读写锁 `RWMutex`**：读可以多人同时读；写还是独占。适合「读多写少」。

> 口诀对照：channel 偏「把结果递出去」；锁偏「大家轮流改同一份」。

## 为什么要锁：奶茶菜单被同时加料

场景：菜单里已有三杯奶茶，三个协程各自再加一杯。  
共享的是同一个切片 `arr`，改它时要排队。

### 加了互斥锁：每次都能累加成功

```go
package main

import (
	"fmt"
	"sync"
)

func worker(arr *[]string, wg *sync.WaitGroup, mu *sync.Mutex, tea string) {
	defer wg.Done()
	mu.Lock()                 // 上锁：别人先别动 arr
	*arr = append(*arr, tea)  // 临界区：只有拿到锁的协程能改
	mu.Unlock()               // 解锁：下一位可以进来
}

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	arr := []string{"芋泥波波", "绿茶", "红茶"}

	wg.Add(3)
	go worker(&arr, &wg, &mu, "柠檬茶")
	go worker(&arr, &wg, &mu, "霸王茶姬")
	go worker(&arr, &wg, &mu, "爷爷不泡茶")
	wg.Wait()

	fmt.Println(arr)
}
```

多次运行，切片里都会有 **6 个名字**（原来 3 个 + 新加 3 个）。顺序可能每次不同（谁先抢到锁谁先加），但**不会少**：

```text
[芋泥波波 绿茶 红茶 柠檬茶 霸王茶姬 爷爷不泡茶]
```

或：

```text
[芋泥波波 绿茶 红茶 霸王茶姬 柠檬茶 爷爷不泡茶]
```

要点：

| 写法 | 含义 |
| --- | --- |
| `mu.Lock()` | 拿到锁；别人正在锁里时，这里会**卡住等待** |
| `mu.Unlock()` | 放开锁；漏写会让别的协程永远等 |
| `*arr` | 传的是切片指针，改的才是 `main` 里那份 |
| `wg` / `mu` 也要传指针 | 和上一篇 WaitGroup 一样，传值等于改副本，外面感知不到 |

`defer wg.Done()` 放在函数开头，保证协程怎么退出都会报到；锁则建议**尽快 Unlock**，不要把无关的慢操作包在锁里。

### 不加锁会怎样：看起来像「被盖掉」

把 `Lock` / `Unlock` 拿掉，三个协程同时 `append` 同一份切片：

```go
func worker(arr *[]string, wg *sync.WaitGroup, tea string) {
	defer wg.Done()
	*arr = append(*arr, tea) // 没有锁，多人同时改
}
```

`append` 内部要读长度、扩容、写元素——几步叠在一起。多协程交错执行时，可能出现：

- 有的茶**没加进去**（少了几杯，像被覆盖）
- 偶发 **panic**（例如索引越界）
- 用 `go run -race` 会报 **data race**

所以入门结论很简单：**多人同时改同一份切片 / map，先加锁（或改成只用 channel 串行化）。**

## 读写锁：读多写少时更合适

`Mutex` 读写都互斥：有人读菜单时，别人也不能读。  
若业务是「很多人都在看菜单，偶尔改一次」，可以用 **`sync.RWMutex`**：

- `RLock` / `RUnlock`：加**读锁**，多个读可以同时进行。
- `Lock` / `Unlock`：加**写锁**，写的时候别人既不能读也不能写。

下面继续用菜单：两个协程读、一个协程写。  
写这边故意 `Sleep(1ms)`，先让读锁先拿到，方便体会「读可以并行、写要等读完」。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func read(arr *[]string, wg *sync.WaitGroup, mu *sync.RWMutex) {
	defer wg.Done()
	mu.RLock()
	fmt.Println(*arr)
	time.Sleep(20 * time.Millisecond) // 模拟慢慢读
	mu.RUnlock()
}

func write(arr *[]string, wg *sync.WaitGroup, mu *sync.RWMutex, tea string) {
	defer wg.Done()
	time.Sleep(1 * time.Millisecond) // 先让读锁触发
	mu.Lock()
	time.Sleep(10 * time.Millisecond) // 模拟写入耗时
	*arr = append(*arr, tea)
	mu.Unlock()
}

func main() {
	var wg sync.WaitGroup
	var rwmu sync.RWMutex
	menu := []string{"芋泥波波", "绿茶", "红茶"}

	wg.Add(3)
	go read(&menu, &wg, &rwmu)
	go read(&menu, &wg, &rwmu)
	go write(&menu, &wg, &rwmu, "爷爷不泡茶")
	wg.Wait()

	fmt.Println("最终菜单", menu)
}
```

可能的输出（两个读可以重叠，所以往往先看到两次「还没加料」的菜单；写在读锁释放后才能完成）：

```text
[芋泥波波 绿茶 红茶]
[芋泥波波 绿茶 红茶]
最终菜单 [芋泥波波 绿茶 红茶 爷爷不泡茶]
```

要点：

| 写法 | 含义 |
| --- | --- |
| `mu.RLock()` / `RUnlock()` | 读锁；多个 `read` 可以同时拿着 |
| `mu.Lock()` / `Unlock()` | 写锁；要等读锁都放开，且期间别人读不了、也写不了 |
| `write` 里先 `Sleep(1ms)` | 演示用：让读先执行；真实代码一般不需要这样凑时序 |

记住：

- 只读 → `RLock` / `RUnlock`
- 要改 → 还是 `Lock` / `Unlock`（写锁）
- 入门阶段：**不确定就用 `Mutex`**；确认读远多于写，再换成 `RWMutex`

## 和 channel 怎么选（入门版）

| 场景 | 建议 |
| --- | --- |
| 协程算完，把结果交回 | channel |
| 多个结果依次汇总、或流水线 | channel |
| 多个协程共改同一个切片 / map / 计数器 | `Mutex`（或 `RWMutex`） |
| 读很多、写很少的共享数据 | 优先考虑 `RWMutex` |
| 只是等任务做完、不改共享数据 | 继续用 `WaitGroup` 即可 |

Go 社区常说「不要通过共享内存来通信，而要通过通信来共享内存」——意思是**优先 channel**。  
但计数器、缓存、就地改结构体字段这类场景，**加锁往往更直接**。两种都会用，按场景选。

## 作业

1. **复现丢数据**：用本文「不加锁」版本跑十几次，观察切片长度是否总是 6；有条件的话用 `go run -race .` 看是否报 data race。
2. **加上 Mutex**：恢复 `Lock` / `Unlock`，确认每次长度都是 6，并打印出完整菜单。
3. **计数器**：多个协程给同一个 `var n int` 各加 1000 次。分别写「无锁」和「`Mutex` 保护 `n++`」两版，对比最终 `n` 是否等于期望值。
4. **读写锁小练习**：一个 `map[string]int` 表示库存。启动 3 个协程用 `RLock` 打印某商品库存；再启动 1 个协程用写锁把该商品库存 `+1`。用 `WaitGroup` 等全部结束。
