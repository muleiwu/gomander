package gomander

import (
	"os"
)

// Config 存储 gomander 的配置
type Config struct {
	PidFile  string
	LogFile  string
	userFunc func()
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

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		PidFile: "./gomander.pid",
		LogFile: "./gomander.log",
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
	rootCmd := createRootCommand(config)
	rootCmd.AddCommand(
		createStartCommand(config),
		createStopCommand(config),
		createRestartCommand(config),
		createReloadCommand(config),
		createStatusCommand(config),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
