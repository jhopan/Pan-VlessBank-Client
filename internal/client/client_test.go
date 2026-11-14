package client

import "testing"

func TestRejectPublicPanel(t *testing.T) {
	if (Config{ServerURL: "https://panel.example.com", ClientToken: "token", Listen: "0.0.0.0"}).Validate() == nil {
		t.Fatal("public panel accepted")
	}
}
func TestRejectsNonHTTPSPanel(t *testing.T) {
	if (Config{ServerURL: "http://panel.example.com", ClientToken: "token"}).Validate() == nil {
		t.Fatal("HTTP panel accepted")
	}
}
func TestRequiresClientToken(t *testing.T) {
	if (Config{ServerURL: "https://panel.example.com"}).Validate() == nil {
		t.Fatal("empty client token accepted")
	}
}
