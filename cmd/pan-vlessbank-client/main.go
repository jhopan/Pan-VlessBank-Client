package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"pan-vlessbank-client/internal/client"
	"path/filepath"
	"strconv"
)

const ui = `<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1"><title>Pan-VlessBank Client</title><style>:root{--bg:#0f1115;--s:#161a21;--s2:#1b2027;--b:#2a3038;--t:#e6e9ef;--m:#9aa4b2;--a:#3b82f6;--ok:#4ade80;--okbg:#14301f;--er:#fca5a5;--erbg:#3a1a1c}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--t);font:14px system-ui}.shell{max-width:1180px;margin:auto;padding:26px}.top{display:flex;justify-content:space-between;gap:16px;border-bottom:1px solid var(--b);padding-bottom:20px;margin-bottom:20px}.brand{font-size:20px;font-weight:750}.sub,.hint,label{color:var(--m);font-size:12px}.sub{margin-top:4px}.lock{color:var(--ok);background:var(--okbg);padding:5px 9px;border-radius:7px;height:min-content}.grid{display:grid;grid-template-columns:minmax(0,1fr) 330px;gap:18px}.card{background:var(--s);border:1px solid var(--b);border-radius:12px;padding:18px}h2{font-size:15px;margin:0 0 6px}.stats{display:flex;gap:10px;margin:16px 0}.stat{flex:1;background:var(--s2);border:1px solid var(--b);border-radius:9px;padding:12px;min-width:0}.n{font-size:20px;font-weight:750;overflow-wrap:anywhere}.l{font-size:11px;color:var(--m);margin-top:3px}label{display:block;font-weight:600;margin:12px 0 4px}input{width:100%;background:var(--s2);border:1px solid var(--b);border-radius:8px;color:var(--t);padding:9px}button{min-height:38px;border:0;border-radius:8px;background:var(--a);color:#08111f;font-weight:700;padding:0 12px}.sec{background:var(--s2);color:var(--t);border:1px solid var(--b)}.danger{color:var(--er);border-color:var(--er)}.actions{display:flex;gap:8px;flex-wrap:wrap}.flash{padding:10px 12px;background:var(--erbg);color:var(--er);border-radius:8px}table{width:100%;border-collapse:collapse;margin-top:14px}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid var(--b);font-size:13px}th{color:var(--m);font-size:11px}code{font-family:ui-monospace,Consolas,monospace;overflow-wrap:anywhere}.state{color:var(--ok);font-weight:700;font-size:12px}.empty{color:var(--m);padding:24px;text-align:center}@media(max-width:760px){.shell{padding:14px}.top{display:block}.lock{display:inline-block;margin-top:12px}.grid{grid-template-columns:1fr}.stats{flex-wrap:wrap}.stat{min-width:45%}th:nth-child(4),td:nth-child(4){display:none}}</style><main class="shell"><header class="top"><div><div class="brand">Pan-VlessBank Client</div><div class="sub">Impor SOCKS dan HTTP dari URL Panel Pan-VlessBank</div></div><span class="lock">Loopback only · 127.0.0.1</span></header>{{if .Message}}<p class="flash">{{.Message}}</p>{{end}}<div class="grid"><section class="card"><h2>Proxy terimpor</h2><p class="hint">Start membuka Cloudflare Access TCP untuk setiap proxy publik yang tersedia di server.</p><div class="stats"><div class="stat"><div class="n">{{.Accounts}}</div><div class="l">AKUN</div></div><div class="stat"><div class="n">{{.Active}}</div><div class="l">LISTENER AKTIF</div></div><div class="stat"><div class="n">{{.Server}}</div><div class="l">PANEL SERVER</div></div></div><div class="actions"><form method="post" action="/start"><button>Impor dan Start</button></form><form method="post" action="/stop"><button class="sec danger">Stop Semua</button></form></div>{{if .Rows}}<table><tr><th>Akun</th><th>Jenis</th><th>Endpoint Camoufox</th><th>Hostname Cloudflare</th><th>Status</th></tr>{{range .Rows}}<tr><td>{{.Index}}</td><td>{{.Kind}}</td><td><code>{{.URI}}</code></td><td><code>{{.Host}}</code></td><td class="state">{{.State}} {{.Error}}</td></tr>{{end}}</table>{{else}}<div class="empty">Tempel URL panel dan login di samping, lalu klik Impor dan Start.</div>{{end}}</section><aside class="card"><h2>Impor dari Panel</h2><p class="hint">Password hanya disimpan lokal permission 0600 dan tidak ditampilkan ulang.</p><form method="post" action="/save"><label>URL Panel</label><input name="server" value="{{.Config.ServerURL}}" placeholder="https://panel.example.com"><label>Username panel</label><input name="user" value="{{.Config.Username}}"><label>Password panel</label><input name="password" type="password" placeholder="Simpan ulang bila ingin ganti"><label>Cloudflared path</label><input name="cloudflared" value="{{.Config.Cloudflared}}"><label>SOCKS base</label><input name="socks" value="{{.Config.SocksBase}}"><label>HTTP base</label><input name="http" value="{{.Config.HTTPBase}}"><p class="actions"><button class="sec">Simpan dan impor</button></p></form></aside></div></main>`

var page = template.Must(template.New("p").Parse(ui))

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
	rs := []row{}
	active := 0
	for _, l := range ls {
		if l.State == "running" {
			active++
		}
		scheme := "socks5"
		if l.Kind == "http" {
			scheme = "http"
		}
		rs = append(rs, row{l.Index, l.Kind, l.Host, l.State, l.Error, fmt.Sprintf("%s://127.0.0.1:%d", scheme, l.Port)})
	}
	_ = page.Execute(w, map[string]any{"Config": a.cfg, "Rows": rs, "Accounts": len(ls) / 2, "Active": active, "Server": a.cfg.ServerURL, "Message": a.msg})
	a.msg = ""
}
func main() {
	dir, _ := os.UserConfigDir()
	dir = filepath.Join(dir, "pan-vlessbank-client")
	_ = os.MkdirAll(dir, 0700)
	port := flag.Int("port", 9900, "port panel lokal")
	flag.Parse()
	cfg, e := client.Load(client.ConfigPath(dir))
	if e != nil {
		cfg = client.Config{Cloudflared: "cloudflared", SocksBase: 7001, HTTPBase: 8001, Listen: "127.0.0.1"}
	}
	a := &app{dir: dir, cfg: cfg, mgr: client.New(cfg)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { a.view(w) })
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		a.cfg.ServerURL = r.FormValue("server")
		a.cfg.Username = r.FormValue("user")
		if p := r.FormValue("password"); p != "" {
			a.cfg.Password = p
		}
		a.cfg.Cloudflared = r.FormValue("cloudflared")
		a.cfg.SocksBase, _ = strconv.Atoi(r.FormValue("socks"))
		a.cfg.HTTPBase, _ = strconv.Atoi(r.FormValue("http"))
		a.cfg.Listen = "127.0.0.1"
		if e := client.Save(client.ConfigPath(a.dir), a.cfg); e != nil {
			a.msg = e.Error()
		} else {
			a.mgr = client.New(a.cfg)
			a.msg = "konfigurasi disimpan, klik Impor dan Start"
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
	fmt.Println(http.ListenAndServe(addr, nil))
}
