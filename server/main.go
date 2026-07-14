package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

const realStatsigIP = "104.18.32.47"

var (
	cachedResponse *statsigResponse
	cacheMu        sync.RWMutex
	httpClient     = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:         "ab.chatgpt.com",
				InsecureSkipVerify: true,
			},
		},
	}
)

type statsigResponse struct {
	FeatureGates   map[string]json.RawMessage `json:"feature_gates"`
	DynamicConfigs map[string]json.RawMessage `json:"dynamic_configs"`
	LayerConfigs   map[string]json.RawMessage `json:"layer_configs"`
	HasUpdates     bool                       `json:"has_updates"`
	Time           int64                      `json:"time,omitempty"`
	DerivedFields  map[string]string          `json:"derived_fields,omitempty"`
	EvaluatedKeys  map[string]any             `json:"evaluated_keys,omitempty"`
}

func stripModelRestrictions(data map[string]any) {
	for _, key := range []string{"dynamic_configs", "layer_configs"} {
		cfgs, ok := data[key].(map[string]any)
		if !ok {
			continue
		}
		for _, cfg := range cfgs {
			entry, ok := cfg.(map[string]any)
			if !ok {
				continue
			}
			val, ok := entry["value"].(map[string]any)
			if !ok {
				continue
			}
			delete(val, "available_models")
			delete(val, "allowed_models")
			delete(val, "use_hidden_models")
		}
	}
}

func fetchInitialize(body []byte, path string) (*statsigResponse, error) {
	url := "https://" + realStatsigIP + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "ab.chatgpt.com")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	stripModelRestrictions(data)

	out := &statsigResponse{
		HasUpdates: true,
	}
	if fg, ok := data["feature_gates"].(map[string]any); ok {
		out.FeatureGates = make(map[string]json.RawMessage, len(fg))
		for k, v := range fg {
			b, _ := json.Marshal(v)
			out.FeatureGates[k] = b
		}
	}
	if dc, ok := data["dynamic_configs"].(map[string]any); ok {
		out.DynamicConfigs = make(map[string]json.RawMessage, len(dc))
		for k, v := range dc {
			b, _ := json.Marshal(v)
			out.DynamicConfigs[k] = b
		}
	}
	if lc, ok := data["layer_configs"].(map[string]any); ok {
		out.LayerConfigs = make(map[string]json.RawMessage, len(lc))
		for k, v := range lc {
			b, _ := json.Marshal(v)
			out.LayerConfigs[k] = b
		}
	}

	cacheMu.Lock()
	cachedResponse = out
	cacheMu.Unlock()

	return out, nil
}

func jsonReply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	if certFile == "" {
		certFile = "server.crt"
	}
	if keyFile == "" {
		keyFile = "server.key"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, 200, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/v1/initialize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			jsonReply(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		resp, err := fetchInitialize(body, r.URL.RequestURI())
		if err != nil {
			cacheMu.RLock()
			cached := cachedResponse
			cacheMu.RUnlock()
			if cached != nil {
				log.Printf("  /v1/initialize  →  %d gates (cache hit)", len(cached.FeatureGates))
				jsonReply(w, 200, cached)
				return
			}
			log.Printf("  !  proxy failed, no cache: %v", err)
			jsonReply(w, 200, map[string]bool{"success": true})
			return
		}
		log.Printf("  /v1/initialize  →  %d gates (proxy)", len(resp.FeatureGates))
		jsonReply(w, 200, resp)
	})

	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, 200, map[string]bool{"success": true})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, 404, map[string]string{"error": "not_found"})
	})

	addr := ":443"
	if a := os.Getenv("ADDR"); a != "" {
		addr = a
	}

	if _, err := os.Stat(certFile); err == nil {
		srv := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		log.Printf("  https://%s  (/v1/initialize -> real Statsig, stripped)", addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("  http://%s  (no cert)", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
	}
}
