package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTunnelDiagnosticsUsesBearerAndAcceptsHealthyTunnel(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cf/tunnel-diagnostics" || r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`{"connectors":1,"service_active":true,"validation":{"ok":true,"entries":2,"errors":[]}}`))
	}))
	defer s.Close()
	d, e := TunnelDiagnostics(Config{ServerURL: s.URL, ClientToken: "token"})
	if e != nil || !d.Ready() || d.Connectors != 1 {
		t.Fatalf("diagnostic=%#v err=%v", d, e)
	}
}
func TestTunnelDiagnosticsRejectsUnhealthyTunnel(t *testing.T) {
	d := TunnelStatus{Connectors: 0, ServiceActive: false, Validation: TunnelValidation{OK: false, Errors: []string{"config tunnel belum ada"}}}
	if d.Ready() {
		t.Fatal("unhealthy tunnel marked ready")
	}
}
