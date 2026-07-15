//go:build darwin

package app

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func (a *App) IsCertInstalled() bool {
	localPEM, err := os.ReadFile(filepath.Join(a.YuanshuDir, "ca.crt"))
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

func (a *App) InstallCert() bool {
	if a.IsCertInstalled() {
		log.Printf("[ca] already installed, skipping")
		return true
	}
	certPath := filepath.Join(a.YuanshuDir, "ca.crt")
	a.ensureCA()

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
	script := `do shell script "security delete-certificate -c 'YuanshuCA'" with administrator privileges`
	exec.Command("osascript", "-e", script).Run()
}
