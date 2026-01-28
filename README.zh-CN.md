# Gomander

Gomander 是一个基于 [Cobra](https://github.com/spf13/cobra) 的 Go 进程守护化库，让你的程序轻松支持前台运行和后台 daemon 模式，并提供完整的进程生命周期管理。

## 功能特性

- 🚀 **子命令架构** - 基于 Cobra，提供 `start`、`stop`、`restart`、`reload`、`status` 子命令
- 🔄 **守护进程模式** - 支持 `-d` 参数将进程后台运行
- 📁 **PID 文件管理** - 自动创建和清理 PID 文件
- 📝 **日志重定向** - 守护模式下自动将输出重定向到日志文件
- 🛑 **优雅退出** - 支持 SIGTERM、SIGINT 信号优雅停止
- ♻️ **热重载** - 支持 SIGHUP 信号触发配置重载
- ⚙️ **灵活配置** - 使用函数选项模式自定义 PID 和日志文件路径

## 安装

```bash
go get github.com/muleiwu/gomander
```

## 快速开始

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
        // 你的业务逻辑
    }, 
        gomander.WithPidFile("./myapp.pid"),
        gomander.WithLogFile("./myapp.log"),
    )
}
```

## 命令行使用

编译你的程序后，即可使用以下子命令：

```bash
go build -o myapp
```

### start - 启动进程

```bash
# 前台运行（日志输出到终端）
./myapp start

# 后台守护进程运行
./myapp start -d
# 或
./myapp start --daemon
```

守护模式下：
- 进程在后台运行，脱离终端
- 日志重定向到日志文件（默认 `./gomander.log`）
- PID 保存到文件（默认 `./gomander.pid`）

### stop - 停止进程

```bash
./myapp stop
```

读取 PID 文件并发送 SIGTERM 信号，优雅停止守护进程。

### restart - 重启进程

```bash
./myapp restart
```

停止当前运行的进程，然后以守护模式重新启动。

### reload - 重载配置

```bash
./myapp reload
```

向守护进程发送 SIGHUP 信号，可用于触发配置重载（需要在业务代码中实现重载逻辑）。

### status - 查看状态

```bash
./myapp status
```

显示守护进程的当前状态，包括：
- 运行状态（running / stopped）
- 进程 PID
- PID 文件路径
- 日志文件路径

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithPidFile(path)` | PID 文件路径 | `./gomander.pid` |
| `WithLogFile(path)` | 日志文件路径 | `./gomander.log` |

## 工作原理

### 前台模式 (start)

```
myapp start → 直接执行用户函数 → 日志输出到终端
```

### 守护模式 (start -d)

```
myapp start -d → Fork 子进程 → 父进程退出
                      ↓
                子进程（守护进程）
                      ↓
                创建新会话（setsid）
                      ↓
                写入 PID 文件
                      ↓
                重定向输出到日志文件
                      ↓
                执行用户函数
```

### 信号处理

| 信号 | 行为 |
|------|------|
| SIGTERM | 优雅退出，清理 PID 文件 |
| SIGINT | 优雅退出，清理 PID 文件 |
| SIGHUP | 触发重载（不退出进程） |

## 完整示例

查看 [example/main.go](example/main.go) 获取完整示例。

```bash
cd example
go build -o myapp

# 启动守护进程
./myapp start -d

# 查看状态
./myapp status

# 查看日志
tail -f myapp.log

# 重载配置
./myapp reload

# 重启进程
./myapp restart

# 停止进程
./myapp stop
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `GOMANDER_DAEMON=1` | 内部使用，标识当前进程是守护进程子进程 |

## 注意事项

1. 确保有权限在指定路径创建 PID 和日志文件
2. 停止进程前确保 PID 文件存在且进程正在运行
3. 信号处理会自动清理 PID 文件
4. `restart` 命令会等待原进程退出（最多 10 秒）后再启动新进程

## 依赖

- [cobra](https://github.com/spf13/cobra) - 命令行框架

## License

MIT
