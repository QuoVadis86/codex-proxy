package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed ca.crt
var caCert []byte

func installCert() {
	certPath := filepath.Join(yuanshuDir, "ca.crt")
	os.MkdirAll(filepath.Dir(certPath), 0755)
	os.WriteFile(certPath, caCert, 0644)

	fmt.Println("  → 安装 SSL 证书...")

	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(
			`do shell script "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '%s'" with administrator privileges`,
			certPath,
		)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			fmt.Println("  ⚠️  证书安装失败")
			return
		}
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("certutil", "-addstore", "Root", certPath)
		if err := cmd.Run(); err != nil {
			fmt.Println("  ⚠️  证书安装失败")
			return
		}
	}
	fmt.Println("  ✅ 证书已安装")
}

func removeCert() {
	if runtime.GOOS == "darwin" {
		script := `do shell script "security delete-certificate -c 'YuanshuStatsigCA'" with administrator privileges`
		cmd := exec.Command("osascript", "-e", script)
		cmd.Run()
	} else if runtime.GOOS == "windows" {
		exec.Command("certutil", "-delstore", "Root", "YuanshuStatsigCA").Run()
	}
	os.Remove(filepath.Join(yuanshuDir, "ca.crt"))
}
