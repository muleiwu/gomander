package gomander

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const daemonEnvKey = "GOMANDER_DAEMON"

// isDaemonChild 检查当前进程是否是守护进程子进程
func isDaemonChild() bool {
	return os.Getenv(daemonEnvKey) == "1"
}

// ensureDir 确保文件所在的目录存在
func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}

// forkDaemon 将当前进程 fork 为守护进程
func forkDaemon(config *Config) error {
	// 确保日志文件所在目录存在
	if err := ensureDir(config.LogFile); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 打开日志文件
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// 准备命令：重新执行当前程序
	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	// 设置环境变量，标识这是守护进程
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", daemonEnvKey))

	// 重定向输出到日志文件
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 创建新会话，脱离终端
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// 启动子进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("Daemon started with PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Log file: %s\n", config.LogFile)

	return nil
}

// writePidFile 写入 PID 文件
func writePidFile(config *Config) error {
	// 确保 PID 文件所在目录存在
	if err := ensureDir(config.PidFile); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	pid := os.Getpid()
	content := []byte(strconv.Itoa(pid))

	if err := os.WriteFile(config.PidFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// removePidFile 删除 PID 文件
func removePidFile(config *Config) error {
	if err := os.Remove(config.PidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}
	return nil
}

// readPidFile 读取 PID 文件
func readPidFile(config *Config) (int, error) {
	content, err := os.ReadFile(config.PidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	return pid, nil
}

// setupSignalHandler 设置信号处理器，用于优雅退出
func setupSignalHandler(config *Config, cleanup func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go func() {
		for {
			sig := <-sigChan
			fmt.Printf("Received signal: %v\n", sig)

			switch sig {
			case syscall.SIGHUP:
				// SIGHUP 用于重新加载，不退出进程
				fmt.Println("Reloading configuration...")
				// 这里可以添加重新加载逻辑
				// 用户可以在自己的代码中监听 SIGHUP 信号

			case syscall.SIGTERM, syscall.SIGINT:
				// SIGTERM 或 SIGINT 用于停止进程
				// 执行清理函数
				if cleanup != nil {
					cleanup()
				}

				// 删除 PID 文件
				removePidFile(config)

				os.Exit(0)
			}
		}
	}()
}

// isProcessRunning 检查给定 PID 的进程是否正在运行
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送信号 0 来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// waitForProcessExit 等待进程退出
func waitForProcessExit(pid int, timeoutSeconds int) bool {
	for i := 0; i < timeoutSeconds*10; i++ {
		if !isProcessRunning(pid) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
