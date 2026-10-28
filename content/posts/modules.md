+++
date = '2026-10-28T18:30:00+08:00'
draft = false
title = 'Go Modules 与项目结构'
tags = ["Go", "go mod", "模块", "包"]
categories = ["教程"]
summary = "go mod init / tidy、自定义包与导入路径、导出规则、cmd/internal/pkg、引入第三方依赖"
weight = 21
+++

# Go Modules

前面写的都是「单文件 `package main`」。真正做项目时，代码要拆成多个包、依赖要版本可控——这就靠 **Go Modules**。

本篇按动手顺序来：

1. `go mod init` 建模块  
2. `go mod tidy` 整理依赖  
3. 自定义包和导入路径  
4. 大写导出 / 小写私有  
5. `cmd`、`internal`、`pkg` 简单目录  
6. 引入一个第三方依赖演示  

---

## 1. go mod init：创建模块

在空目录里初始化：

```bash
mkdir hello-mod
cd hello-mod
go mod init example.com/hello
```

会生成 `go.mod`：

```go
module example.com/hello

go 1.26.4
```

| 字段 | 含义 |
| --- | --- |
| `module example.com/hello` | **模块路径**，也是包的导入前缀 |
| `go 1.22` | 本模块使用的 Go 语言版本 |

模块路径习惯写成「能当导入前缀」的名字，常见两种：

- 学习演示：`example.com/hello`、`myapp`  
- 真仓库：`github.com/你的用户名/仓库名`

> `go.mod` 管的是**整个模块**（一个仓库通常一个），不是每个子目录一个。

---

## 2. go mod tidy：整理依赖

写完代码、增删 `import` 之后执行：

```bash
go mod tidy
```

它会：

- 把代码里**真正用到**的依赖写进 `go.mod`，并生成/更新 `go.sum`（校验和）  
- 删掉**已经不用**的依赖  

记住分工：

| 命令 | 什么时候用 |
| --- | --- |
| `go mod init` | 项目一开始，只跑一次 |
| `go mod tidy` | 改过 import / 依赖后，随时跑 |

`go.sum` 不要手改，提交到 Git 即可，用来保证大家下到同一份依赖。

---

## 3. 自定义包和导入路径

目录名 ≈ 包名；导入路径 = **模块路径 + 子目录**。

```text
hello-mod/
├── go.mod                 # module example.com/hello
├── main.go                # package main
└── greet/
    └── greet.go           # package greet
```

`greet/greet.go`：

```go
package greet

import "fmt"

func Hello(name string) {
	fmt.Println("你好,", name)
}
```

根目录 `main.go`：

```go
package main

import "example.com/hello/greet" // 模块路径 + 子目录

func main() {
	greet.Hello("小满")
}
```

运行：

```bash
go run .
# 你好, 小满
```

要点：

- 导入写的是**路径**，不是文件名：`example.com/hello/greet`  
- 调用时用**包名**：`greet.Hello(...)`（一般等于最后一级目录名）  
- 同一个目录里所有 `.go` 必须是同一个 `package` 名  

---

## 4. 大写导出 / 小写私有

Go **没有** `public` / `private` 关键字，靠首字母大小写：

| 写法 | 可见范围 |
| --- | --- |
| `Hello`、`User`、`MaxRetry` | **导出**：别的包可以访问 |
| `hello`、`user`、`maxRetry` | **未导出**：仅本包内部可用 |

```go
package greet

import "fmt"

// 导出：外部可调用
func Hello(name string) {
	fmt.Println(prefix(), name)
}

// 未导出：外部写 greet.prefix() 会编译失败
func prefix() string {
	return "你好,"
}
```

外面只能：

```go
greet.Hello("小满") // OK
// greet.prefix()  // 报错：prefix 未导出
```

结构体字段同样规则：`Name` 导出，`age` 仅包内可见（JSON 标签、跨包赋值都会受影响）。

---

## 5. cmd、internal、pkg 简单目录

小项目一个 `greet` 包就够。稍大一点，社区常见这种骨架：

```text
hello-mod/
├── go.mod
├── cmd/
│   └── hello/
│       └── main.go      # 可执行入口：go run ./cmd/hello
├── internal/
│   └── auth/
│       └── auth.go      # 仅本模块能导入（编译器强制）
└── pkg/
    └── greeter/
        └── greeter.go   # 可被外部模块导入的库代码
```

| 目录 | 放什么 | 谁能 import |
| --- | --- | --- |
| `cmd/<名字>/` | `package main`，程序入口 | 不作为库被导入，用来 `go build` / `go run` |
| `internal/...` | 模块内部实现 | **只有本模块**（及 `internal` 的父树）能导入；外人导入会编译失败 |
| `pkg/...` | 打算给别人用的库 | 任意模块都可导入 |

`cmd/hello/main.go` 示例：

```go
package main

import (
	"example.com/hello/internal/auth"
	"example.com/hello/pkg/greeter"
)

func main() {
	auth.Check()
	greeter.Hello("小满")
}
```

`internal/auth/auth.go`：

```go
package auth

import "fmt"

func Check() {
	fmt.Println("auth ok")
}
```

`pkg/greeter/greeter.go`：

```go
package greeter

import "fmt"

func Hello(name string) {
	fmt.Println("你好,", name)
}
```

说明：

- `internal` 是语言级约定：别的仓库 `import "example.com/hello/internal/auth"` 会直接报错  
- `pkg` 不是强制的，很多项目把可导出库直接放在模块根下；放 `pkg` 只是「这里是对外 API」更清晰  
- 入门项目不必硬套三层；**有多个可执行文件再用 `cmd`，有不想被外人用的代码再放 `internal`**

---

## 6. 引入第三方依赖演示

### 先配置国内代理（重要）

默认的官方代理 `proxy.golang.org` 在国内经常超时或拉不下来。拉第三方库之前，建议先把模块代理改成国内镜像。

推荐使用七牛云维护的 **[https://goproxy.cn](https://goproxy.cn)**：

```bash
# 永久生效（推荐）
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

验证是否设置成功：

```bash
go env GOPROXY
# 应输出：https://goproxy.cn,direct
```

说明：

| 写法 | 含义 |
| --- | --- |
| `https://goproxy.cn` | 优先走国内代理下载模块 |
| `,direct` | 代理没有时，再直连源站（如 GitHub） |

也可以临时设置（只对当前终端有效）：

```bash
# Windows PowerShell
$env:GOPROXY = "https://goproxy.cn,direct"

# macOS / Linux
export GOPROXY=https://goproxy.cn,direct
```

其他可选镜像（备用）：

| 代理地址 | 说明 |
| --- | --- |
| `https://goproxy.cn,direct` | **首选**，七牛云，国内最常用 |
| `https://proxy.golang.com.cn,direct` | 官方域名的国内镜像站 |
| `https://goproxy.io,direct` | 另一家公共代理 |

公司私有仓库不要走公共代理时，可再加：

```bash
go env -w GOPRIVATE=git.公司域名.com
```

### 演示：安装并使用第三方库

以常用的颜色输出库 [`github.com/fatih/color`](https://github.com/fatih/color) 为例。

在 `main.go`（或 `cmd/hello/main.go`）里先写上 import 和用法：

```go
package main

import (
	"github.com/fatih/color"
)

func main() {
	color.Green("go mod 依赖引入成功")
	color.Red("这是红色警告")
}
```

然后拉依赖并整理：

```bash
go get github.com/fatih/color
go mod tidy
```

`go.mod` 里会出现类似：

```go
module example.com/hello

go 1.22

require github.com/fatih/color v1.18.0
```

再运行：

```bash
go run .
# 终端里会看到绿色 / 红色文字
```

常用依赖命令：

| 命令 | 作用 |
| --- | --- |
| `go get 路径@版本` | 添加或升级某个依赖，如 `@v1.18.0`、`@latest` |
| `go mod tidy` | 按实际 import 增删依赖 |
| `go list -m all` | 查看当前模块解析到的依赖树 |

换依赖时：改代码里的 `import` → `go get`（可选）→ `go mod tidy`。

---

## 小结

| 主题 | 记住什么 |
| --- | --- |
| `go mod init` | 定模块路径，生成 `go.mod` |
| `go mod tidy` | 按代码同步依赖与 `go.sum` |
| 导入路径 | `模块路径 + 子目录` |
| 导出规则 | 大写对外，小写包内 |
| `cmd` / `internal` / `pkg` | 入口 / 私有实现 / 对外库 |
| 第三方库 | 先配 `GOPROXY=https://goproxy.cn,direct`，再 `go get` + import + `tidy` |

最小心智模型：

> **一个 `go.mod` = 一个模块；一个目录 = 一个包；导入靠路径，可见性靠大小写。**
