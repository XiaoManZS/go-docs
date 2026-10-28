+++
date = '2026-10-28T01:00:00+00:00'
draft = false
title = 'Go 反射'
tags = ["Go", "反射", "reflect"]
categories = ["教程"]
summary = "详解 Go 反射：类型判断、修改值与类型断言、动态调用方法"
weight = 20
+++

# 反射

上一篇讲了**自定义类型**。类型在编译期就定死了——那运行时怎么「看」一个值到底是什么类型、能不能改、有没有某个方法？

答案是标准库 `reflect`：**在运行时检查类型、读写值、调用方法**。框架、ORM、JSON 编解码底层都大量用到它。

> 反射灵活，但代码更绕、也更慢。日常业务能用接口 / 泛型解决的，优先别上反射。

下面用同一个自定义类型 `Age`，分三个案例：

```go
type Age int
```

---

## 1. 用反射做判断

先拿到类型信息，再看「叫什么名」「底层是什么」。

```go
package main

import (
	"fmt"
	"reflect"
)

type Age int

func main() {
	var age Age = 99
	t := reflect.TypeOf(age)

	fmt.Println(t.Name()) // Age —— 类型名（自定义类型的名字）
	fmt.Println(t.Kind()) // int —— 底层种类（拿来做判断）

	if t.Kind() == reflect.Int {
		fmt.Println("他绝对是一个 int 类型")
	}
}
```

| API | 作用 | 本例结果 |
| --- | --- | --- |
| `reflect.TypeOf(x)` | 拿到 `x` 的类型描述 | `Age` 的 `reflect.Type` |
| `t.Name()` | 类型名称 | `"Age"` |
| `t.Kind()` | 底层种类（int / string / struct…） | `reflect.Int` |

`Name` 和 `Kind` 别混：

- `Name()`：你起的名字（`Age`）
- `Kind()`：真正的底层形态（`int`）

做分支判断时，几乎总是比 `Kind()`，例如 `reflect.Int`、`reflect.String`、`reflect.Struct`。

---

## 2. 通过反射修改值 + 类型断言

想改原变量，必须传**指针**，再用 `.Elem()` 解引用，得到可写的 `reflect.Value`。

```go
package main

import (
	"fmt"
	"reflect"
)

type Age int

func main() {
	var age Age = 99

	// 必须传 &age，再 Elem()，否则 CanSet() 为 false，改不了
	val := reflect.ValueOf(&age).Elem()

	// val.CanSet()  是否允许修改
	// val.IsValid() 是否是有效的 reflect.Value
	if val.CanSet() && val.IsValid() && val.Kind() == reflect.Int {
		val.SetInt(10086) // 通过反射写入
	}

	// 类型断言：Interface() 先变成 any，再 .(Age) 还原成具体类型
	v := val.Interface().(Age)

	fmt.Println(val) // 10086（reflect.Value，打印时会显示内部值）
	fmt.Println(age) // 10086（原变量已被改掉）
	fmt.Println(v)   // 10086（断言回来的 Age）
}
```

要点：

1. **`reflect.ValueOf(age)` 得到的是副本**，`CanSet()` 为 `false`，`SetInt` 会 panic  
2. **`reflect.ValueOf(&age).Elem()`** 才指向原变量，可以 `SetInt` / `SetString` / `Set`  
3. 改之前先检查：`CanSet()`、`IsValid()`、`Kind()` 对不对  
4. **类型断言**：`val.Interface().(Age)` 把反射世界里的值拿回静态类型，之后就能当普通 `Age` 用

更稳妥的断言写法（带 `ok`）：

```go
v, ok := val.Interface().(Age)
if !ok {
	fmt.Println("断言失败")
	return
}
fmt.Println(v)
```

---

## 3. 通过反射调用方法

给 `Age` 挂一个方法，再用方法名动态调用。

```go
package main

import (
	"fmt"
	"reflect"
)

type Age int

func (a Age) Say(params string, n int) {
	fmt.Println(params, n)
}

func main() {
	var age Age = 99
	val := reflect.ValueOf(age)

	m := val.MethodByName("Say") // 按名字找方法（必须导出，首字母大写）
	if m.IsValid() {
		m.Call([]reflect.Value{
			reflect.ValueOf("我是字符串"),
			reflect.ValueOf(123123),
		})
		// 输出：我是字符串 123123
	}
}
```

流程拆开看：

| 步骤 | 代码 | 含义 |
| --- | --- | --- |
| 1 | `reflect.ValueOf(age)` | 拿到值的反射包装 |
| 2 | `MethodByName("Say")` | 按名字找方法；找不到则 `IsValid()` 为 false |
| 3 | `m.Call([]reflect.Value{...})` | 传入参数列表（每个参数都要先 `ValueOf`） |

注意：

- 方法名必须**导出**（大写开头），否则 `MethodByName` 找不到  
- `Call` 的参数个数、类型要和方法签名一致，否则 panic  
- 有返回值时，`Call` 会返回 `[]reflect.Value`，再按需 `Interface()` 取出来

---

## 小结

| 案例 | 核心 API | 用途 |
| --- | --- | --- |
| 判断 | `TypeOf` / `Name` / `Kind` | 运行时识别类型 |
| 改值 | `ValueOf(&x).Elem()` + `SetInt` + `Interface().(T)` | 修改原变量，再断言回具体类型 |
| 调方法 | `MethodByName` + `Call` | 按名字动态调用 |

三板斧记牢：

1. **看类型**用 `Type`，**动数据**用 `Value`  
2. **要改值就传指针再 `Elem()`**  
3. **反射进出都要装箱/拆箱**：进去 `ValueOf`，出来 `Interface()` + 类型断言
