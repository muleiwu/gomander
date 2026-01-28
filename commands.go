package gomander

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

var daemonMode bool

// createRootCommand 创建根命令
func createRootCommand(config *Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Process manager with daemon support",
		Long:  "A process manager that supports foreground and daemon modes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRootCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// 添加 -d, --daemon flag
	rootCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in daemon mode")

	return rootCmd
}

// runRootCommand 执行根命令逻辑
func runRootCommand(config *Config) error {
	// 如果是 daemon 模式且不是守护进程子进程，则 fork
	if daemonMode && !isDaemonChild() {
		// Fork 自身为守护进程
		if err := forkDaemon(config); err != nil {
			return fmt.Errorf("failed to fork daemon: %w", err)
		}
		// 父进程退出
		return nil
	}

	// 如果是守护进程子进程，写入 PID 文件
	if isDaemonChild() {
		if err := writePidFile(config); err != nil {
			return fmt.Errorf("failed to write PID file: %w", err)
		}

		// 设置信号处理器
		setupSignalHandler(config, nil)

		fmt.Printf("Daemon process started with PID: %d\n", os.Getpid())
		fmt.Printf("PID file: %s\n", config.PidFile)
	}

	// 执行用户函数
	if config.userFunc != nil {
		config.userFunc()
	}

	// 如果是守护进程，清理 PID 文件
	if isDaemonChild() {
		removePidFile(config)
	}

	return nil
}

// createStopCommand 创建 stop 子命令
func createStopCommand(config *Config) *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the daemon process",
		Long:  "Stop the daemon process by sending SIGTERM signal",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStopCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return stopCmd
}

// runStopCommand 执行 stop 命令逻辑
func runStopCommand(config *Config) error {
	// 读取 PID 文件
	pid, err := readPidFile(config)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// 发送 SIGTERM 信号
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	fmt.Printf("Sent SIGTERM signal to process %d\n", pid)

	// 删除 PID 文件
	if err := removePidFile(config); err != nil {
		fmt.Printf("Warning: failed to remove PID file: %v\n", err)
	}

	return nil
}
