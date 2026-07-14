package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var fallbackResp = []byte(`{"feature_gates":{},"dynamic_configs":{},"layer_configs":{},"has_updates":true}`)

func proxyToReal(body []byte, path string) ([]byte, error) {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", "104.18.32.47:443",
		&tls.Config{ServerName: "ab.chatgpt.com", InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	req := fmt.Sprintf("POST %s HTTP/1.1\r\nHost: ab.chatgpt.com\r\nContent-Type: application/json\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", path, len(body), body)
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	if _, err := conn.Write([]byte(req)); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	respRaw, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	// 找 HTTP body
	parts := strings.SplitN(string(respRaw), "\r\n\r\n", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("no http body")
	}
	headerLine := strings.SplitN(parts[0], "\r\n", 2)[0]
	if !strings.Contains(headerLine, "200") {
		return nil, fmt.Errorf("status not 200: %s", headerLine)
	}

	return []byte(parts[1]), nil
}

func stripModels(data []byte) []byte {
	var parsed map[string]any
	if json.Unmarshal(data, &parsed) != nil {
		return data
	}
	for _, key := range []string{"dynamic_configs", "layer_configs"} {
		m, _ := parsed[key].(map[string]any)
		for _, c := range m {
			x, _ := c.(map[string]any)["value"].(map[string]any)
			delete(x, "available_models")
			delete(x, "allowed_models")
			delete(x, "use_hidden_models")
		}
	}
	cleaned, _ := json.Marshal(parsed)
	return cleaned
}

func main() {
	ca, _ := os.ReadFile("ca.crt")
	mux := http.NewServeMux()

	mux.HandleFunc("/ca.crt", func(w http.ResponseWriter, r *http.Request) {
		w.Write(ca)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/v1/initialize", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		if strings.Contains(r.URL.RequestURI(), "k=client-") {
			realData, err := proxyToReal(body, r.URL.RequestURI())
			if err == nil {
				cleaned := stripModels(realData)
				log.Printf("  /v1/initialize → proxy OK")
				w.Header().Set("Content-Type", "application/json")
				w.Write(cleaned)
				return
			}
			log.Printf("  ⚠️  /v1/initialize proxy failed: %v, using fallback", err)
		} else {
			log.Printf("  ⚠️  /v1/initialize no API key in URL, using fallback")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fallbackResp)
	})

	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		realData, err := proxyToReal(body, r.URL.RequestURI())
		if err != nil {
			log.Printf("  /v1/* proxy fail: %v", err)
			w.WriteHeader(502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(realData)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	log.Println("  :443  (raw TCP proxy → real Statsig)")
	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", mux))
}
