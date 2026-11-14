# Pan-VlessBank Client

Client Cloudflare Access TCP lintas Windows, Linux, dan macOS untuk proxy yang di-host Pan-VlessBank. Jalankan di mesin Camoufox, bukan server proxy.

## Cara kerja

```text
Camoufox → 127.0.0.1:7001 → cloudflared access tcp → Cloudflare → Pan-VlessBank/Xray
```

Cloudflare hanya transport. IP exit tetap ditentukan akun VLESS upstream.

## Buat token client di server

Login sebagai admin pada Panel Pan-VlessBank. Buat token client read-only melalui API admin:

```bash
curl -u admin:password -X POST https://panel.example.com/api/client-tokens \
  -d 'label=camoufox-laptop&hours=720'
```

Simpan nilai `token` yang dikembalikan. Nilai ini hanya muncul sekali. Token hanya dapat membaca hostname proxy publik dari `GET /api/client/proxies`; token tidak dapat mengubah akun, Xray, tunnel, atau konfigurasi panel. Cabut token bila perangkat hilang atau tidak dipakai lagi:

```bash
curl -u admin:password -X POST https://panel.example.com/api/client-tokens/revoke \
  -d 'id=TOKEN_ID'
```

## Jalankan

Install `cloudflared` di mesin client, lalu jalankan binary:

```bash
./pan-vlessbank-client
```

Buka `http://127.0.0.1:9900`, lalu isi:

1. URL Panel Pan-VlessBank, misalnya `https://panel.example.com`.
2. Token client read-only.
3. Path `cloudflared` bila tidak ada di PATH.

Klik **Simpan konfigurasi**, lalu **Impor dan Start**. Client mengambil `/api/client/proxies` memakai `Authorization: Bearer <token>`, otomatis membaca hostname SOCKS dan HTTP publik, lalu membuka listener lokal. Password admin panel, API token Cloudflare, UUID VLESS, dan link VLESS tidak masuk client.

Port lokal default:

| Akun | SOCKS | HTTP |
|---:|---:|---:|
| 1 | `127.0.0.1:7001` | `127.0.0.1:8001` |
| 2 | `127.0.0.1:7002` | `127.0.0.1:8002` |

Cloudflared akan meminta login Cloudflare Access di mesin client jika sesi belum ada.

## Camoufox

```python
proxy = {
    "server": "socks5://127.0.0.1:7001",
    "username": "USER_PROXY",
    "password": "PASSWORD_PROXY",
}
```

HTTP:

```python
proxy = {
    "server": "http://127.0.0.1:8001",
    "username": "USER_PROXY",
    "password": "PASSWORD_PROXY",
}
```

Panel client hanya bind `127.0.0.1`. Config lokal disimpan permission `0600`, dan token tidak ditampilkan ulang.
