package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverUsesBearerClientToken(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/proxies" || r.Header.Get("Authorization") != "Bearer client-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"proxies":[{"idx":1,"name":"one","socks_host":"socks1.proxy.example.com","http_host":"http1.proxy.example.com"},{"idx":2,"name":"two","socks_host":"socks2.proxy.example.com","http_host":"http2.proxy.example.com"}]}`))
	}))
	defer s.Close()
	jobs, err := Discover(Config{ServerURL: s.URL, ClientToken: "client-token", SocksBase: 7001, HTTPBase: 8001})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 4 || jobs[0].Host != "socks1.proxy.example.com" || jobs[0].Port != 7001 || jobs[3].Host != "http2.proxy.example.com" || jobs[3].Port != 8002 {
		t.Fatalf("jobs=%#v", jobs)
	}
}
