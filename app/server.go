package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

const proxyPort = ":9090"

func (a *App) startProxy() {
	ca, _ := a.ensureCA()

	goproxy.GoproxyCa = *ca
	goproxy.OkConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: goproxy.TLSConfigFromCA(ca),
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.OnRequest(goproxy.ReqHostIs("ab.chatgpt.com:443")).HandleConnect(goproxy.AlwaysMitm)

	proxy.Tr = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, "udp", "8.8.8.8:53")
				},
			},
		}).DialContext,
	}

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil {
			return resp
		}
		if strings.Contains(resp.Request.URL.Path, "/v1/initialize") {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[proxy] read body failed: %v", err)
				return resp
			}
			resp.Body.Close()

			modified := stripModels(body)
			resp.Body = io.NopCloser(bytes.NewReader(modified))
			resp.ContentLength = int64(len(modified))
			log.Printf("[proxy] modified /v1/initialize response (%d → %d bytes)", len(body), len(modified))
		}
		return resp
	})

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if strings.Contains(req.URL.Host, "ab.chatgpt.com") {
			log.Printf("[proxy] <- %s %s", req.Method, req.URL.String())
		}
		return req, nil
	})

	log.Printf("[proxy] MITM proxy listening on %s", proxyPort)
	err := http.ListenAndServe(proxyPort, proxy)
	if err != nil {
		log.Printf("[proxy] server error: %v", err)
	}
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

var fallbackData = []byte(`{"feature_gates":{},"dynamic_configs":{},"layer_configs":{},"has_updates":true}`)

func proxyToReal(originalReq *http.Request) ([]byte, error) {
	ips, err := net.LookupHost("ab.chatgpt.com")
	if err != nil {
		return nil, fmt.Errorf("dns lookup: %w", err)
	}
	log.Printf("[proxy] DNS ab.chatgpt.com → %s", ips[0])

	realURL := fmt.Sprintf("https://%s%s", ips[0], originalReq.URL.RequestURI())
	proxyReq, err := http.NewRequest(originalReq.Method, realURL, originalReq.Body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	proxyReq.Header = originalReq.Header.Clone()
	proxyReq.Host = "ab.chatgpt.com"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: "ab.chatgpt.com"},
	}
	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("proxy request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("remote status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (a *App) CmdServer() {
	a.startProxy()
}
