//go:build windows

package app

import (
	"fmt"
	"os/exec"
	"strings"
)

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	// os.FindProcess always returns a valid proc on Windows,
	// so use tasklist to verify
	out, _ := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	return strings.Contains(string(out), fmt.Sprintf("%d", pid))
}
