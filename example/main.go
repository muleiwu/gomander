package main

import (
	"fmt"
	"time"

	"github.com/muleiwu/gomander"
)

func main() {
	// 使用 gomander 运行业务逻辑
	gomander.Run(func() {
		fmt.Println("应用程序启动...")
		fmt.Println("开始执行业务逻辑")

		// 模拟长时间运行的服务
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		count := 0
		for {
			select {
			case <-ticker.C:
				count++
				fmt.Printf("[%s] 心跳 #%d\n", time.Now().Format("2006-01-02 15:04:05"), count)
			}
		}
	}, gomander.WithPidFile("./myapp.pid"), gomander.WithLogFile("./myapp.log"))
}
