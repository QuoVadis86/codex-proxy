package app

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func New() *App {
	log.SetFlags(log.Ltime)
	loadEnvFile(".env")

	home, _ := os.UserHomeDir()
	a := &App{
		Plat:       newPlatform(),
		CodexHome:  filepath.Join(home, ".codex"),
		YuanshuDir: filepath.Join(home, ".codex", "yuanshu"),
		ProxyURL:   env("YUANSHU_PROXY_URL", defaultProxyURL),
	}
	return a
}
