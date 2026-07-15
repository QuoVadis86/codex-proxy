package app

import (
	"log"
	"os"
	"path/filepath"
)

type App struct {
	CodexHome  string
	YuanshuDir string
	ProxyURL   string
}

func New() *App {
	log.SetFlags(log.Ltime)
	home, _ := os.UserHomeDir()
	return &App{
		CodexHome:  filepath.Join(home, ".codex"),
		YuanshuDir: filepath.Join(home, ".codex", "yuanshu"),
		ProxyURL:   "http://113.90.157.107:8317/v1",
	}
}
