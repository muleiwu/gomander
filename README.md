# Gomander

Gomander 是一个 Go 进程守护化库，让你的程序轻松支持前台运行和后台 daemon 模式。

## 功能特性

- ✅ 前台阻塞运行
- ✅ 后台守护进程模式（`-d` 参数）
- ✅ PID 文件管理
- ✅ 日志文件重定向
- ✅ 优雅停止（`stop` 命令）
- ✅ 信号处理（SIGTERM、SIGINT）
- ✅ 可配置的 PID 和日志文件路径

## 安装

```bash
go get github.com/muleiwu/gomander
```

## 使用方法

### 基础用法

```go
package main

import (
    "fmt"
    "time"
    "github.com/muleiwu/gomander"
)

func main() {
    gomander.Run(func() {
        // 你的业务逻辑
        fmt.Println("应用程序启动...")
        
        for {
            time.Sleep(5 * time.Second)
            fmt.Println("运行中...")
        }
    })
}
```

### 自定义配置

```go
func main() {
    gomander.Run(func() {
        // 业务逻辑
    }, 
        gomander.WithPidFile("./myapp.pid"),
        gomander.WithLogFile("./myapp.log"),
    )
}
```

## 命令行使用

编译你的程序：

```bash
go build -o myapp
```

### 前台运行

```bash
./myapp
```

日志会直接输出到终端（Stdout/Stderr）。

### 后台守护进程运行

```bash
./myapp -d
```

- 进程会在后台运行
- 日志会重定向到日志文件（默认 `./gomander.log`）
- PID 会保存到文件（默认 `./gomander.pid`）

### 停止守护进程

```bash
./myapp stop
```

会读取 PID 文件并发送 SIGTERM 信号停止进程。

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithPidFile(path)` | PID 文件路径 | `./gomander.pid` |
| `WithLogFile(path)` | 日志文件路径 | `./gomander.log` |

## 工作原理

### 前台模式

```
用户程序 → gomander.Run() → 直接执行用户函数 → 日志输出到屏幕
```

### Daemon 模式

```
用户程序 -d → gomander.Run() → Fork 自身 → 父进程退出
                                    ↓
                              子进程（守护进程）
                                    ↓
                              写入 PID 文件
                                    ↓
                              执行用户函数
                                    ↓
                              日志写入文件
```

### Stop 命令

```
用户程序 stop → 读取 PID 文件 → 发送 SIGTERM → 删除 PID 文件
```

## 示例

查看 `example/main.go` 获取完整示例。

运行示例：

```bash
cd example
go build -o myapp
./myapp           # 前台运行
./myapp -d        # 后台运行
./myapp stop      # 停止
```

查看日志：

```bash
tail -f myapp.log
```

## 环境变量

- `GOMANDER_DAEMON=1`: 内部使用，标识当前进程是守护进程子进程

## 注意事项

1. 确保有权限在指定路径创建 PID 和日志文件
2. 停止进程前确保 PID 文件存在
3. 信号处理会自动清理 PID 文件

## License

MIT
