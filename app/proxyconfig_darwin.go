//go:build darwin

package app

import (
	"log"
	"os/exec"
	"strings"
)

func (a *App) setPAC() {
	pacURL := "http://127.0.0.1:18900/proxy.pac"
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

func (a *App) unsetPAC() {
	out, _ := exec.Command("networksetup", "-listallnetworkservices").Output()
	for _, svc := range strings.Fields(string(out)) {
		if svc == "An asterisk (*) denotes that a network service is disabled." {
			continue
		}
		exec.Command("networksetup", "-setautoproxystate", svc, "off").Run()
	}
	log.Printf("[pac] cleared")
}
