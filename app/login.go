package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func (a *App) backupOrigin() {
	os.MkdirAll(a.YuanshuDir, 0755)
	configPath := filepath.Join(a.CodexHome, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil && !strings.Contains(string(data), a.ProxyURL) {
		os.WriteFile(filepath.Join(a.YuanshuDir, "backup.config.toml"), data, 0644)
	}
	authPath := filepath.Join(a.CodexHome, "auth.json")
	if data, err := os.ReadFile(authPath); err == nil && !strings.Contains(string(data), a.ProxyURL) {
		os.WriteFile(filepath.Join(a.YuanshuDir, "backup.auth.json"), data, 0644)
	}
}

func (a *App) restoreOrigin() {
	backupPath := filepath.Join(a.YuanshuDir, "backup.config.toml")
	configPath := filepath.Join(a.CodexHome, "config.toml")
	if data, err := os.ReadFile(backupPath); err == nil {
		os.WriteFile(configPath, data, 0644)
	} else {
		os.Remove(configPath)
	}
	authPath := filepath.Join(a.CodexHome, "auth.json")
	if data, err := os.ReadFile(filepath.Join(a.YuanshuDir, "backup.auth.json")); err == nil {
		os.WriteFile(authPath, data, 0644)
	}
	os.Remove(filepath.Join(a.YuanshuDir, "metaproxy-models.json"))
}

func (a *App) setAuthAPIKey(apiKey string) {
	auth := map[string]any{"OPENAI_API_KEY": apiKey}
	written, _ := json.MarshalIndent(auth, "", "  ")
	os.WriteFile(filepath.Join(a.CodexHome, "auth.json"), written, 0644)
	log.Printf("[login] auth.json updated with API key")
}

func (a *App) CmdLogin() {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()

	configPath := filepath.Join(a.CodexHome, "config.toml")
	alreadyLoggedIn := false
	if data, err := os.ReadFile(configPath); err == nil && strings.Contains(string(data), a.ProxyURL) {
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

		a.backupOrigin()
		a.writeModelCatalog(models)
		a.writeConfig(models[0], apiKey)
		a.setAuthAPIKey(apiKey)

		fmt.Println("\n  可用模型:")
		for _, m := range models {
			fmt.Printf("    • %s\n", m)
		}
	} else {
		log.Printf("[login] refresh mode")
		os.MkdirAll(a.YuanshuDir, 0755)
		if data, err := os.ReadFile(filepath.Join(a.CodexHome, "auth.json")); err == nil {
			var auth struct {
				APIKey string `json:"OPENAI_API_KEY"`
			}
			if json.Unmarshal(data, &auth) == nil && auth.APIKey != "" {
				if models, err := a.fetchModels(auth.APIKey); err == nil && len(models) > 0 {
					a.writeModelCatalog(models)
					fmt.Printf("  ✅ 已刷新，共 %d 个模型\n", len(models))
				}
			}
		}
	}

	log.Printf("[login] installing CA...")
	a.Plat.InstallCert(filepath.Join(a.YuanshuDir, "ca.crt"))
	log.Printf("[login] setting PAC...")
	a.Plat.SetPAC("http://127.0.0.1:18900/proxy.pac")
	log.Printf("[login] starting proxy...")
	go a.startProxy()

	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           🎉 登录成功                     ║")
	fmt.Println("  ║   按 Ctrl+C 退出                         ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Printf("[login] shutting down...")
	a.runLogoutCleanup()
}

func (a *App) CmdLogout() {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()

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
	a.Plat.UnsetPAC()
	a.restoreOrigin()
	log.Printf("[logout] done")
}
