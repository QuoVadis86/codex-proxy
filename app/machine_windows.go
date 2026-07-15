//go:build windows

package app

import (
	"os/exec"
	"strings"
)

func machineID() string {
	out, _ := exec.Command("wmic", "csproduct", "get", "uuid").Output()
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown"
}
