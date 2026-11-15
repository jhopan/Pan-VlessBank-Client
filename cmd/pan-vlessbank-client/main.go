package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"pan-vlessbank-client/internal/client"
)

//go:embed ui.html
var ui string

var page = template.Must(template.New("p").Parse(ui))

type row struct {
	Index                                int
	Kind, Host, State, Error, URI, Class string
}
type app struct {
	dir       string
	cfg       client.Config
	mgr       *client.Manager
	msg, diag string
}

func (a *app) view(w http.ResponseWriter) {
	listeners := a.mgr.Snapshot()
	rows := make([]row, 0, len(listeners))
	active := 0
	for _, listener := range listeners {
		if listener.State == "running" {
			active++
		}
		scheme, class := "socks5", "run"
		if listener.Kind == "http" {
			scheme = "http"
		}
		if listener.State != "running" {
			class = "err"
		}
		rows = append(rows, row{listener.Index, listener.Kind, listener.Host, listener.State, listener.Error, fmt.Sprintf("%s://127.0.0.1:%d", scheme, listener.Port), class})
	}
	_ = page.Execute(w, map[string]any{"Config": a.cfg, "Rows": rows, "Accounts": len(listeners) / 2, "Active": active, "Server": a.cfg.ServerURL, "Message": a.msg, "Diag": a.diag})
	a.msg = ""
}

func main() {
	dir, _ := os.UserConfigDir()
	dir = filepath.Join(dir, "pan-vlessbank-client")
	_ = os.MkdirAll(dir, 0700)
	port := flag.Int("port", 9900, "port panel lokal")
	flag.Parse()
	cfg, err := client.Load(client.ConfigPath(dir))
	if err != nil {
		cfg = client.Config{Cloudflared: "cloudflared", SocksBase: 7001, HTTPBase: 8001, Listen: "127.0.0.1"}
	}
	a := &app{dir: dir, cfg: cfg, mgr: client.New(cfg)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { a.view(w) })
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		a.cfg.ServerURL = r.FormValue("server")
		if token := r.FormValue("token"); token != "" {
			a.cfg.ClientToken = token
		}
		if auth := r.FormValue("proxy_auth"); auth != "" {
			a.cfg.ProxyAuth = auth
		}
		a.cfg.Cloudflared = r.FormValue("cloudflared")
		a.cfg.SocksBase, _ = strconv.Atoi(r.FormValue("socks"))
		a.cfg.HTTPBase, _ = strconv.Atoi(r.FormValue("http"))
		a.cfg.Listen = "127.0.0.1"
		if err := client.Save(client.ConfigPath(a.dir), a.cfg); err != nil {
			a.msg = err.Error()
		} else {
			a.mgr = client.New(a.cfg)
			a.msg = "konfigurasi disimpan, cek tunnel lalu start listener"
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		d, err := client.TunnelDiagnostics(a.cfg)
		if err != nil {
			a.msg = err.Error()
		} else {
			a.diag = fmt.Sprintf("connector: %d\nservice: %t\ningress valid: %t\nentries: %d\naccess: %s", d.Connectors, d.ServiceActive, d.Validation.OK, d.Validation.Entries, d.Access)
			a.msg = "diagnostics tunnel diperbarui"
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		d, err := client.TunnelDiagnostics(a.cfg)
		if err != nil {
			a.msg = err.Error()
		} else if !d.Ready() {
			a.msg = "tunnel server belum siap: cek diagnostics"
		} else if err := a.mgr.Start(); err != nil {
			a.msg = err.Error()
		} else {
			a.msg = "listener tunnel dimulai"
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		a.mgr.Stop()
		a.msg = "semua listener dihentikan"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	addr := "127.0.0.1:" + strconv.Itoa(*port)
	fmt.Println("Pan-VlessBank Client:", "http://"+addr)
	fmt.Println(http.ListenAndServe(addr, nil))
}
