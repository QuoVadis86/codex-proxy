package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	respData, err := os.ReadFile("statsig_response.json")
	if err != nil {
		log.Fatalf("read response.json: %v", err)
	}

	ca, _ := os.ReadFile("ca.crt")
	mux := http.NewServeMux()

	mux.HandleFunc("/ca.crt", func(w http.ResponseWriter, r *http.Request) {
		w.Write(ca)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/v1/initialize", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(respData)
	})

	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	go func() { log.Println("  :80  (/ca.crt)"); http.ListenAndServe(":80", mux) }()

	log.Println("  :443  (Statsig proxy - static response)")
	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", mux))
}
