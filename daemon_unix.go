//go:build !windows

package gomander

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// setDaemonSysProcAttr 在 Unix 系统上创建新会话，脱离控制终端
func setDaemonSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

// notifyShutdownSignals 订阅 Unix 下的退出与重载信号
func notifyShutdownSignals(sigChan chan<- os.Signal) {
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
}

// isReloadSignal 判断是否为重载信号（Unix 下为 SIGHUP）
func isReloadSignal(sig os.Signal) bool {
	return sig == syscall.SIGHUP
}

// processRunning 通过发送 signal 0 检测进程是否存活
func processRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
