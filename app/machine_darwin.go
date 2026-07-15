//go:build darwin

package app

import (
	"os/exec"
	"strings"
)

func machineID() string {
	out, _ := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}
