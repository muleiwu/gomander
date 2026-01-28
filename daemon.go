package gomander

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const daemonEnvKey = "GOMANDER_DAEMON"

// isDaemonChild 检查当前进程是否是守护进程子进程
func isDaemonChild() bool {
	return os.Getenv(daemonEnvKey) == "1"
}

// forkDaemon 将当前进程 fork 为守护进程
func forkDaemon(config *Config) error {
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
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %v\n", sig)

		// 执行清理函数
		if cleanup != nil {
			cleanup()
		}

		// 删除 PID 文件
		removePidFile(config)

		os.Exit(0)
	}()
}
