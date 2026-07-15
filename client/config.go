package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const proxyURL = "http://113.90.157.107:8317/v1"
const statsigServer = "127.0.0.1"

var codexHome = filepath.Join(os.Getenv("HOME"), ".codex")
var yuanshuDir = filepath.Join(codexHome, "yuanshu")

func cmdLogin() {
	alreadyLoggedIn := false
	configPath := filepath.Join(codexHome, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil && strings.Contains(string(data), "custom-proxy") {
		alreadyLoggedIn = true
	}

	if !alreadyLoggedIn {
		// 输入 API Key
		fmt.Print("  请输入你的 API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		if apiKey == "" {
			fmt.Println("  ❌ API Key 不能为空")
			return
		}

		// 连接服务器
		fmt.Println("\n  → 正在连接服务器...")
		models, err := fetchModels(apiKey)
		if err != nil {
			fmt.Printf("  ❌ 登录失败: %v\n", err)
			return
		}
		if len(models) == 0 {
			fmt.Println("  ❌ 未获取到模型列表，登录失败")
			return
		}
		fmt.Printf("  ✅ 连接成功！共 %d 个模型\n", len(models))

		// 备份
		os.MkdirAll(yuanshuDir, 0755)
		if data, err := os.ReadFile(configPath); err == nil && !strings.Contains(string(data), "custom-proxy") {
			os.WriteFile(filepath.Join(yuanshuDir, "backup.config.toml"), data, 0644)
		}

		// 生成模型配置
		writeModelCatalog(models)
		writeConfig(models[0], apiKey)

		fmt.Println("\n  可用模型:")
		for _, m := range models {
			fmt.Printf("    • %s\n", m)
		}
	} else {
		// 已登录，但刷新模型配置（修复旧版本缺少的字段）
		fmt.Println("  → 刷新模型配置...")
		os.MkdirAll(yuanshuDir, 0755)
		if data, err := os.ReadFile(configPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "experimental_bearer_token = ") {
					apiKey := strings.Trim(strings.SplitN(line, "\"", 3)[1], "\"")
					if models, err := fetchModels(apiKey); err == nil && len(models) > 0 {
						writeModelCatalog(models)
						fmt.Printf("  ✅ 已刷新，共 %d 个模型\n", len(models))
					}
					break
				}
			}
		}
	}

	// 加 hosts + 启动本地劫持服务
	addHosts()
	startHijackServer()

	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           🎉 登录成功                     ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")
}

func startHijackServer() {
	fmt.Println("  → 启动 Statsig 劫持服务...")

	// 启动子进程运行 server
	exe, _ := os.Executable()
	cmd := exec.Command(exe, "server", "3000")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = setProcessGroupAttr()

	if err := cmd.Start(); err != nil {
		fmt.Println("  ⚠️  启动失败:", err)
		return
	}

	// 记下 PID
	pidPath := filepath.Join(yuanshuDir, "proxy.pid")
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
	fmt.Println("  ✅ Statsig 劫持服务已启动")
}

func cmdLogout() {
	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║       元数智慧 AI Proxy · 退出            ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")

	// 停掉劫持服务
	pidPath := filepath.Join(yuanshuDir, "proxy.pid")
	if data, err := os.ReadFile(pidPath); err == nil {
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)
		proc, err := os.FindProcess(pid)
		if err == nil {
			proc.Kill()
			fmt.Println("  ✅ 劫持服务已停止")
		}
		os.Remove(pidPath)
	}

	// 恢复配置
	backupPath := filepath.Join(yuanshuDir, "backup.config.toml")
	configPath := filepath.Join(codexHome, "config.toml")
	if data, err := os.ReadFile(backupPath); err == nil {
		os.WriteFile(configPath, data, 0644)
	} else {
		os.Remove(configPath)
	}

	// 删 hosts
	removeHosts()

	// 清理临时文件
	os.Remove(filepath.Join(yuanshuDir, "metaproxy-models.json"))
	os.Remove(filepath.Join(yuanshuDir, "custom-proxy.config.toml"))
	os.Remove(filepath.Join(yuanshuDir, "custom-proxy-fast.config.toml"))

	fmt.Println("\n  ╔═══════════════════════════════════════════╗")
	fmt.Println("  ║           🎉 已退出                       ║")
	fmt.Println("  ╚═══════════════════════════════════════════╝")
}

func fetchModels(apiKey string) ([]string, error) {
	req, _ := http.NewRequest("GET", proxyURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("API Key 错误，请检查后重试")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	var models []string
	for _, m := range result.Data {
		if m.ID != "wan2.7-image" {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

func writeModelCatalog(models []string) {
	type Level struct {
		Effort      string `json:"effort"`
		Description string `json:"description"`
	}

	truncPolicy := map[string]any{
		"mode": "tokens", "limit": 100000,
	}
	modalities := []string{"text", "image"}

	template := map[string]any{
		"shell_type": "shell_command", "visibility": "list",
		"supported_in_api": true,
		"additional_speed_tiers": []any{},
		"service_tiers":         []any{},
		"availability_nux":      nil,
		"upgrade":               nil,
		"base_instructions":      "You are Codex, a coding agent.",
		"model_messages":         map[string]any{},
		"include_skills_usage_instructions": true,
		"supports_reasoning_summaries":      true,
		"default_reasoning_summary":         "none",
		"support_verbosity": true,
		"default_verbosity":  "low",
		"apply_patch_tool_type":    "freeform",
		"web_search_tool_type":     "text_and_image",
		"truncation_policy":        truncPolicy,
		"supports_parallel_tool_calls":       true,
		"supports_image_detail_original":     true,
		"context_window":            1000000,
		"max_context_window":        1000000,
		"effective_context_window_percent": 95,
		"experimental_supported_tools": []any{},
		"input_modalities":   modalities,
		"supports_search_tool": true,
		"use_responses_lite":   false,
		"default_reasoning_level": "medium",
	}

	known := map[string]string{
		"deepseek": "DeepSeek", "gpt": "GPT", "qwen": "Qwen",
		"codex": "Codex", "glm": "GLM", "claude": "Claude", "gemini": "Gemini",
	}

	slugDisplay := func(slug string) string {
		parts := strings.Split(strings.ReplaceAll(slug, "-", " "), " ")
		res := make([]string, len(parts))
		for i, p := range parts {
			l := strings.ToLower(p)
			if v, ok := known[l]; ok {
				res[i] = v
			} else {
				res[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		return strings.Join(res, " ")
	}

	levels := func(slug string) []Level {
		s := strings.ToLower(slug)
		if strings.HasPrefix(s, "gpt") {
			return []Level{
				{"low", "Fast"}, {"medium", "Balanced"}, {"high", "Deep"},
				{"xhigh", "Extra deep"}, {"max", "Max"}, {"ultra", "Ultra"},
			}
		}
		if strings.HasPrefix(s, "deepseek") {
			return []Level{
				{"low", "Fast"}, {"medium", "Balanced"}, {"high", "Default"},
				{"xhigh", "Extra deep"}, {"max", "Max"},
			}
		}
		return []Level{{"low", "Fast"}, {"medium", "Balanced"}, {"high", "Deep"}}
	}

	var entries []map[string]any
	for i, slug := range models {
		e := make(map[string]any)
		for k, v := range template {
			e[k] = v
		}
		e["slug"] = slug
		e["display_name"] = slugDisplay(slug)
		e["description"] = slugDisplay(slug) + " via proxy"
		e["supported_reasoning_levels"] = levels(slug)
		e["priority"] = i + 1

		if strings.Contains(strings.ToLower(slug), "qwen") {
			e["context_window"] = 131072
			e["max_context_window"] = 131072
		}
		entries = append(entries, e)
	}

	data, _ := json.MarshalIndent(map[string]any{"models": entries}, "", "  ")
	os.WriteFile(filepath.Join(yuanshuDir, "metaproxy-models.json"), data, 0644)
}

func writeConfig(firstModel, apiKey string) {
	config := fmt.Sprintf(`model = "%s"
model_provider = "custom-proxy"
model_reasoning_effort = "medium"
model_catalog_json = "%s"

[model_providers.custom-proxy]
name = "元数智慧 · Codex AI Proxy"
base_url = "%s"
experimental_bearer_token = "%s"
requires_openai_auth = false
`, firstModel, filepath.Join(yuanshuDir, "metaproxy-models.json"), proxyURL, apiKey)
	os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(config), 0644)
}
