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
		Use:           os.Args[0],
		Short:         "Process manager with daemon support",
		Long:          "A process manager that supports start, stop, restart, reload, and status commands",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return rootCmd
}

// createStartCommand 创建 start 子命令
func createStartCommand(config *Config) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the process",
		Long:  "Start the process in foreground or daemon mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStartCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// 添加 -d, --daemon flag
	startCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in daemon mode")

	return startCmd
}

// runStartCommand 执行 start 命令逻辑
func runStartCommand(config *Config) error {
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

// createRestartCommand 创建 restart 子命令
func createRestartCommand(config *Config) *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the daemon process",
		Long:  "Stop the daemon process and start it again in daemon mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestartCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return restartCmd
}

// runRestartCommand 执行 restart 命令逻辑
func runRestartCommand(config *Config) error {
	// 先尝试停止进程
	pid, err := readPidFile(config)
	if err == nil {
		// 进程存在，停止它
		process, err := os.FindProcess(pid)
		if err == nil {
			if err := process.Signal(syscall.SIGTERM); err == nil {
				fmt.Printf("Stopping process %d...\n", pid)
				// 等待进程退出
				waitForProcessExit(pid, 10)
			}
		}
		removePidFile(config)
	}

	// 以 daemon 模式启动
	fmt.Println("Starting daemon process...")
	daemonMode = true
	return runStartCommand(config)
}

// createReloadCommand 创建 reload 子命令
func createReloadCommand(config *Config) *cobra.Command {
	reloadCmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload the daemon process",
		Long:  "Send SIGHUP signal to the daemon process to trigger reload",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReloadCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return reloadCmd
}

// runReloadCommand 执行 reload 命令逻辑
func runReloadCommand(config *Config) error {
	// 读取 PID 文件
	pid, err := readPidFile(config)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// 检查进程是否存在
	if !isProcessRunning(pid) {
		return fmt.Errorf("process %d is not running", pid)
	}

	// 发送 SIGHUP 信号
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send SIGHUP to process %d: %w", pid, err)
	}

	fmt.Printf("Sent SIGHUP signal to process %d\n", pid)

	return nil
}

// createStatusCommand 创建 status 子命令
func createStatusCommand(config *Config) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show the daemon process status",
		Long:  "Check and display the current status of the daemon process",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusCommand(config)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return statusCmd
}

// runStatusCommand 执行 status 命令逻辑
func runStatusCommand(config *Config) error {
	// 读取 PID 文件
	pid, err := readPidFile(config)
	if err != nil {
		fmt.Println("Status: stopped")
		fmt.Printf("PID file: %s (not found)\n", config.PidFile)
		return nil
	}

	// 检查进程是否存在
	if isProcessRunning(pid) {
		fmt.Println("Status: running")
		fmt.Printf("PID: %d\n", pid)
		fmt.Printf("PID file: %s\n", config.PidFile)
		fmt.Printf("Log file: %s\n", config.LogFile)
	} else {
		fmt.Println("Status: stopped (stale PID file)")
		fmt.Printf("PID file: %s (stale)\n", config.PidFile)
	}

	return nil
}
