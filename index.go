package main

import "fmt"

func main() {
	var hobbys = []string{"吃饭", "睡觉", "打豆豆"}
	for i := 0; i < len(hobbys); i++ {
		fmt.Println(hobbys[i], i)
	}
}
