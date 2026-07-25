+++
date = '2026-07-01T19:00:00+08:00'
draft = false
title = 'Go 循环'
tags = ["Go", "循环"]
categories = ["教程"]
summary = "详解Go循环：for、range、break、continue、goto"
weight = 6
+++

# 五谷轮回

Go 里**只有一种循环关键字：`for`**。没有 `while`、也没有 `do-while`，所有循环形态都靠 `for` 表达。

## 普通 for 循环

我们声明了一个切片，用经典的「初始化；条件；步进」写法遍历：

```go
package main

import "fmt"

func main() {
	var hobbys = []string{"吃饭", "睡觉", "打豆豆"}
	for i := 0; i < len(hobbys); i++ {
		fmt.Println(hobbys[i])
	}
}
```

`for` 后面分成三段，用分号隔开：

| 部分 | 作用 | 上面示例 |
| --- | --- | --- |
| 初始化 | 循环开始前执行一次 | `i := 0` |
| 条件 | 每次循环前判断，为真才继续 | `i < len(hobbys)` |
| 步进 | 每次循环体结束后执行 | `i++` |

输出依次为：`吃饭`、`睡觉`、`打豆豆`。

三段都可以省略。只留条件时，就相当于其他语言里的 `while`：

```go
package main

import "fmt"

func main() {
	i := 0
	for i < 3 {
		fmt.Println(i)
		i++
	}
}
```

条件也省略时，就是**无限循环**，需要靠 `break` 自己跳出：

```go
package main

import "fmt"

func main() {
	i := 0
	for {
		fmt.Println(i)
		i++
		if i >= 3 {
			break
		}
	}
}
```

## range 循环

遍历切片、数组、字符串、map、channel 时，更常用 `range`，写法更简洁：

```go
package main

import "fmt"

func main() {
	var hobbys = []string{"吃饭", "睡觉", "打豆豆"}
	for index, hobby := range hobbys {
		fmt.Println(index, hobby)
	}
}
```

`range` 每次迭代会返回两个值：

- 第一个是**下标**（或 map 的 key）
- 第二个是**对应的元素值**

输出为：

```text
0 吃饭
1 睡觉
2 打豆豆
```

如果只要值、不要下标，用 `_` 丢掉第一个返回值：

```go
for _, hobby := range hobbys {
	fmt.Println(hobby)
}
```

如果只要下标、不要值，写一个变量即可：

```go
for index := range hobbys {
	fmt.Println(index)
}
```

### 遍历 map

```go
package main

import "fmt"

func main() {
	person := map[string]string{
		"name": "张三",
		"city": "杭州",
	}
	for key, value := range person {
		fmt.Println(key, value)
	}
}
```

注意：map 的遍历**顺序不固定**，每次运行打印顺序可能不同。

### 遍历字符串

```go
package main

import "fmt"

func main() {
	text := "你好Go"
	for index, char := range text {
		fmt.Printf("下标 %d，字符 %c\n", index, char)
	}
}
```

用 `range` 遍历字符串时，拿到的是 **Unicode 码点（rune）**，不是单个字节。像「你」「好」这种汉字各占 3 个字节，下标会按字节位置跳着走。

| 写法 | 适用场景 |
| --- | --- |
| `for i := 0; i < len(x); i++` | 需要精确控制下标、步长，或从中间开始/倒序遍历 |
| `for i, v := range x` | 顺序遍历切片、数组、字符串、map |

## break 与 continue

循环中经常要「提前结束」或「跳过本次」：

| 关键字 | 含义 |
| --- | --- |
| `break` | 立刻跳出**当前**循环 |
| `continue` | 跳过本次剩余代码，进入下一次循环 |

### break：跳出循环

```go
package main

import "fmt"

func main() {
	hobbys := []string{"吃饭", "睡觉", "打豆豆"}
	for _, hobby := range hobbys {
		if hobby == "睡觉" {
			break
		}
		fmt.Println(hobby)
	}
}
```

遇到「睡觉」就 `break`，后面的元素不再遍历。输出只有：`吃饭`。

### continue：跳过本次

```go
package main

import "fmt"

func main() {
	hobbys := []string{"吃饭", "睡觉", "打豆豆"}
	for _, hobby := range hobbys {
		if hobby == "睡觉" {
			continue
		}
		fmt.Println(hobby)
	}
}
```

遇到「睡觉」就 `continue`，跳过打印，但循环继续。输出为：`吃饭`、`打豆豆`。

### 带标签的 break

多层循环嵌套时，普通 `break` 只跳出最内层。若要一次跳出外层，可以给循环加**标签**：

```go
package main

import "fmt"

func main() {
Outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				break Outer
			}
			fmt.Println(i, j)
		}
	}
}
```

`break Outer` 会直接跳出名为 `Outer` 的那一层循环。`continue` 也可以配合标签使用，表示跳到指定循环的下一次迭代。

## goto

`goto` 可以跳转到同一函数内的某个标签位置。能用，但可读性往往更差，日常代码里更推荐 `break` / `continue` / `return`。

```go
package main

import "fmt"

func main() {
	i := 0
Loop:
	fmt.Println(i)
	i++
	if i < 3 {
		goto Loop
	}
}
```

上面用 `goto` 模拟了一个简单循环：打印 `0`、`1`、`2`。实际项目中除非处理错误清理这类特殊场景，否则尽量少用。

## 作业

1. **普通 for**：声明一个 `1` 到 `10` 的循环，打印所有偶数。
2. **range 练习**：定义切片 `[]string{"春", "夏", "秋", "冬"}`，分别用「只要下标」「只要值」「下标+值」三种 `range` 写法打印。
3. **break / continue**：遍历 `1` 到 `20`，遇到能被 `3` 整除的数就 `continue` 跳过；一旦遇到大于 `15` 的数就 `break` 结束。观察最终打印了哪些数。
4. **类 while**：用「只写条件」的 `for`（不要初始化段和步进段）计算 `1 + 2 + ... + 100` 的和并打印。
5. **综合练习**：定义 `map[string]int` 表示各科分数（如语文、数学、英语），用 `range` 遍历：低于 60 的科目打印「不及格」，其余打印「及格」，并统计及格科目数量。
