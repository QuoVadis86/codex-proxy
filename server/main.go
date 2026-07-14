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

const statsigHost = "104.18.32.47"

var (
	cache   *result
	cacheMu sync.RWMutex
	client  = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: "ab.chatgpt.com", InsecureSkipVerify: true},
	}}
)

type result struct {
	Gates map[string]json.RawMessage `json:"feature_gates"`
	Conf  map[string]json.RawMessage `json:"dynamic_configs"`
	Layer map[string]json.RawMessage `json:"layer_configs"`
	Up    bool                       `json:"has_updates"`
}

func strip(v map[string]any) {
	for _, k := range []string{"dynamic_configs", "layer_configs"} {
		m, _ := v[k].(map[string]any)
		for _, c := range m {
			x, _ := c.(map[string]any)["value"].(map[string]any)
			delete(x, "available_models")
			delete(x, "allowed_models")
			delete(x, "use_hidden_models")
		}
	}
}

func proxy(body []byte, path string) (*result, error) {
	r, e := client.Post("https://"+statsigHost+path, "application/json", bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	var raw map[string]any
	json.Unmarshal(b, &raw)
	strip(raw)
	out := &result{Up: true}
	fg, _ := raw["feature_gates"].(map[string]any)
	out.Gates = make(map[string]json.RawMessage, len(fg))
	for k, v := range fg {
		out.Gates[k], _ = json.Marshal(v)
	}
	dc, _ := raw["dynamic_configs"].(map[string]any)
	out.Conf = make(map[string]json.RawMessage, len(dc))
	for k, v := range dc {
		out.Conf[k], _ = json.Marshal(v)
	}
	lc, _ := raw["layer_configs"].(map[string]any)
	out.Layer = make(map[string]json.RawMessage, len(lc))
	for k, v := range lc {
		out.Layer[k], _ = json.Marshal(v)
	}
	cacheMu.Lock()
	cache = out
	cacheMu.Unlock()
	return out, nil
}

func main() {
	mux := http.NewServeMux()

	ca, _ := os.ReadFile("ca.crt")

	mux.HandleFunc("/ca.crt", func(w http.ResponseWriter, r *http.Request) {
		w.Write(ca)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/v1/initialize", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		res, err := proxy(body, r.URL.RequestURI())
		if err != nil {
			cacheMu.RLock()
			c := cache
			cacheMu.RUnlock()
			if c != nil {
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(c)
				return
			}
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]bool{"ok": true})
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	// HTTP (port 80) — for ca.crt download before cert is trusted
	go func() {
		log.Println("  :80  (/ca.crt)")
		http.ListenAndServe(":80", mux)
	}()

	log.Println("  :443  (Statsig proxy)")
	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", mux))
}
