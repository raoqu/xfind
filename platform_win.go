//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func initCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
