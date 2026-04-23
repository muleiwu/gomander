//go:build windows

package gomander

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Windows 平台常量（stdlib syscall 未导出）
const (
	detachedProcess         = 0x00000008
	processQueryLimitedInfo = 0x1000
	stillActive             = 259
)

// setDaemonSysProcAttr 在 Windows 下让守护进程脱离父进程的控制台
func setDaemonSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: detachedProcess | syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// notifyShutdownSignals 订阅 Windows 下可用的退出信号（无 SIGHUP）
func notifyShutdownSignals(sigChan chan<- os.Signal) {
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
}

// isReloadSignal Windows 没有 SIGHUP，始终返回 false
func isReloadSignal(sig os.Signal) bool {
	return false
}

// processRunning 通过打开进程句柄并读取退出状态检测进程是否存活
func processRunning(pid int) bool {
	handle, err := syscall.OpenProcess(processQueryLimitedInfo, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	if err := syscall.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}
	return exitCode == stillActive
}
