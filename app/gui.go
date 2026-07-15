package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed web/index.html
var webHTML []byte

func (a *App) CmdGUI() {
	fmt.Println("  正在启动界面...")
	os.MkdirAll(a.YuanshuDir, 0755)

	go a.startProxy()
	time.Sleep(500 * time.Millisecond)

	if url, ok := a.runningGUIURL(); ok {
		openBrowser(url)
		fmt.Printf("  已有实例正在运行: %s\n", url)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(webHTML)
	})
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/login", a.handleAPILogin)
	mux.HandleFunc("/api/logout", a.handleAPILogout)
	mux.HandleFunc("/proxy.pac", handlePAC)

	port := "18900"
	addr := "127.0.0.1:" + port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("  启动界面失败: %v\n", err)
			return
		}
		addr = listener.Addr().String()
	}

	url := "http://" + addr
	a.writeGUIState(url)
	defer a.cleanupGUIState()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	go func() {
		<-stop
		listener.Close()
	}()

	openBrowser(url)
	fmt.Printf("  浏览器已打开: %s\n", url)
	fmt.Println("  关闭窗口后按 Ctrl+C 退出")
	if err := http.Serve(listener, mux); err != nil {
		if !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Printf("  界面服务已停止: %v\n", err)
		}
	}
}

func handlePAC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Write([]byte(`function FindProxyForURL(url, host) {
    if (host == "ab.chatgpt.com" || host == "ab.chatgpt.com:443") {
        return "PROXY 127.0.0.1:9090";
    }
    return "DIRECT";
}`))
}

func (a *App) setPAC() {
	pacURL := "http://127.0.0.1:18900/proxy.pac"
	if runtime.GOOS == "darwin" {
		out, _ := exec.Command("networksetup", "-listallnetworkservices").Output()
		for _, svc := range strings.Fields(string(out)) {
			if svc == "An asterisk (*) denotes that a network service is disabled." {
				continue
			}
			exec.Command("networksetup", "-setautoproxyurl", svc, pacURL).Run()
			exec.Command("networksetup", "-setautoproxystate", svc, "on").Run()
		}
		log.Printf("[pac] set for all network services (macOS)")
	} else if runtime.GOOS == "windows" {
		exec.Command("reg", "add",
			"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
			"/v", "AutoConfigURL",
			"/t", "REG_SZ",
			"/d", pacURL,
			"/f").Run()
		log.Printf("[pac] set via registry (Windows)")
	}
}

func (a *App) unsetPAC() {
	if runtime.GOOS == "darwin" {
		out, _ := exec.Command("networksetup", "-listallnetworkservices").Output()
		for _, svc := range strings.Fields(string(out)) {
			if svc == "An asterisk (*) denotes that a network service is disabled." {
				continue
			}
			exec.Command("networksetup", "-setautoproxystate", svc, "off").Run()
		}
		log.Printf("[pac] cleared for all network services (macOS)")
	} else if runtime.GOOS == "windows" {
		exec.Command("reg", "delete",
			"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
			"/v", "AutoConfigURL",
			"/f").Run()
		log.Printf("[pac] cleared via registry (Windows)")
	}
}

func (a *App) runningGUIURL() (string, bool) {
	pidPath := filepath.Join(a.YuanshuDir, "gui.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return "", false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil || !processRunning(pid) {
		a.cleanupGUIState()
		return "", false
	}

	urlPath := filepath.Join(a.YuanshuDir, "gui.url")
	if data, err := os.ReadFile(urlPath); err == nil {
		if url := strings.TrimSpace(string(data)); url != "" {
			return url, true
		}
	}
	return "http://127.0.0.1:18900", true
}

func (a *App) writeGUIState(url string) {
	os.WriteFile(filepath.Join(a.YuanshuDir, "gui.pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	os.WriteFile(filepath.Join(a.YuanshuDir, "gui.url"), []byte(url), 0644)
}

func (a *App) cleanupGUIState() {
	os.Remove(filepath.Join(a.YuanshuDir, "gui.pid"))
	os.Remove(filepath.Join(a.YuanshuDir, "gui.url"))
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return proc != nil
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func openBrowser(url string) {
	go func() {
		if runtime.GOOS == "darwin" {
			exec.Command("open", url).Start()
		} else if runtime.GOOS == "windows" {
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		}
	}()
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	configPath := filepath.Join(a.CodexHome, "config.toml")
	loggedIn := false
	if data, err := os.ReadFile(configPath); err == nil {
		loggedIn = strings.Contains(string(data), "custom-proxy")
	}

	modelsPath := filepath.Join(a.YuanshuDir, "metaproxy-models.json")
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

	log.Printf("[gui] status: logged_in=%v models=%d", loggedIn, len(modelNames))
	json.NewEncoder(w).Encode(map[string]any{
		"logged_in": loggedIn,
		"models":    modelNames,
	})
}

func (a *App) handleAPILogin(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req struct {
		APIKey string `json:"api_key"`
	}
	json.Unmarshal(body, &req)
	log.Printf("[gui] POST /api/login")

	if req.APIKey == "" {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "API Key 不能为空"})
		return
	}

	models, err := a.fetchModels(req.APIKey)
	if err != nil {
		log.Printf("[gui] fetchModels failed: %v", err)
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": fmt.Sprintf("登录失败: %v", err)})
		return
	}
	if len(models) == 0 {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "未获取到模型列表"})
		return
	}

	os.MkdirAll(a.YuanshuDir, 0755)
	configPath := filepath.Join(a.CodexHome, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil && !strings.Contains(string(data), "custom-proxy") {
		os.WriteFile(filepath.Join(a.YuanshuDir, "backup.config.toml"), data, 0644)
	}
	a.writeModelCatalog(models)
	a.writeConfig(models[0], req.APIKey)

	a.InstallCert()
	a.setPAC()

	log.Printf("[gui] login OK — %d models", len(models))
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"message": fmt.Sprintf("登录成功！共 %d 个模型可用", len(models)),
		"models":  models,
	})
}

func (a *App) handleAPILogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("[gui] POST /api/logout")
	a.runLogoutCleanup()
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "message": "已退出并清理"})
}
