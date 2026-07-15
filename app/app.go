package app

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type App struct {
	Plat       Platform
	CodexHome  string
	YuanshuDir string
	ProxyURL   string
	loginMu    sync.Mutex
}

const defaultProxyURL = "http://113.90.157.107:8317/v1"

func New() *App {
	log.SetFlags(log.Ltime)
	home, _ := os.UserHomeDir()
	a := &App{
		Plat:       newPlatform(),
		CodexHome:  filepath.Join(home, ".codex"),
		YuanshuDir: filepath.Join(home, ".codex", "yuanshu"),
		ProxyURL:   defaultProxyURL,
	}
	if env := os.Getenv("YUANSHU_PROXY_URL"); env != "" {
		a.ProxyURL = env
	} else if data, err := os.ReadFile(filepath.Join(a.YuanshuDir, "settings.json")); err == nil {
		var s struct {
			ProxyURL string `json:"proxy_url"`
		}
		json.Unmarshal(data, &s)
		if s.ProxyURL != "" {
			a.ProxyURL = s.ProxyURL
		}
	}
	return a
}
