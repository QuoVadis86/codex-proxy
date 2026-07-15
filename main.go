package main

import (
	"os"

	"codex-proxy/app"
)

func main() {
	a := app.New()

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "login":
			a.CmdLogin()
			return
		case "logout":
			a.CmdLogout()
			return
		case "server":
			a.CmdServer()
			return
		case "uninstall":
			a.RemoveCert()
			return
		}
	}
	a.CmdGUI()
}
