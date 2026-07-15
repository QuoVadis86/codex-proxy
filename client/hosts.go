package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func hostsPath() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
	}
	return "/etc/hosts"
}

func addHosts() {
	path := hostsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("  ⚠️  无法读取 hosts 文件")
		return
	}
	if strings.Contains(string(data), "ab.chatgpt.com") {
		return
	}

	entry := statsigServer + " ab.chatgpt.com"

	if runtime.GOOS == "darwin" {
		// macOS 用 osascript 提权
		script := fmt.Sprintf(
			"do shell script \"echo '%s' >> /etc/hosts\" with administrator privileges",
			entry,
		)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			fmt.Println("  ⚠️  添加 hosts 失败")
			return
		}
	} else {
		// Windows 直接写（已提权）
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("  ⚠️  添加 hosts 失败，请以管理员身份运行")
			return
		}
		defer f.Close()
		f.WriteString("\n" + entry)
	}
	fmt.Println("  ✅ hosts 已添加")
}

func removeHosts() {
	path := hostsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if !strings.Contains(string(data), "ab.chatgpt.com") {
		return
	}

	if runtime.GOOS == "darwin" {
		script := `do shell script "sed -i '' '/ab.chatgpt.com/d' /etc/hosts" with administrator privileges`
		cmd := exec.Command("osascript", "-e", script)
		cmd.Run()
	} else {
		lines := strings.Split(string(data), "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, "ab.chatgpt.com") {
				newLines = append(newLines, line)
			}
		}
		os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	}
}
