package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) fetchModels(apiKey string) ([]string, error) {
	log.Printf("[config] fetching models from %s", a.ProxyURL+"/models")
	req, _ := http.NewRequest("GET", a.ProxyURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[config] HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[config] response status: %d", resp.StatusCode)
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
		log.Printf("[config] JSON decode failed: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	var models []string
	for _, m := range result.Data {
		if m.ID != "wan2.7-image" {
			models = append(models, m.ID)
		}
	}
	log.Printf("[config] got %d models from proxy", len(models))
	return models, nil
}

func (a *App) writeModelCatalog(models []string) {
	type Level struct {
		Effort      string `json:"effort"`
		Description string `json:"description"`
	}

	truncPolicy := map[string]any{
		"mode": "tokens", "limit": 10000,
	}
	modalities := []string{"text", "image"}

	template := map[string]any{
		"shell_type": "shell_command", "visibility": "list",
		"supported_in_api":                  true,
		"additional_speed_tiers":            []any{},
		"service_tiers":                     []any{},
		"availability_nux":                  nil,
		"upgrade":                           nil,
		"base_instructions":                 "You are Codex, a coding agent.",
		"model_messages":                    map[string]any{},
		"include_skills_usage_instructions": true,
		"supports_reasoning_summaries":      true,
		"default_reasoning_summary":         "none",
		"support_verbosity":                 true,
		"default_verbosity":                 "low",
		"apply_patch_tool_type":             "freeform",
		"web_search_tool_type":              "text_and_image",
		"truncation_policy":                 truncPolicy,
		"supports_parallel_tool_calls":      true,
		"supports_image_detail_original":    true,
		"context_window":                    1000000,
		"max_context_window":                1000000,
		"effective_context_window_percent":  95,
		"experimental_supported_tools":      []any{},
		"input_modalities":                  modalities,
		"supports_search_tool":              true,
		"use_responses_lite":                false,
		"default_reasoning_level":           "medium",
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
	os.WriteFile(filepath.Join(a.YuanshuDir, "metaproxy-models.json"), data, 0644)
}

func (a *App) writeConfig(firstModel, apiKey string) {
	configPath := filepath.Join(a.CodexHome, "config.toml")
	existing, _ := os.ReadFile(configPath)

	ourKeys := fmt.Sprintf(`
model = "%s"
openai_base_url = "%s"
model_catalog_json = "%s"
model_reasoning_effort = "medium"
`, firstModel, a.ProxyURL, filepath.Join(a.YuanshuDir, "metaproxy-models.json"))

	existingStr := string(existing)
	for _, key := range []string{"model", "openai_base_url", "model_catalog_json"} {
		lines := strings.Split(existingStr, "\n")
		var keep []string
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), key+" =") {
				continue
			}
			keep = append(keep, line)
		}
		existingStr = strings.Join(keep, "\n")
	}

	merged := ourKeys + "\n" + existingStr
	os.WriteFile(configPath, []byte(merged), 0644)
	log.Printf("[config] config written")
}
