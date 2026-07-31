+++
date = '2026-07-01T19:00:00+08:00'
draft = false
title = 'Go 类型转换'
tags = ["Go", "类型转换"]
categories = ["教程"]
summary = "Go 基础类型转换"
weight = 4
+++

# 类型转换

实际开发里，不同类型之间经常需要互相转换：字符串转数字、数字转字符串、浮点数转整数、整数转浮点数，等等。

### 数字的转换

比如有一个浮点数 `3.1415926535`，想把它变成整数，可以用 `int()` 来转换：

```go
package main
import "fmt"

func main() {
	num := 3.1415926535
	intNum := int(num)
	fmt.Println(intNum) // 3
}
```

可以看到，`int()` 会把浮点数转成整数，并直接舍去小数部分（不是四舍五入）。

其他数值类型的转换写法也类似，记起来很简单——跟定义类型时的名字一模一样：

- `int()`：转为 `int`
- `int8()`：转为 `int8`
- `int16()`：转为 `int16`
- `int32()`：转为 `int32`
- `int64()`：转为 `int64`
- `uint()`：转为 `uint`
- `uint8()`：转为 `uint8`
- `uint16()`：转为 `uint16`
- `uint32()`：转为 `uint32`
- `uint64()`：转为 `uint64`
- `float32()`：转为 `float32`
- `float64()`：转为 `float64`

### 数字转字符串

数字和字符串之间的转换要用 `strconv`（string convert，字符串转换）包。比如把整数 `3` 转成字符串：

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	num := 3
	strNum := strconv.Itoa(num)
	fmt.Printf("类型: %T,值: %v", strNum, strNum)
	// 类型: string,值: 3
}
```

几点说明：

1. 导入多个包时，用 `()` 分组，每个包独占一行。
2. `strconv.Itoa()` 就是 **Int to ASCII**：把整数转成字符串（字符串本质上是字符序列）。
3. `fmt.Printf()` 里，`%T` 打印类型，`%v` 打印值。

### 字符串转数字

反过来，字符串转数字也很常见。Go 里字符串和数字不能直接运算，必须先转成同一类型。比如这样会报错：

```go
var a = "10"
var b = 20
var c = a + b // 错误：字符串和数字不能直接运算
```

这时可以用 `strconv.Atoi()` 把字符串转成整数：

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	str := "10"
	num, _ := strconv.Atoi(str)
	fmt.Printf("类型: %T,值: %v", num, num)
	// 类型: int,值: 10
}
```

`Atoi` 是 **ASCII to Int** 的缩写。它会返回两个值：转换结果和错误。转换失败时第二个返回值不为 `nil`；这里用 `_` 忽略错误，表示暂时不处理失败情况。

### 字符串转布尔值

字符串转布尔用 `strconv.ParseBool()`。它同样返回「结果 + 错误」两个值：

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	str := "true"
	b, _ := strconv.ParseBool(str)
	fmt.Printf("类型: %T,值: %v", b, b)
	// 类型: bool,值: true
}
```

`ParseBool` 能识别的真值有：`1`、`t`、`T`、`TRUE`、`true`、`True`；假值有：`0`、`f`、`F`、`FALSE`、`false`、`False`。其他字符串都会转换失败。

### 布尔值转字符串

反过来，用 `strconv.FormatBool()` 把布尔值转成字符串，结果只会是 `"true"` 或 `"false"`：

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	flag := true
	str := strconv.FormatBool(flag)
	fmt.Printf("类型: %T,值: %v", str, str)
	// 类型: string,值: true
}
```
