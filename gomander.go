package gomander

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config 存储 gomander 的配置
type Config struct {
	PidFile    string
	LogFile    string
	userFunc   func()
	commands   []*cobra.Command
	daemonMode bool
}

// Option 是配置选项函数类型
type Option func(*Config)

// WithPidFile 设置 PID 文件路径
func WithPidFile(path string) Option {
	return func(c *Config) {
		c.PidFile = path
	}
}

// WithLogFile 设置日志文件路径
func WithLogFile(path string) Option {
	return func(c *Config) {
		c.LogFile = path
	}
}

// WithCommands 注册额外的顶层 Cobra 子命令。
// 自定义命令独立执行，不参与 gomander 的 daemon 生命周期管理。
func WithCommands(commands ...*cobra.Command) Option {
	return func(c *Config) {
		c.commands = append(c.commands, commands...)
	}
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		PidFile: "./runtime/gomander.pid",
		LogFile: "./runtime/logs/gomander.log",
	}
}

// Run 是 gomander 的主入口函数
// fn: 用户的业务逻辑函数
// opts: 可选的配置选项
func Run(fn func(), opts ...Option) {
	// 创建配置
	config := defaultConfig()
	config.userFunc = fn

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 创建并执行 Cobra 命令
	rootCmd, err := buildRootCommand(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
