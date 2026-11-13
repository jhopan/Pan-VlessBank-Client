package client

import "testing"

func TestJobsBuildStablePorts(t *testing.T) {
	got := Jobs("proxy.example.com", 2, 7001, 8001)
	want := []Job{{1, "socks", "socks1.proxy.example.com", 7001}, {1, "http", "http1.proxy.example.com", 8001}, {2, "socks", "socks2.proxy.example.com", 7002}, {2, "http", "http2.proxy.example.com", 8002}}
	if len(got) != len(want) {
		t.Fatalf("jobs=%d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("job=%#v want=%#v", got[i], want[i])
		}
	}
}
func TestRejectPublicPanel(t *testing.T) {
	if (Config{Domain: "proxy.example.com", Accounts: 1, Listen: "0.0.0.0"}).Validate() == nil {
		t.Fatal("public panel accepted")
	}
}
