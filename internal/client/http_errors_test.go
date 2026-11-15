package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscoverExplainsRejectedClientToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()
	_, err := Discover(Config{ServerURL: server.URL, ClientToken: "token"})
	if err == nil || !strings.Contains(err.Error(), "token client ditolak") {
		t.Fatalf("err = %v", err)
	}
}

func TestTunnelDiagnosticsExplainsUnavailableEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	defer server.Close()
	_, err := TunnelDiagnostics(Config{ServerURL: server.URL, ClientToken: "token"})
	if err == nil || !strings.Contains(err.Error(), "belum mendukung") {
		t.Fatalf("err = %v", err)
	}
}
