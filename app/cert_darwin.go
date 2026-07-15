//go:build darwin

package app

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *App) IsCertInstalled() bool {
	out, _ := exec.Command("security", "find-certificate", "-c", "YuanshuCA").Output()
	return strings.Contains(string(out), "YuanshuCA")
}

func (a *App) InstallCert() bool {
	if a.IsCertInstalled() {
		log.Printf("[ca] already installed, skipping")
		return true
	}

	a.ensureCA()
	certPath := filepath.Join(a.YuanshuDir, "ca.crt")

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

func (a *App) RemoveCert() {
	if !a.IsCertInstalled() {
		return
	}
	script := `do shell script "security delete-certificate -c 'YuanshuCA'" with administrator privileges`
	exec.Command("osascript", "-e", script).Run()
}
