+++
date = '2026-11-15T20:40:00+08:00'
draft = false
title = 'Gin 上传文件'
tags = ["Go", "Gin", "上传", "文件"]
categories = ["教程"]
summary = "单文件 FormFile 落盘、用位移限制 5MB 并校验后缀，以及 MultipartForm 一次接收多个文件"
weight = 24
+++

# Gin 上传文件

上传走的是 `multipart/form-data`：浏览器把文件拆成一段一段放进请求正文，服务端按**表单字段名**把文件取出来，再写到磁盘。

本篇按动手顺序来：

1. 单文件上传  
2. 校验大小和后缀  
3. 多文件上传  

上一篇的 `DefaultPostForm` 只能读普通表单字段，读不到文件。文件用 `FormFile` 或 `MultipartForm`。

---

## 1. 单文件上传

前端字段名是什么，服务端就按什么取。下面约定字段名是 `file`：

```html
<form action="/upload" method="post" enctype="multipart/form-data">
  <input type="file" name="file">
  <button type="submit">上传</button>
</form>
```

`enctype="multipart/form-data"` 必须写上，否则正文里没有文件。

```go
router.POST("/upload", func(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// etc/static/1.png → 1.png，只留文件名
	dst := filepath.Join("./files/", filepath.Base(file.Filename))
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.String(http.StatusInternalServerError, "保存失败")
		return
	}
	ctx.String(http.StatusOK, "上传成功")
})
```

| API | 作用 |
| --- | --- |
| `ctx.FormFile("file")` | 按字段名取出**一个**文件。返回的是 `*multipart.FileHeader`，里面有原始文件名和大小 |
| `file.Filename` | 客户端带来的文件名，可能带目录，例如 `etc/static/1.png` |
| `filepath.Base(file.Filename)` | 只留最后一段：`etc/static/1.png` → `1.png` |
| `filepath.Join("./files/", 文件名)` | 拼出落盘路径 `files/1.png`。分隔符按当前系统处理 |
| `ctx.SaveUploadedFile(file, dst)` | 把文件写到 `dst`。父目录不存在时会自己建 |

`Base` 要加上。客户端可以带上 `../` 或一整段路径，直接拿 `file.Filename` 去拼接，文件会写到 `./files/` 外面。`Base` 丢掉目录，只留下 `1.png` 这种名字，再和 `./files/` 拼在一起。

`SaveUploadedFile` 的错误要判断。保存失败时如果继续往下写「上传成功」，前端看到的结果是错的。

用 `curl` 试（`-F` 会自动带上 `multipart/form-data`）：

```bash
curl -F "file=@1.png" http://localhost:8080/upload
```

`file=` 对应 `FormFile("file")`，`@1.png` 表示把本地文件 `1.png` 传上去。成功后项目下会出现 `files/1.png`。

---

## 2. 校验大小和后缀

文件取出来之后、写盘之前做两道检查：太大的丢掉，后缀不在白名单里的丢掉。

### 大小：`5 << 20` 就是 5MB

`<<` 是左移。二进制往左挪 n 位，数值乘上 2 的 n 次方。内存里文件大小的单位是**字节**，所以用位移写出 KB、MB、GB：

| 写法 | 等于 | 含义 |
| --- | --- | --- |
| `1 << 10` | 1024 字节 | 1KB |
| `1 << 20` | 1024 × 1024 字节 | 1MB |
| `1 << 30` | 1024 × 1024 × 1024 字节 | 1GB |
| `5 << 20` | 5 × 1MB | 5MB |

左移的位数决定单位：`10` 是 KB，`20` 是 MB，`30` 是 GB。前面的数字是「几个」这种单位。`5 << 20` 就是 5 个 MB。

```go
const MaxSize = 5 << 20 // 5MB

if file.Size > MaxSize {
	ctx.String(http.StatusRequestEntityTooLarge, "文件过大")
	return
}
```

`file.Size` 是字节数。超过 5MB 回 **413**（`StatusRequestEntityTooLarge`），然后 `return`，后面的保存不再执行。

### 后缀：只允许 png 和 jpg

用 map 当白名单，查找是 O(1)，不在里面的键读出来是 `false`：

```go
var allowExt = map[string]bool{
	".png": true,
	".jpg": true,
}

ext := strings.ToLower(filepath.Ext(file.Filename))
if !allowExt[ext] {
	ctx.String(http.StatusBadRequest, "该后缀暂时不支持")
	return
}
```

`filepath.Ext` 带点，`"1.png"` 得到 `".png"`，所以白名单里要写成 `".png"`，不能写成 `"png"`。

后缀大小写不固定，Windows 上经常是 `.PNG`、`.JPG`。`Ext` 不会帮你转小写，比较前用 `strings.ToLower`，否则合法图片会被拒。

两道检查都放在 `SaveUploadedFile` 前面：

```go
router.POST("/upload", func(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if file.Size > MaxSize {
		ctx.String(http.StatusRequestEntityTooLarge, "文件过大")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowExt[ext] {
		ctx.String(http.StatusBadRequest, "该后缀暂时不支持")
		return
	}

	dst := filepath.Join("./files/", filepath.Base(file.Filename))
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.String(http.StatusInternalServerError, "保存失败")
		return
	}
	ctx.String(http.StatusOK, "上传成功")
})
```

| 请求 | 结果 |
| --- | --- |
| `1.png`，小于 5MB | `files/1.png`，响应「上传成功」 |
| 6MB 的 png | `413`，「文件过大」，磁盘上没有这个文件 |
| `1.gif` | `400`，「该后缀暂时不支持」 |
| `photo.JPG` | 转小写后是 `.jpg`，允许保存 |

---

## 3. 多文件上传

一次选好几个文件时，`FormFile` 只会拿到该字段的第一个。整包解析用 `MultipartForm`，再按字段名取出文件切片。

前端字段名改成 `files`，并加上 `multiple`：

```html
<input type="file" name="files" multiple>
```

```go
form, err := ctx.MultipartForm()
if err != nil {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	return
}

files := form.File["files"]
```

`form.File` 的类型是 `map[string][]*multipart.FileHeader`。键是字段名 `"files"`，值是这个字段上的全部文件。字段名和前端 `name` 不一致时，取出来是 `nil`，循环一次都不会进。

大小、后缀和单文件是同一套判断，放进循环里，每个文件单独看。有一个不合格就 `return`，不再保存后面的：

```go
router.POST("/upload", func(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.String(http.StatusBadRequest, "没有文件")
		return
	}

	for _, file := range files {
		if file.Size > MaxSize {
			ctx.String(http.StatusRequestEntityTooLarge, "文件过大")
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowExt[ext] {
			ctx.String(http.StatusBadRequest, "该后缀暂时不支持")
			return
		}

		dst := filepath.Join("./files/", filepath.Base(file.Filename))
		if err := ctx.SaveUploadedFile(file, dst); err != nil {
			ctx.String(http.StatusInternalServerError, "保存失败")
			return
		}
	}
	ctx.String(http.StatusOK, "上传成功")
})
```

循环结束后再回复「上传成功」。写在循环里面的话，第一个文件保存完就会结束响应。

两个文件一起传：

```bash
curl -F "files=@1.png" -F "files=@2.jpg" http://localhost:8080/upload
```

两次 `-F` 的字段名都是 `files`，对应 `form.File["files"]`。

### 合在一起

```go
package main

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 1 << 10 是 1KB，1 << 20 是 1MB，1 << 30 是 1GB
const MaxSize = 5 << 20 // 5MB

var allowExt = map[string]bool{
	".png": true,
	".jpg": true,
}

func main() {
	router := gin.Default()

	router.POST("/upload", func(ctx *gin.Context) {
		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			ctx.String(http.StatusBadRequest, "没有文件")
			return
		}

		for _, file := range files {
			if file.Size > MaxSize {
				ctx.String(http.StatusRequestEntityTooLarge, "文件过大")
				return
			}
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !allowExt[ext] {
				ctx.String(http.StatusBadRequest, "该后缀暂时不支持")
				return
			}

			dst := filepath.Join("./files/", filepath.Base(file.Filename))
			if err := ctx.SaveUploadedFile(file, dst); err != nil {
				ctx.String(http.StatusInternalServerError, "保存失败")
				return
			}
		}
		ctx.String(http.StatusOK, "上传成功")
	})

	if err := router.Run(); err != nil {
		panic(err)
	}
}
```

只传一个文件时，用同一个接口即可，字段名仍然是 `files`：

```bash
curl -F "files=@1.png" http://localhost:8080/upload
```

单文件那节的 `FormFile("file")` 适合确定只有一个文件的接口；多个文件用 `MultipartForm`。

---

## 小结

| 主题 | 记住什么 |
| --- | --- |
| 单文件 | `ctx.FormFile("file")`，字段名和前端 `name` 一致 |
| 落盘 | `filepath.Join("./files/", filepath.Base(file.Filename))`，`Base` 去掉客户端带来的目录 |
| 保存 | `ctx.SaveUploadedFile(file, dst)`，错误要 `return` |
| 大小 | `5 << 20` 是 5MB。`10` / `20` / `30` 分别对应 KB / MB / GB。超限回 413 |
| 后缀 | `filepath.Ext` 的结果带点，白名单写成 `".png"`。比较前 `strings.ToLower` |
| 多文件 | `ctx.MultipartForm()`，用 `form.File["files"]` 取出切片，循环里做同一套校验再保存 |
