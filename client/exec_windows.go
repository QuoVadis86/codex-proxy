//go:build windows

package main

import "syscall"

func setProcessGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
