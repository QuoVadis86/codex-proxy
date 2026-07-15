package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) CmdLogin() {
	configPath := filepath.Join(a.CodexHome, "config.toml")

	alreadyLoggedIn := false
	if data, err := os.ReadFile(configPath); err == nil && strings.Contains(string(data), "custom-proxy") {
		alreadyLoggedIn = true
		log.Printf("[login] detected existing config, refresh mode")
	}

	if !alreadyLoggedIn {
		fmt.Print("  请输入你的 API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		if apiKey == "" {
			fmt.Println("  ❌ API Key 不能为空")
			return
		}

		log.Printf("[login] connecting to proxy server...")
		models, err := a.fetchModels(apiKey)
		if err != nil {
			fmt.Printf("  ❌ 登录失败: %v\n", err)
			return
		}
		if len(models) == 0 {
			fmt.Println("  ❌ 未获取到模型列表")
			return
		}
		fmt.Printf("  ✅ 连接成功！共 %d 个模型\n", len(models))

		os.MkdirAll(a.YuanshuDir, 0755)
		if data, err := os.ReadFile(configPath); err == nil && !strings.Contains(string(data), "custom-proxy") {
			os.WriteFile(filepath.Join(a.YuanshuDir, "backup.config.toml"), data, 0644)
		}

		a.writeModelCatalog(models)
		a.writeConfig(models[0], apiKey)

		fmt.Println("\n  可用模型:")
		for _, m := range models {
			fmt.Printf("    • %s\n", m)
		}
	} else {
		log.Printf("[login] refresh mode")
		os.MkdirAll(a.YuanshuDir, 0755)
		if data, err := os.ReadFile(configPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "experimental_bearer_token = ") {
					apiKey := strings.Trim(strings.SplitN(line, "\"", 3)[1], "\"")
					if models, err := a.fetchModels(apiKey); err == nil && len(models) > 0 {
						a.writeModelCatalog(models)
						fmt.Printf("  ✅ 已刷新，共 %d 个模型\n", len(models))
					}
					break
				}
			}
		}
	}

	log.Printf("[login] installing CA...")
	a.InstallCert()

	log.Printf("[login] setting PAC...")
	a.setPAC()

	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           🎉 登录成功                     ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")
}

func (a *App) CmdLogout() {
	log.Printf("[logout] starting logout")
	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           退出                           ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")

	a.runLogoutCleanup()

	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           🎉 已退出                       ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")
}

func (a *App) runLogoutCleanup() {
	a.unsetPAC()

	backupPath := filepath.Join(a.YuanshuDir, "backup.config.toml")
	configPath := filepath.Join(a.CodexHome, "config.toml")
	if data, err := os.ReadFile(backupPath); err == nil {
		os.WriteFile(configPath, data, 0644)
		log.Printf("[logout] config restored")
	} else {
		os.Remove(configPath)
		log.Printf("[logout] config removed")
	}

	os.Remove(filepath.Join(a.YuanshuDir, "metaproxy-models.json"))
	log.Printf("[logout] done")
}
