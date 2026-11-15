package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerStartsAndStopsScopedListeners(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/proxies" || r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"proxies":[{"idx":2,"name":"account","socks_host":"socks2.example.com","http_host":"http2.example.com"}]}`))
	}))
	defer server.Close()
	stub := filepath.Join(t.TempDir(), "cloudflared.cmd")
	if err := os.WriteFile(stub, []byte("@echo off\r\nif \"%1\"==\"--version\" exit /b 0\r\nping -n 30 127.0.0.1 >nul\r\n"), 0700); err != nil {
		t.Fatal(err)
	}
	manager := New(Config{ServerURL: server.URL, ClientToken: "token", Cloudflared: stub, SocksBase: 7001, HTTPBase: 8001, Listen: "127.0.0.1"})
	if err := manager.Start(); err != nil {
		t.Fatal(err)
	}
	listeners := manager.Snapshot()
	if len(listeners) != 2 || listeners[0].State != "running" || listeners[1].State != "running" {
		t.Fatalf("listeners = %#v", listeners)
	}
	manager.Stop()
	if listeners := manager.Snapshot(); len(listeners) != 0 {
		t.Fatalf("listeners remain after stop: %#v", listeners)
	}
}
