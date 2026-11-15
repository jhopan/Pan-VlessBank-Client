package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Job struct {
	Index      int
	Kind, Host string
	Port       int
}
type Config struct {
	ServerURL   string `json:"server_url"`
	ClientToken string `json:"client_token"`
	ProxyAuth   string `json:"proxy_auth"`
	Cloudflared string `json:"cloudflared"`
	SocksBase   int    `json:"socks_base"`
	HTTPBase    int    `json:"http_base"`
	Listen      string `json:"listen"`
}

func (c Config) Validate() error {
	u, e := url.Parse(c.ServerURL)
	if e != nil || u.Host == "" || (u.Scheme != "https" && !(u.Scheme == "http" && isLoopbackHost(u.Hostname()))) {
		return errors.New("URL panel harus HTTPS lengkap")
	}
	if c.ClientToken == "" {
		return errors.New("token client kosong")
	}
	if c.Listen != "" && !isLoopback(c.Listen) {
		return errors.New("panel client wajib loopback")
	}
	if c.ProxyAuth != "" {
		user, pass, ok := strings.Cut(c.ProxyAuth, ":")
		if !ok || user == "" || pass == "" || len(user) > 255 || len(pass) > 255 {
			return errors.New("auth proxy harus USER:PASSWORD")
		}
	}
	return nil
}
func isLoopback(s string) bool     { ip := net.ParseIP(s); return ip != nil && ip.IsLoopback() }
func isLoopbackHost(s string) bool { ip := net.ParseIP(s); return ip != nil && ip.IsLoopback() }
func Jobs(domain string, accounts, socks, http int) []Job {
	out := make([]Job, 0, accounts*2)
	for i := 1; i <= accounts; i++ {
		out = append(out, Job{i, "socks", fmt.Sprintf("socks%d.%s", i, domain), socks + i - 1}, Job{i, "http", fmt.Sprintf("http%d.%s", i, domain), http + i - 1})
	}
	return out
}

type remote struct {
	Proxies []struct {
		Index int    `json:"idx"`
		Name  string `json:"name"`
		Socks string `json:"socks_host"`
		HTTP  string `json:"http_host"`
	} `json:"proxies"`
}

func Discover(c Config) ([]Job, error) {
	if e := c.Validate(); e != nil {
		return nil, e
	}
	base := strings.TrimRight(c.ServerURL, "/")
	req, e := http.NewRequest(http.MethodGet, base+"/api/client/proxies", nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+c.ClientToken)
	res, e := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
			return nil, errors.New("token client ditolak atau sudah dicabut")
		}
		return nil, fmt.Errorf("panel mengembalikan HTTP %d", res.StatusCode)
	}
	var r remote
	if e := json.NewDecoder(res.Body).Decode(&r); e != nil {
		return nil, e
	}
	jobs := make([]Job, 0, len(r.Proxies)*2)
	for i, p := range r.Proxies {
		idx := p.Index
		if idx == 0 {
			idx = i + 1
		}
		if p.Socks == "" || p.HTTP == "" {
			return nil, errors.New("server belum mengatur tunnel proxy publik")
		}
		jobs = append(jobs, Job{idx, "socks", p.Socks, c.SocksBase + i}, Job{idx, "http", p.HTTP, c.HTTPBase + i})
	}
	if len(jobs) == 0 {
		return nil, errors.New("server tidak memiliki akun proxy")
	}
	return jobs, nil
}
func Load(path string) (Config, error) {
	var c Config
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	return c, json.Unmarshal(b, &c)
}
func Save(path string, c Config) error {
	if e := c.Validate(); e != nil {
		return e
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0600)
}
func ConfigPath(dir string) string { return filepath.Join(dir, "client.settings.json") }

type Listener struct {
	Job
	State, Error string
	PID          int
	Started      time.Time
	cmd          *exec.Cmd
	relay        *SocksRelay
}
type Manager struct {
	mu        sync.Mutex
	cfg       Config
	jobs      []Job
	listeners map[int]*Listener
}

func New(c Config) *Manager {
	if c.Cloudflared == "" {
		c.Cloudflared = "cloudflared"
	}
	if c.SocksBase == 0 {
		c.SocksBase = 7001
	}
	if c.HTTPBase == 0 {
		c.HTTPBase = 8001
	}
	return &Manager{cfg: c, listeners: map[int]*Listener{}}
}
func (m *Manager) Refresh() error {
	jobs, e := Discover(m.cfg)
	if e != nil {
		return e
	}
	m.mu.Lock()
	m.jobs = jobs
	m.mu.Unlock()
	return nil
}
func (m *Manager) Start() error {
	if e := CheckCloudflared(m.cfg.Cloudflared); e != nil {
		return e
	}
	if e := m.Refresh(); e != nil {
		return e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	started := []*Listener{}
	for _, j := range m.jobs {
		if _, ok := m.listeners[j.Port]; ok {
			continue
		}
		bridgePort := j.Port
		if j.Kind == "socks" {
			bridgePort += 10000
		}
		cmd := exec.Command(m.cfg.Cloudflared, "access", "tcp", "--hostname", j.Host, "--url", fmt.Sprintf("127.0.0.1:%d", bridgePort))
		if e := cmd.Start(); e != nil {
			for _, l := range started {
				if l.relay != nil {
					_ = l.relay.Close()
				}
				if l.cmd.Process != nil {
					_ = l.cmd.Process.Kill()
				}
				delete(m.listeners, l.Port)
			}
			return fmt.Errorf("listener %s gagal: %w", j.Host, e)
		}
		l := &Listener{Job: j, State: "running", PID: cmd.Process.Pid, Started: time.Now(), cmd: cmd}
		if j.Kind == "socks" && m.cfg.ProxyAuth != "" {
			relay, e := StartSocksRelay(fmt.Sprintf("127.0.0.1:%d", j.Port), fmt.Sprintf("127.0.0.1:%d", bridgePort), m.cfg.ProxyAuth)
			if e != nil {
				_ = cmd.Process.Kill()
				return fmt.Errorf("relay SOCKS %s gagal: %w", j.Host, e)
			}
			l.relay = relay
		}
		m.listeners[j.Port] = l
		started = append(started, l)
		go m.wait(l)
	}
	return nil
}

func (m *Manager) wait(l *Listener) {
	e := l.cmd.Wait()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listeners[l.Port] != l {
		return
	}
	if l.relay != nil {
		_ = l.relay.Close()
	}
	l.State = "error"
	l.PID = 0
	if e != nil {
		l.Error = e.Error()
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.listeners {
		if l.relay != nil {
			_ = l.relay.Close()
		}
		if l.cmd.Process != nil {
			_ = l.cmd.Process.Kill()
		}
	}
	m.listeners = map[int]*Listener{}
}
func (m *Manager) Snapshot() []Listener {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Listener, 0, len(m.listeners))
	for _, l := range m.listeners {
		out = append(out, *l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}
