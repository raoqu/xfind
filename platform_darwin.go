//go:build !windows

package main

import "os/exec"

func initCmdSysProcAttr(cmd *exec.Cmd) {
}
