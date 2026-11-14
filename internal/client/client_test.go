package client

import "testing"

func TestRejectPublicPanel(t *testing.T) {
	if (Config{ServerURL: "https://panel.example.com", Username: "u", Password: "p", Listen: "0.0.0.0"}).Validate() == nil {
		t.Fatal("public panel accepted")
	}
}
func TestRejectsNonHTTPSPanel(t *testing.T) {
	if (Config{ServerURL: "http://panel.example.com", Username: "u", Password: "p"}).Validate() == nil {
		t.Fatal("HTTP panel accepted")
	}
}
