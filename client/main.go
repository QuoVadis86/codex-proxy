package main

import "os"

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "login":
			cmdLogin()
			return
		case "logout":
			cmdLogout()
			return
		case "server":
			cmdServer()
			return
		}
	}
	// 默认启动 GUI 界面
	cmdGUI()
}
