//go:build windows

package windows

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Windows struct{}

func (*Windows) MachineID() string {
	out, _ := exec.Command("wmic", "csproduct", "get", "uuid").Output()
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown"
}

func (*Windows) IsCertInstalled(certPath string) bool {
	out, _ := exec.Command("certutil", "-store", "Root", "YuanshuCA").Output()
	return strings.Contains(string(out), "YuanshuCA")
}

func (*Windows) InstallCert(certPath string) bool {
	tmpFile := certPath
	log.Printf("[ca] installing...")
	cmd := exec.Command("certutil", "-addstore", "Root", tmpFile)
	if err := cmd.Run(); err != nil {
		log.Printf("[ca] install failed: %v", err)
		return false
	}
	log.Printf("[ca] installed")
	return true
}

func (*Windows) RemoveCert() {
	exec.Command("certutil", "-delstore", "Root", "YuanshuCA").Run()
}

func (*Windows) SetPAC(pacURL string) {
	exec.Command("reg", "add",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
		"/v", "AutoConfigURL",
		"/t", "REG_SZ",
		"/d", pacURL,
		"/f").Run()
	log.Printf("[pac] set")
}

func (*Windows) UnsetPAC() {
	exec.Command("reg", "delete",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
		"/v", "AutoConfigURL",
		"/f").Run()
	log.Printf("[pac] cleared")
}

func (*Windows) OpenBrowser(url string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func (*Windows) ProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	out, _ := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	return strings.Contains(string(out), fmt.Sprintf("%d", pid))
}
