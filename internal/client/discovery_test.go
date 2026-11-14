package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverImportsPublicProxyHosts(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "panel" || p != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":2,"proxies":[{"idx":1,"public":{"socks":"socks1.proxy.example.com","http":"http1.proxy.example.com"}},{"idx":2,"public":{"socks":"socks2.proxy.example.com","http":"http2.proxy.example.com"}}]}`))
	}))
	defer s.Close()
	jobs, err := Discover(Config{ServerURL: s.URL, Username: "panel", Password: "secret", SocksBase: 7001, HTTPBase: 8001})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 4 || jobs[0].Host != "socks1.proxy.example.com" || jobs[0].Port != 7001 || jobs[3].Host != "http2.proxy.example.com" || jobs[3].Port != 8002 {
		t.Fatalf("jobs=%#v", jobs)
	}
}
