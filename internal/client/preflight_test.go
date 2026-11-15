package client

import "testing"

func TestCloudflaredPathRejectsMissingBinary(t *testing.T) {
	if err := CheckCloudflared("definitely-not-installed-cloudflared"); err == nil {
		t.Fatal("missing cloudflared accepted")
	}
}
