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

type Settings struct {
	ProxyURL string `json:"proxy_url"`
}

func (a *App) settingsPath() string {
	return filepath.Join(a.YuanshuDir, "settings.json")
}

func (a *App) LoadSettings() *Settings {
	s := &Settings{ProxyURL: a.ProxyURL}
	data, err := os.ReadFile(a.settingsPath())
	if err != nil {
		return s
	}
	json.Unmarshal(data, s)
	if s.ProxyURL != "" {
		a.ProxyURL = s.ProxyURL
	}
	return s
}

func (a *App) SaveSettings(s *Settings) {
	a.ProxyURL = s.ProxyURL
	os.MkdirAll(a.YuanshuDir, 0755)
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(a.settingsPath(), data, 0644)
}

func New() *App {
	log.SetFlags(log.Ltime)
	home, _ := os.UserHomeDir()
	a := &App{
		Plat:       newPlatform(),
		CodexHome:  filepath.Join(home, ".codex"),
		YuanshuDir: filepath.Join(home, ".codex", "yuanshu"),
		ProxyURL:   "http://113.90.157.107:8317/v1",
	}
	a.LoadSettings()
	return a
}
