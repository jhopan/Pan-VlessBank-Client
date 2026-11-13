package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"pan-vlessbank-client/internal/client"
)

const ui = `<!doctype html><html lang="id"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Pan-VlessBank Client</title><style>
:root{--bg:#0f1115;--surface:#161a21;--surface2:#1b2027;--border:#2a3038;--text:#e6e9ef;--muted:#9aa4b2;--accent:#3b82f6;--ok:#4ade80;--okbg:#14301f;--err:#fca5a5;--errbg:#3a1a1c}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px system-ui}.shell{max-width:1180px;margin:auto;padding:26px}.top{display:flex;justify-content:space-between;gap:16px;align-items:start;border-bottom:1px solid var(--border);padding-bottom:20px;margin-bottom:20px}.brand{font-size:20px;font-weight:750}.sub{color:var(--muted);font-size:13px;margin-top:4px}.lock{color:var(--ok);font-size:12px;background:var(--okbg);padding:5px 9px;border-radius:7px}.grid{display:grid;grid-template-columns:minmax(0,1fr) 330px;gap:18px}.card{background:var(--surface);border:1px solid var(--border);border-radius:12px;padding:18px}h2{font-size:15px;margin:0 0 5px}.hint{color:var(--muted);font-size:12px;margin:0 0 14px}.stats{display:flex;gap:10px;margin:16px 0}.stat{flex:1;background:var(--surface2);border:1px solid var(--border);border-radius:9px;padding:12px}.n{font-size:20px;font-weight:750}.l{font-size:11px;color:var(--muted);margin-top:3px}label{color:var(--muted);font-size:12px;font-weight:600;display:block;margin:12px 0 4px}input{width:100%;background:var(--surface2);border:1px solid var(--border);border-radius:8px;color:var(--text);padding:9px}button{min-height:38px;border-radius:8px;border:0;background:var(--accent);color:#08111f;font-weight:700;padding:0 12px;cursor:pointer}.sec{background:var(--surface2);color:var(--text);border:1px solid var(--border)}.danger{color:var(--err);border-color:var(--err)}.actions{display:flex;gap:8px;flex-wrap:wrap}.flash{margin:0 0 15px;padding:10px 12px;border-radius:8px;background:var(--errbg);color:var(--err)}table{width:100%;border-collapse:collapse;margin-top:14px}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid var(--border);font-size:13px}th{color:var(--muted);font-size:11px}code{font-family:ui-monospace,Consolas,monospace;font-size:12px;overflow-wrap:anywhere}.state{font-size:12px;font-weight:700;color:var(--ok)}.empty{color:var(--muted);padding:24px;text-align:center}@media(max-width:760px){.shell{padding:14px}.grid{grid-template-columns:1fr}.top{display:block}.lock{display:inline-block;margin-top:12px}th:nth-child(3),td:nth-child(3){display:none}}</style><body><main class="shell"><header class="top"><div><div class="brand">Pan-VlessBank Client</div><div class="sub">Cloudflare Access bridge untuk Camoufox dan aplikasi proxy</div></div><span class="lock">Loopback only · 127.0.0.1</span></header>{{if .Message}}<p class="flash">{{.Message}}</p>{{end}}<div class="grid"><section><div class="card"><h2>Listener lokal</h2><p class="hint">Start akan membuka cloudflared Access TCP. Browser login Cloudflare muncul di mesin ini bila sesi belum ada.</p><div class="stats"><div class="stat"><div class="n">{{.Accounts}}</div><div class="l">AKUN</div></div><div class="stat"><div class="n">{{.Active}}</div><div class="l">LISTENER AKTIF</div></div><div class="stat"><div class="n">{{.Domain}}</div><div class="l">DOMAIN PROXY</div></div></div><div class="actions"><form method="post" action="/start"><button>Start Listener</button></form><form method="post" action="/stop"><button class="sec danger">Stop Semua</button></form></div>{{if .Rows}}<table><tr><th>Akun</th><th>Jenis</th><th>Endpoint Camoufox</th><th>Hostname Cloudflare</th><th>Status</th></tr>{{range .Rows}}<tr><td>{{.Index}}</td><td>{{.Kind}}</td><td><code>{{.URI}}</code></td><td><code>{{.Host}}</code></td><td class="state">{{.State}} {{.Error}}</td></tr>{{end}}</table>{{else}}<div class="empty">Belum ada listener. Simpan domain dan jumlah akun, lalu klik Start Listener.</div>{{end}}</div></section><aside class="card"><h2>Konfigurasi</h2><p class="hint">Simpan di file lokal permission 0600. Tidak ada API token atau link VLESS di client.</p><form method="post" action="/save"><label>Domain proxy penuh</label><input name="domain" value="{{.Config.Domain}}" placeholder="proxy.example.com"><label>Jumlah akun</label><input name="accounts" value="{{.Config.Accounts}}" inputmode="numeric"><label>Cloudflared path</label><input name="cloudflared" value="{{.Config.Cloudflared}}"><label>SOCKS base</label><input name="socks" value="{{.Config.SocksBase}}" inputmode="numeric"><label>HTTP base</label><input name="http" value="{{.Config.HTTPBase}}" inputmode="numeric"><p class="actions" style="margin-top:15px"><button class="sec">Simpan konfigurasi</button></p></form></aside></div></main></body></html>`

var page = template.Must(template.New("client").Parse(ui))

type row struct {
	Index                         int
	Kind, Host, State, Error, URI string
}
type app struct {
	dir string
	cfg client.Config
	mgr *client.Manager
	msg string
}

func (a *app) view(w http.ResponseWriter) {
	ls := a.mgr.Snapshot()
	rows := make([]row, 0, len(ls))
	active := 0
	for _, l := range ls {
		if l.State == "running" {
			active++
		}
		scheme := "socks5"
		if l.Kind == "http" {
			scheme = "http"
		}
		rows = append(rows, row{l.Index, l.Kind, l.Host, l.State, l.Error, fmt.Sprintf("%s://127.0.0.1:%d", scheme, l.Port)})
	}
	_ = page.Execute(w, map[string]any{"Config": a.cfg, "Rows": rows, "Accounts": a.cfg.Accounts, "Active": active, "Domain": a.cfg.Domain, "Message": a.msg})
	a.msg = ""
}
func main() {
	dir, _ := os.UserConfigDir()
	dir = filepath.Join(dir, "pan-vlessbank-client")
	_ = os.MkdirAll(dir, 0700)
	port := flag.Int("port", 9900, "port panel lokal")
	domain := flag.String("domain", "", "domain proxy")
	accounts := flag.Int("accounts", 0, "jumlah akun")
	cf := flag.String("cloudflared", "cloudflared", "path cloudflared")
	flag.Parse()
	cfg, e := client.Load(client.ConfigPath(dir))
	if e != nil {
		cfg = client.Config{Domain: *domain, Accounts: *accounts, Cloudflared: *cf, SocksBase: 7001, HTTPBase: 8001, Listen: "127.0.0.1"}
	}
	if *domain != "" {
		cfg.Domain = *domain
	}
	if *accounts > 0 {
		cfg.Accounts = *accounts
	}
	a := &app{dir: dir, cfg: cfg, mgr: client.New(cfg)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { a.view(w) })
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		a.cfg.Domain = r.FormValue("domain")
		a.cfg.Cloudflared = r.FormValue("cloudflared")
		a.cfg.Listen = "127.0.0.1"
		a.cfg.Accounts, _ = strconv.Atoi(r.FormValue("accounts"))
		a.cfg.SocksBase, _ = strconv.Atoi(r.FormValue("socks"))
		a.cfg.HTTPBase, _ = strconv.Atoi(r.FormValue("http"))
		if e := client.Save(client.ConfigPath(a.dir), a.cfg); e != nil {
			a.msg = e.Error()
		} else {
			a.mgr = client.New(a.cfg)
			a.msg = "konfigurasi disimpan"
		}
		http.Redirect(w, r, "/", 303)
	})
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if e := a.mgr.Start(); e != nil {
			a.msg = e.Error()
		} else {
			a.msg = "listener dimulai, login Cloudflare Access bila diminta"
		}
		http.Redirect(w, r, "/", 303)
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		a.mgr.Stop()
		a.msg = "semua listener dihentikan"
		http.Redirect(w, r, "/", 303)
	})
	addr := "127.0.0.1:" + strconv.Itoa(*port)
	fmt.Println("Pan-VlessBank Client:", "http://"+addr)
	if e := http.ListenAndServe(addr, nil); e != nil {
		fmt.Println(e)
	}
}
