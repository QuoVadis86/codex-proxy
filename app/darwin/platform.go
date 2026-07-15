//go:build darwin

package darwin

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Darwin struct{}

func (*Darwin) MachineID() string {
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

func (*Darwin) IsCertInstalled(certPath string) bool {
	localPEM, err := exec.Command("cat", certPath).Output()
	if err != nil {
		return false
	}
	block, _ := pem.Decode(localPEM)
	if block == nil {
		return false
	}
	localHash := sha256.Sum256(block.Bytes)

	out, _ := exec.Command("security", "find-certificate", "-c", "YuanshuCA", "-p").Output()
	installedBlock, _ := pem.Decode(out)
	if installedBlock == nil {
		return false
	}
	installedHash := sha256.Sum256(installedBlock.Bytes)
	return localHash == installedHash
}

func (d *Darwin) InstallCert(certPath string) bool {
	if d.IsCertInstalled(certPath) {
		log.Printf("[ca] already installed, skipping")
		return true
	}
	log.Printf("[ca] installing...")
	script := fmt.Sprintf(
		`do shell script "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '%s'" with administrator privileges`,
		certPath,
	)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("[ca] install failed: %v", err)
		return false
	}
	log.Printf("[ca] installed")
	return true
}

func (*Darwin) RemoveCert() {
	script := `do shell script "security delete-certificate -c 'YuanshuCA'" with administrator privileges`
	exec.Command("osascript", "-e", script).Run()
}

func (*Darwin) SetPAC(pacURL string) {
	out, _ := exec.Command("networksetup", "-listallnetworkservices").Output()
	for _, svc := range strings.Fields(string(out)) {
		if svc == "An asterisk (*) denotes that a network service is disabled." {
			continue
		}
		exec.Command("networksetup", "-setautoproxyurl", svc, pacURL).Run()
		exec.Command("networksetup", "-setautoproxystate", svc, "on").Run()
	}
	log.Printf("[pac] set")
}

func (*Darwin) UnsetPAC() {
	out, _ := exec.Command("networksetup", "-listallnetworkservices").Output()
	for _, svc := range strings.Fields(string(out)) {
		if svc == "An asterisk (*) denotes that a network service is disabled." {
			continue
		}
		exec.Command("networksetup", "-setautoproxystate", svc, "off").Run()
	}
	log.Printf("[pac] cleared")
}

func (*Darwin) OpenBrowser(url string) {
	exec.Command("open", url).Start()
}

func (*Darwin) ProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid)).Output()
	return err == nil && strings.Contains(string(proc), fmt.Sprintf("%d", pid))
}
