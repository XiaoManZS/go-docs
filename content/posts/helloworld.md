+++
date = '2026-07-01T18:00:00+08:00'
draft = false
title = '第一个Go程序'
tags = ["Go", "helloworld"]
categories = ["教程"]
summary = "认识 Go 语言的特点，写出第一个 Hello World 程序。"
weight = 2
+++

# Hello World

我们可以新建一个`hello.go`文件，然后输入以下代码：

```go
package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
}
```

然后我们打开终端，输入`go run hello.go`命令，运行程序。

```bash
go run hello.go
```

输出> Hello, World!

## package main

```go
package main
```

1. package main = 告诉 Go 编译器，这是程序的入口包（可执行程序）。
2. Go 官方规定，程序入口必须叫 main，名字是固定的，不能修改。

例如我改成 `aaa` 会报错：

```go
package aaa

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

```bash
package command-line-arguments is not a main package # 错误提示
```

## import "fmt"

导入（Import）Go 标准库中的 fmt 包，让当前文件可以使用它提供的功能。

1. fmt：Go 官方提供的格式化输入输出包。
2. Println()：fmt 包中的一个函数，用来打印内容并换行。

## func main()

1. func main()：定义一个名为 main 的函数，这是 Go 程序的入口函数。
2. {}：main 函数的代码块，必须有。
3. fmt.Println("Hello, World!")：调用 fmt 包中的 Println 函数，打印 "Hello, World!" 字符串。

## 作业

1. 修改 `hello.go` 文件，将 "Hello, World!" 改成 "Hello, Go!"。
2. 运行程序，观察输出结果。


