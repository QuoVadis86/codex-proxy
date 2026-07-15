package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed web/index.html
var webHTML []byte

func cmdGUI() {
	fmt.Println("  正在启动界面...")

	mux := http.NewServeMux()

	// 主页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(webHTML)
	})

	// API
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/login", handleAPILogin)
	mux.HandleFunc("/api/logout", handleAPILogout)

	port := "18900"
	addr := "127.0.0.1:" + port

	// 打开浏览器
	url := "http://" + addr
	go func() {
		if runtime.GOOS == "darwin" {
			exec.Command("open", url).Start()
		} else if runtime.GOOS == "windows" {
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		}
	}()

	fmt.Printf("  浏览器已打开: %s\n", url)
	fmt.Println("  关闭页面后按 Ctrl+C 退出")
	http.ListenAndServe(addr, mux)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	configPath := filepath.Join(codexHome, "config.toml")
	loggedIn := false
	if data, err := os.ReadFile(configPath); err == nil {
		loggedIn = strings.Contains(string(data), "custom-proxy")
	}

	serverRunning := false
	pidPath := filepath.Join(yuanshuDir, "proxy.pid")
	if data, err := os.ReadFile(pidPath); err == nil {
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)
		proc, _ := os.FindProcess(pid)
		serverRunning = proc != nil
	}

	modelsPath := filepath.Join(yuanshuDir, "metaproxy-models.json")

	modelNames := []string{}
	if data, err := os.ReadFile(modelsPath); err == nil {
		var result struct {
			Models []struct {
				Slug string `json:"slug"`
			} `json:"models"`
		}
		if json.Unmarshal(data, &result) == nil {
			for _, m := range result.Models {
				modelNames = append(modelNames, m.Slug)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"logged_in":      loggedIn,
		"server_running": serverRunning,
		"models":         modelNames,
	})
}

func handleAPILogin(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req struct {
		APIKey string `json:"api_key"`
	}
	json.Unmarshal(body, &req)

	var msg string
	ok := false

	if req.APIKey == "" {
		msg = "API Key 不能为空"
	} else if models, err := fetchModels(req.APIKey); err != nil {
		msg = fmt.Sprintf("登录失败: %v", err)
	} else if len(models) == 0 {
		msg = "未获取到模型列表"
	} else {
		os.MkdirAll(yuanshuDir, 0755)
		configPath := filepath.Join(codexHome, "config.toml")
		if data, err := os.ReadFile(configPath); err == nil && !strings.Contains(string(data), "custom-proxy") {
			os.WriteFile(filepath.Join(yuanshuDir, "backup.config.toml"), data, 0644)
		}
		writeModelCatalog(models)
		writeConfig(models[0], req.APIKey)
		addHosts()
		startHijackServer()
		msg = fmt.Sprintf("登录成功！共 %d 个模型可用", len(models))
		ok = true
		json.NewEncoder(w).Encode(map[string]any{"ok": ok, "message": msg, "models": models})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"ok": ok, "message": msg})
}

func handleAPILogout(w http.ResponseWriter, r *http.Request) {
	stopHijackServer()

	backupPath := filepath.Join(yuanshuDir, "backup.config.toml")
	configPath := filepath.Join(codexHome, "config.toml")
	if data, err := os.ReadFile(backupPath); err == nil {
		os.WriteFile(configPath, data, 0644)
	} else {
		os.Remove(configPath)
	}
	removeHosts()
	os.Remove(filepath.Join(yuanshuDir, "metaproxy-models.json"))
	os.Remove(filepath.Join(yuanshuDir, "custom-proxy.config.toml"))
	os.Remove(filepath.Join(yuanshuDir, "custom-proxy-fast.config.toml"))

	json.NewEncoder(w).Encode(map[string]any{"ok": true, "message": "已退出并清理"})
}

func stopHijackServer() {
	pidPath := filepath.Join(yuanshuDir, "proxy.pid")
	if data, err := os.ReadFile(pidPath); err == nil {
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)
		if proc, err := os.FindProcess(pid); err == nil {
			proc.Kill()
		}
		os.Remove(pidPath)
	}
}
