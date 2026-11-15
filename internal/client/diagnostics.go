package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type TunnelValidation struct {
	OK      bool     `json:"ok"`
	Entries int      `json:"entries"`
	Errors  []string `json:"errors"`
}
type TunnelStatus struct {
	Connectors    int              `json:"connectors"`
	ServiceActive bool             `json:"service_active"`
	Validation    TunnelValidation `json:"validation"`
	Access        string           `json:"access"`
}

func (d TunnelStatus) Ready() bool { return d.Connectors > 0 && d.ServiceActive && d.Validation.OK }
func TunnelDiagnostics(c Config) (TunnelStatus, error) {
	if e := c.Validate(); e != nil {
		return TunnelStatus{}, e
	}
	req, e := http.NewRequest(http.MethodGet, strings.TrimRight(c.ServerURL, "/")+"/api/cf/tunnel-diagnostics", nil)
	if e != nil {
		return TunnelStatus{}, e
	}
	req.Header.Set("Authorization", "Bearer "+c.ClientToken)
	res, e := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if e != nil {
		return TunnelStatus{}, e
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
			return TunnelStatus{}, errors.New("token client ditolak atau sudah dicabut")
		}
		if res.StatusCode == http.StatusNotFound {
			return TunnelStatus{}, errors.New("server belum mendukung diagnostics tunnel")
		}
		return TunnelStatus{}, fmt.Errorf("diagnostic tunnel HTTP %d", res.StatusCode)
	}
	var d TunnelStatus
	if e := json.NewDecoder(res.Body).Decode(&d); e != nil {
		return TunnelStatus{}, e
	}
	return d, nil
}
func CheckCloudflared(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path cloudflared kosong")
	}
	p, e := exec.LookPath(path)
	if e != nil {
		return fmt.Errorf("cloudflared tidak ditemukan: %s", path)
	}
	cmd := exec.Command(p, "--version")
	if e := cmd.Run(); e != nil {
		return fmt.Errorf("cloudflared tidak bisa dijalankan: %w", e)
	}
	return nil
}
