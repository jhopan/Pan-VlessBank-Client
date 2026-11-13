package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
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
	Domain      string `json:"domain"`
	Accounts    int    `json:"accounts"`
	Cloudflared string `json:"cloudflared"`
	SocksBase   int    `json:"socks_base"`
	HTTPBase    int    `json:"http_base"`
	Listen      string `json:"listen"`
}

func (c Config) Validate() error {
	if c.Domain == "" || !strings.Contains(c.Domain, ".") || c.Accounts < 1 {
		return errors.New("domain atau jumlah akun tidak valid")
	}
	if c.Listen != "" && !isLoopback(c.Listen) {
		return errors.New("panel client wajib loopback")
	}
	return nil
}
func isLoopback(s string) bool { ip := net.ParseIP(s); return ip != nil && ip.IsLoopback() }
func Jobs(domain string, accounts, socksBase, httpBase int) []Job {
	jobs := make([]Job, 0, accounts*2)
	for i := 1; i <= accounts; i++ {
		jobs = append(jobs, Job{i, "socks", fmt.Sprintf("socks%d.%s", i, domain), socksBase + i - 1}, Job{i, "http", fmt.Sprintf("http%d.%s", i, domain), httpBase + i - 1})
	}
	return jobs
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
}
type Manager struct {
	mu        sync.Mutex
	cfg       Config
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
func (m *Manager) Start() error {
	if e := m.cfg.Validate(); e != nil {
		return e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range Jobs(m.cfg.Domain, m.cfg.Accounts, m.cfg.SocksBase, m.cfg.HTTPBase) {
		if _, ok := m.listeners[j.Port]; ok {
			continue
		}
		cmd := exec.Command(m.cfg.Cloudflared, "access", "tcp", "--hostname", j.Host, "--url", fmt.Sprintf("127.0.0.1:%d", j.Port))
		if e := cmd.Start(); e != nil {
			return fmt.Errorf("%s: %w", j.Host, e)
		}
		l := &Listener{Job: j, State: "running", PID: cmd.Process.Pid, Started: time.Now(), cmd: cmd}
		m.listeners[j.Port] = l
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
