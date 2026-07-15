//go:build darwin || linux

package main

import "syscall"

func setProcessGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
