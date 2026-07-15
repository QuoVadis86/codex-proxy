package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: codex-proxy <login|logout|server>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "login":
		cmdLogin()
	case "logout":
		cmdLogout()
	case "server":
		cmdServer()
	default:
		fmt.Printf("未知命令: %s\n", os.Args[1])
		fmt.Println("可用命令: login, logout, server")
		os.Exit(1)
	}
}
