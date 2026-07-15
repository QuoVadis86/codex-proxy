//go:build windows

package app

import (
	"log"
	"os/exec"
)

func (a *App) setPAC() {
	pacURL := "http://127.0.0.1:18900/proxy.pac"
	exec.Command("reg", "add",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
		"/v", "AutoConfigURL",
		"/t", "REG_SZ",
		"/d", pacURL,
		"/f").Run()
	log.Printf("[pac] set")
}

func (a *App) unsetPAC() {
	exec.Command("reg", "delete",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
		"/v", "AutoConfigURL",
		"/f").Run()
	log.Printf("[pac] cleared")
}
