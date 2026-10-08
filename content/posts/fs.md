+++
date = '2026-10-08T23:30:00+08:00'
draft = false
title = 'Go 文件与目录操作'
tags = ["Go", "文件", "os", "OpenFile", "权限"]
categories = ["教程"]
summary = "详解 Go 文件读写删除、os.OpenFile 的 flag 与权限位，以及常见目录操作"
weight = 17
+++

# 文件与目录操作

日常写服务、写脚本，免不了和磁盘打交道：建文件、写内容、读出来、删掉，再进阶一点就是目录。  
这些能力主要在标准库 **`os`**（以及一点 `io`）里。

入门先记住五件事：

1. 简单读写可用 `os.WriteFile` / `os.ReadFile`。
2. 更精细的控制用 **`os.OpenFile`**（flag + 权限）。
3. 文件用完要 **`Close`**，习惯写 `defer f.Close()`。
4. Unix 权限用八进制，如 `0644`、`0755`；**在 Windows 上这些权限位基本无效**。
5. 目录用 `Mkdir` / `MkdirAll`，删目录用 `Remove` / `RemoveAll`。

> 口诀：简单读写用 `ReadFile` / `WriteFile`；要追加、排他创建、自定义权限，再上 `OpenFile`。

## 创建文件

最直接：`os.Create`。文件不存在就创建；**已存在会清空成 0 字节**（相当于截断）。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("note.txt")
	if err != nil {
		fmt.Println("创建失败:", err)
		return
	}
	defer f.Close()

	fmt.Println("创建成功:", f.Name())
}
```

要点：

| 写法 | 含义 |
| --- | --- |
| `os.Create(name)` | 创建或截断文件，返回 `*os.File` |
| `defer f.Close()` | 函数结束时关文件，避免句柄泄漏 |
| `err != nil` | 磁盘满、没权限、路径非法时都会报错 |

> `Create` 底层大致等于：`OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)`。

## 写入内容

### 方式一：拿到 `*File` 再写

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("note.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	n, err := f.WriteString("你好，小满\n")
	if err != nil {
		panic(err)
	}
	fmt.Println("写入字节数:", n)

	_, err = f.Write([]byte("第二行：奶茶菜单\n"))
	if err != nil {
		panic(err)
	}
}
```

| 写法 | 含义 |
| --- | --- |
| `WriteString(s)` | 写字符串 |
| `Write(b)` | 写字节切片 |
| 返回值 `n` | 实际写入的字节数 |

### 方式二：一行写完（更常用）

```go
err := os.WriteFile("note.txt", []byte("你好，小满\n"), 0644)
if err != nil {
	panic(err)
}
```

`WriteFile` 会：创建/截断 → 写入 → 关闭。适合「整份内容一次写完」。

## 读取详情

### 一次读完整文件

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("note.txt")
	if err != nil {
		fmt.Println("读取失败:", err)
		return
	}
	fmt.Println(string(data))
}
```

### 打开后再读（适合大文件、边读边处理）

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("note.txt") // 只读打开；不存在会报错
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 64)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
}
```

### 看文件元信息：`Stat`

```go
info, err := os.Stat("note.txt")
if err != nil {
	panic(err)
}
fmt.Println("名字:", info.Name())
fmt.Println("大小:", info.Size(), "字节")
fmt.Println("是目录吗:", info.IsDir())
fmt.Println("权限:", info.Mode())
fmt.Println("修改时间:", info.ModTime())
```

| API | 用途 |
| --- | --- |
| `os.ReadFile` | 小文件一次性读完 |
| `os.Open` | 只读打开，再 `Read` |
| `os.Stat` | 大小、权限、是否目录、修改时间 |

## 删除文件

```go
err := os.Remove("note.txt")
if err != nil {
	fmt.Println("删除失败:", err)
	return
}
fmt.Println("已删除")
```

| API | 行为 |
| --- | --- |
| `os.Remove(path)` | 删文件；也可删**空目录** |
| `os.RemoveAll(path)` | 递归删除：文件、非空目录都能删（慎用） |

文件不存在时，`Remove` 会返回错误（可用 `os.IsNotExist(err)` 判断）。

---

## 高级：`os.OpenFile`

当「只创建 / 只读 / 只写」不够用时，上 **`OpenFile`**：

```go
func OpenFile(name string, flag int, perm FileMode) (*File, error)
```

三个参数：

1. **`name`**：路径  
2. **`flag`**：怎么打开（可读？追加？不存在是否创建？）——常用常量按位或 `|` 组合  
3. **`perm`**：新建文件时的权限（如 `0644`）——**主要对 Unix/Linux/macOS 有意义**

示例：追加写入（文件不存在就创建）：

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString("又来了一单奶茶\n")
	if err != nil {
		panic(err)
	}
	fmt.Println("追加成功")
}
```

### flag 常用枚举（可 `|` 组合）

| 常量 | 含义 |
| --- | --- |
| `os.O_RDONLY` | 只读 |
| `os.O_WRONLY` | 只写 |
| `os.O_RDWR` | 读写 |
| `os.O_CREATE` | 不存在则创建 |
| `os.O_APPEND` | 写的时候追加到末尾 |
| `os.O_TRUNC` | 打开时把文件截断为 0 字节 |
| `os.O_EXCL` | 与 `O_CREATE` 一起用：文件**已存在则失败**（排他创建） |
| `os.O_SYNC` | 尽量同步写到磁盘（更安全，更慢） |

读写权限类（`O_RDONLY` / `O_WRONLY` / `O_RDWR`）通常**三选一**；其它用 `|` 叠上去：

```go
// 读写 + 不存在则创建 + 追加
os.O_RDWR | os.O_CREATE | os.O_APPEND

// 只写 + 不存在则创建 + 每次打开清空（覆盖写）
os.O_WRONLY | os.O_CREATE | os.O_TRUNC

// 创建但已存在就报错（防止误覆盖）
os.O_WRONLY | os.O_CREATE | os.O_EXCL
```

和快捷 API 的对应关系（帮助记忆）：

| 快捷写法 | 大致等价 |
| --- | --- |
| `os.Open(name)` | `OpenFile(name, O_RDONLY, 0)` |
| `os.Create(name)` | `OpenFile(name, O_RDWR\|O_CREATE\|O_TRUNC, 0666)` |

### 权限 `perm`：`0777` / `0644` 到底什么意思

`perm` 是 **Unix 文件模式**，Go 里常写成**八进制**字面量：以 **`0` 开头**，如 `0644`、`0755`、`0777`。

#### 先记三个基本位：`r` `w` `x`

每一位权限用一个数字相加（不是 4、3、2）：

| 符号 | 含义 | 数值 |
| --- | --- | --- |
| **r** | read 读 | **4** |
| **w** | write 写 | **2** |
| **x** | execute 执行（对目录表示「能否进入」） | **1** |

所以一个人的权限可以是：

| 组合 | 计算 | 常见读法 |
| --- | --- | --- |
| `---` | 0 | 什么都不能 |
| `--x` | 1 | 只能执行 |
| `-w-` | 2 | 只能写 |
| `-wx` | 2+1=3 | 写 + 执行 |
| `r--` | 4 | 只能读 |
| `r-x` | 4+1=5 | 读 + 执行 |
| `rw-` | 4+2=6 | 读 + 写 |
| `rwx` | 4+2+1=**7** | 全部可以 |

#### 再记三段人：所有者、组、其他人

一个完整权限是 **三组** `rwx`，从左到右：

```text
0  7  7  7
│  │  │  └─ 其他人（other）
│  │  └──── 所属组（group）
│  └─────── 所有者（owner / user）
└────────── 八进制前缀（Go 字面量写法）
```

| 位置 | 是谁 |
| --- | --- |
| 左数第一位数字 | **文件所有者**（创建者账号） |
| 中间一位 | **所属组**（同一组的用户） |
| 右边一位 | **其他人**（既不是主人也不是同组） |

例子：

| 权限 | 拆解 | 含义 |
| --- | --- | --- |
| `0644` | owner=`rw-`(6)，group=`r--`(4)，other=`r--`(4) | 自己可读写；别人只读。**文件很常见** |
| `0755` | owner=`rwx`(7)，group=`r-x`(5)，other=`r-x`(5) | 自己全能；别人可读可进。**目录 / 可执行文件很常见** |
| `0777` | 三段都是 `rwx`(7) | **所有人都能读、写、执行**——方便但危险，生产少用 |
| `0600` | 只有 owner=`rw-` | 只有自己能读写，适合密钥、token |

对照 Linux 上常见显示：

```text
-rw-r--r--   →  0644
drwxr-xr-x   →  0755
-rwxrwxrwx   →  0777
```

最左边的 `-` / `d` 表示普通文件还是目录，不算在三位数字里。

#### 重要：在 Windows 上这些权限基本无效

Go 在 Windows 上**仍然要你传 `perm`**，但系统**不会按 Unix 那套 owner/group/other 去落实**。

入门可以记：

- 写跨平台代码时，`0644`、`0755` 照常写，Linux/macOS 上生效。
- 到了 **Windows**，别指望 `0777`、`0644` 真的控制「谁能读谁能写」。
- Windows 更认自己的 ACL 权限模型；Go 的 `perm` 在这里往往被忽略或只产生很有限影响。

所以：  
**权限位是 Unix 世界的规则；Windows 上别拿 `0777` 当安全配置。**

---

## 文件夹（目录）操作

文件会了，目录也是同一套 `os` 思路。

### 创建目录

```go
// 只创建一层；父目录不存在会失败
err := os.Mkdir("data", 0755)

// 递归创建：a/b/c 一层层建好
err = os.MkdirAll("data/logs/2026", 0755)
```

| API | 行为 |
| --- | --- |
| `os.Mkdir(path, perm)` | 创建单层目录 |
| `os.MkdirAll(path, perm)` | 递归创建；已存在则通常不报错 |

目录权限常用 **`0755`**：自己可进可改，别人可进可列，不能随便改。

### 读目录里有什么

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	entries, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		kind := "文件"
		if e.IsDir() {
			kind = "目录"
		}
		fmt.Println(kind, e.Name())
	}
}
```

`ReadDir` 返回 `[]fs.DirEntry`，比老的 `Readdir` 更轻量，入门优先用它。

### 删除 / 重命名 / 判断存在

```go
// 删空目录（里面还有东西会失败）
os.Remove("data")

// 递归删整个目录树（危险，先确认路径）
os.RemoveAll("data")

// 重命名或移动
os.Rename("note.txt", "note.bak.txt")

// 判断路径是否存在
_, err := os.Stat("note.txt")
if os.IsNotExist(err) {
	fmt.Println("不存在")
} else if err != nil {
	fmt.Println("其它错误:", err)
} else {
	fmt.Println("存在")
}
```

| API | 用途 |
| --- | --- |
| `os.Remove` | 删文件或**空**目录 |
| `os.RemoveAll` | 递归删除 |
| `os.Rename` | 改名 / 移动 |
| `os.Stat` + `os.IsNotExist` | 判断在不在 |

### 小综合：建目录 → 写文件 → 读回来 → 清理

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := "demo_data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	path := filepath.Join(dir, "menu.txt")
	content := []byte("芋泥波波\n绿茶\n红茶\n")

	if err := os.WriteFile(path, content, 0644); err != nil {
		panic(err)
	}

	// 追加一行
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	_, _ = f.WriteString("爷爷不泡茶\n")
	f.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("文件内容:\n" + string(data))

	// 演示完清理
	_ = os.RemoveAll(dir)
}
```

`path/filepath.Join` 会按当前系统拼路径（Windows 用 `\`，Linux 用 `/`），比手写 `"a/" + "b"` 更稳。

## 速查表

| 你想做的事 | 优先用 |
| --- | --- |
| 整文件写入 | `os.WriteFile` |
| 整文件读取 | `os.ReadFile` |
| 追加日志 | `OpenFile(..., O_CREATE\|O_WRONLY\|O_APPEND, 0644)` |
| 防止覆盖已有文件 | `O_CREATE\|O_EXCL` |
| 看大小 / 是否目录 | `os.Stat` |
| 删一个文件 | `os.Remove` |
| 建多层目录 | `os.MkdirAll` |
| 列目录 | `os.ReadDir` |
| 删整棵目录树 | `os.RemoveAll`（慎用） |

入门阶段先把 **`WriteFile` / `ReadFile` / `Remove` / `MkdirAll` / `OpenFile+APPEND`** 练熟；权限先记住文件 `0644`、目录 `0755`，并清楚 **Windows 上 Unix 权限位基本当摆设**。

## 作业

1. 用 `WriteFile` 写一个 `hello.txt`，再用 `ReadFile` 打印出来。  
2. 用 `OpenFile` + `O_APPEND` 往同一文件追加 3 行，确认不会覆盖旧内容。  
3. 用 `O_CREATE|O_EXCL` 创建文件两次：第二次应失败。  
4. 手算并说明：`0755`、`0644`、`0777` 对 owner / group / other 分别意味着什么。  
5. `MkdirAll("a/b/c", 0755)`，在 `c` 下写文件，`ReadDir` 列出，最后 `RemoveAll("a")` 清理。
