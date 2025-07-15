package main

import (
	"fmt"
	"math/rand"
	run "runtime"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano() + int64(run.NumGoroutine()) + time.Now().UnixMicro())

	// 生成范围 30000-32760 的随机数
	min := 30000
	max := 32760
	randomNum := min + rand.Intn(max-min+1) // +1 包含上限值
	fmt.Printf("ran:%v", randomNum)
}
