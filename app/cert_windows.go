//go:build windows

package app

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *App) IsCertInstalled() bool {
	out, _ := exec.Command("certutil", "-store", "Root", "YuanshuCA").Output()
	return strings.Contains(string(out), "YuanshuCA")
}

func (a *App) InstallCert() bool {
	if a.IsCertInstalled() {
		log.Printf("[ca] already installed, skipping")
		return true
	}

	a.ensureCA()
	tmpFile := filepath.Join(os.TempDir(), "yuanshu-ca.crt")
	os.WriteFile(tmpFile, a.CACertPEM(), 0644)
	defer os.Remove(tmpFile)

	log.Printf("[ca] installing...")
	cmd := exec.Command("certutil", "-addstore", "Root", tmpFile)
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
	exec.Command("certutil", "-delstore", "Root", "YuanshuCA").Run()
}
