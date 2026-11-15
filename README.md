# Pan-VlessBank Client

Pan-VlessBank Client membuka bridge loopback untuk proxy Pan-VlessBank melalui Cloudflare Tunnel. Jalankan di komputer user atau mesin Camoufox, bukan server proxy.

Client menjalankan `cloudflared access tcp` per hostname yang token izinkan. VLESS URI, Xray, dan Cloudflare API token tidak pernah masuk client.

## Cara kerja

```text
Camoufox
→ SOCKS tanpa auth, 127.0.0.1:7001
→ relay client loopback
→ SOCKS dengan auth internal
→ cloudflared access tcp, 127.0.0.1:17001
→ socksN.example.com
→ Cloudflare Tunnel
→ Xray server
→ VLESS upstream
→ exit IP akun
```

Relay lokal menyuntikkan credential proxy ke koneksi internal. Camoufox tidak perlu dan tidak boleh diberi username/password SOCKS.

> [!IMPORTANT]
> Firefox dan Camoufox tidak mendukung SOCKS5 username/password. Gunakan URI SOCKS loopback tanpa credential dari client ini.

## Yang diperlukan

| Item | Diperlukan | Catatan |
|---|---:|---|
| Pan-VlessBank Client | Ya | Binary Windows, Linux, atau macOS |
| `cloudflared` | Ya | Ada di `PATH` atau isi path pada panel |
| URL panel HTTPS | Ya | Contoh `https://panel.example.com` |
| Token client | Ya | Read-only, dapat dibatasi scope akun |
| Auth Proxy | Ya | `USER_PROXY:PASSWORD_PROXY`, disimpan lokal permission `0600` |
| VLESS URI, Xray, Cloudflare API token | Tidak | Hanya berada di server |

## Install dan update

Release punya dua channel: `latest` untuk build rolling terbaru, dan `vX.Y.Z` untuk versi tetap. Installer memverifikasi SHA-256 sebelum replace dan tidak mengubah `client.settings.json`.

### Linux dan macOS

```bash
curl -fsSL https://raw.githubusercontent.com/jhopan/Pan-VlessBank-Client/master/install.sh | bash
```

### Windows

```bat
curl.exe -fsSL -o install.bat https://raw.githubusercontent.com/jhopan/Pan-VlessBank-Client/master/install.bat
install.bat
```

Pin versi atau folder:

```bash
VLESSBANK_CLIENT_RELEASE=v0.4.1 VLESSBANK_CLIENT_DIR=$HOME/pan-vlessbank-client \
  curl -fsSL https://raw.githubusercontent.com/jhopan/Pan-VlessBank-Client/master/install.sh | bash
```

Installer tidak memasang `cloudflared`. Install sendiri atau isi path penuh pada panel client.

## Jalankan

```bash
# Windows
pan-vlessbank-client.exe

# Linux/macOS
./pan-vlessbank-client
```

Buka `http://127.0.0.1:9900`.

1. Isi **URL Panel**.
2. Isi **Token Client**.
3. Isi **Auth Proxy**: `USER_PROXY:PASSWORD_PROXY`. Field ini wajib untuk relay SOCKS Camoufox.
4. Isi path `cloudflared` bila tidak ada di `PATH`.
5. Klik **Simpan Konfigurasi**, **Cek Tunnel**, lalu **Impor dan Start**.

> [!IMPORTANT]
> Config client lama tidak memiliki `proxy_auth`. Buka panel, isi **Auth Proxy**, dan simpan ulang sebelum memakai SOCKS Camoufox.

Panel dan seluruh listener bind hanya ke loopback. Credential proxy hanya dipakai relay SOCKS lokal dan disimpan pada config lokal permission `0600`.

## Token client dan scope

Operator membuat token client dari panel atau API admin:

```bash
curl -u ADMIN:PASSWORD -X POST https://panel.example.com/api/client-tokens \
  -d 'label=camoufox-laptop&hours=720&accounts=1,2'
```

| Scope token | Hostname di client |
|---|---|
| `accounts=1,2` | `socks1`, `http1`, `socks2`, `http2` |
| `accounts=2` | hanya `socks2` dan `http2` |
| kosong | seluruh akun |

Token hanya membaca hostname publik dan diagnostics Tunnel. Token tidak dapat membaca VLESS URI, credential proxy, IP LAN, port Xray, daftar token, atau endpoint admin.

Cabut token perangkat yang hilang:

```bash
curl -u ADMIN:PASSWORD -X POST https://panel.example.com/api/client-tokens/revoke \
  -d 'id=TOKEN_ID'
```

## Port lokal

| Urutan scope | SOCKS Camoufox | SOCKS internal cloudflared | HTTP |
|---:|---:|---:|---:|
| 1 | `127.0.0.1:7001` | `127.0.0.1:17001` | `127.0.0.1:8001` |
| 2 | `127.0.0.1:7002` | `127.0.0.1:17002` | `127.0.0.1:8002` |

Gunakan port kolom pertama untuk Camoufox. Port internal `17001+` hanya dipakai client dan tidak perlu disentuh user.

Jika token hanya scope akun 2, akun itu menjadi urutan lokal pertama dan memakai `:7001`/`:8001` secara default. Lihat hostname pada tabel, jangan menebak nomor akun server.

## Camoufox

```python
proxy = {
    "server": "socks5://127.0.0.1:7001",
}
```

Jangan isi username/password pada konfigurasi Camoufox. Relay menyuntikkan **Auth Proxy** yang disimpan client ke Xray.

HTTP bridge masih tersedia untuk aplikasi HTTP proxy:

```python
proxy = {
    "server": "http://127.0.0.1:8001",
    "username": "USER_PROXY",
    "password": "PASSWORD_PROXY",
}
```

## Verifikasi

Camoufox SOCKS tanpa credential harus dapat membuka endpoint IP:

```bash
curl -sS --socks5-hostname 127.0.0.1:7001 \
  --max-time 25 'http://ip-api.com/json/?fields=query,countryCode'
```

Output berbentuk:

```json
{"countryCode":"CC","query":"EXIT_IP"}
```

## Troubleshooting

| Gejala | Penyebab umum | Perbaikan |
|---|---|---|
| `token client ditolak atau sudah dicabut` | Token salah, expired, atau revoked | Minta token baru |
| `tunnel server belum siap` | Connector, service, atau ingress Tunnel bermasalah | Operator cek diagnostics Tunnel |
| `cloudflared tidak ditemukan` | Binary tidak ada di PATH atau path salah | Install `cloudflared` atau isi path penuh |
| `address already in use` | Listener lama memakai port | Klik Stop Semua atau ganti base port |
| Camoufox tidak konek ke SOCKS | Auth Proxy kosong atau salah | Isi `USER_PROXY:PASSWORD_PROXY`, simpan config, lalu Start ulang |
| Semua port punya IP sama | Ingress server salah atau upstream exit sama | Operator cek hostname-per-port dan `vlessbank check` |

## Keamanan

- Panel client, relay SOCKS, bridge cloudflared, dan HTTP listener hanya bind `127.0.0.1`.
- Auth Proxy disimpan lokal agar Camoufox dapat memakai SOCKS tanpa SOCKS5 auth.
- Jangan salin config client ke perangkat lain.
- Jangan set listener ke `0.0.0.0`.
- Klik **Stop Semua** saat selesai.

## Build dari source

```bash
go test ./... -count=1
go vet ./...
go build -o pan-vlessbank-client.exe ./cmd/pan-vlessbank-client
```

Build release lintas platform ada di `.github/workflows/release.yml`.