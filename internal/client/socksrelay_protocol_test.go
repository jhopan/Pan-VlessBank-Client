package client

import (
	"io"
	"net"
	"testing"
)

func TestSocksRelayRejectsClientWithoutNoAuthMethod(t *testing.T) {
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	relay, err := StartSocksRelay("127.0.0.1:0", upstream.Addr().String(), "proxy:secret")
	if err != nil {
		t.Fatal(err)
	}
	defer relay.Close()
	conn, err := net.Dial("tcp", relay.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte{5, 1, 2})
	got := make([]byte, 2)
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "\x05\xff" {
		t.Fatalf("method reply = %v", got)
	}
}

func TestConfigRejectsMalformedProxyAuth(t *testing.T) {
	cfg := Config{ServerURL: "https://panel.example.com", ClientToken: "token", ProxyAuth: "broken", Listen: "127.0.0.1"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("malformed proxy auth accepted")
	}
}
