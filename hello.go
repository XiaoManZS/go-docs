package main

import "fmt"

func main() {
	var person = map[string]string{
		"name": "张三",
		"age":  "18",
		"sex":  "男",
	}
	fmt.Println(len(person))
}
