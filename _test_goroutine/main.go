package main

import (
	"fmt"
	"sync"
	"time"
)

func notify(orderID string, cost time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(cost)
	fmt.Println("已通知订单:", orderID)
}

func sumScores(name string, scores []int, wg *sync.WaitGroup) {
	defer wg.Done()
	total := 0
	for _, s := range scores {
		total += s
	}
	fmt.Printf("%s 总分: %d\n", name, total)
}

func main() {
	var wg sync.WaitGroup
	orders := []struct {
		id   string
		cost time.Duration
	}{
		{"A1001", 50 * time.Millisecond},
		{"A1002", 30 * time.Millisecond},
		{"A1003", 40 * time.Millisecond},
	}
	wg.Add(len(orders))
	for _, o := range orders {
		go notify(o.id, o.cost, &wg)
	}
	wg.Wait()
	fmt.Println("全部通知完成")

	students := map[string][]int{
		"小满": {90, 85, 88},
		"张三": {70, 92, 80},
		"李四": {95, 91, 89},
	}
	wg.Add(len(students))
	for name, scores := range students {
		n, s := name, scores
		go sumScores(n, s, &wg)
	}
	wg.Wait()
	fmt.Println("统计结束")
}
