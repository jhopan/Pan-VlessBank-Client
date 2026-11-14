# Pan-VlessBank Client

Client Cloudflare Access TCP lintas Windows, Linux, dan macOS untuk proxy yang di-host Pan-VlessBank. Jalankan di mesin Camoufox, bukan server proxy.

## Cara kerja

```text
Camoufox → 127.0.0.1:7001 → cloudflared access tcp → Cloudflare → Pan-VlessBank/Xray
```

Cloudflare hanya transport. IP exit tetap ditentukan akun VLESS upstream.

## Jalankan

Install `cloudflared` di mesin client, lalu jalankan binary:

```bash
./pan-vlessbank-client
```

Buka `http://127.0.0.1:9900`, lalu isi:

1. URL Panel Pan-VlessBank, misalnya `https://panel.example.com`.
2. Username dan password Panel.
3. Path `cloudflared` bila tidak ada di PATH.

Klik **Simpan dan impor**, lalu **Impor dan Start**. Client mengambil `/api/proxies` dari panel, otomatis membaca setiap hostname SOCKS dan HTTP publik, dan membuka listener lokal. API token Cloudflare, UUID VLESS, dan link VLESS tidak dikirim ke client.

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

Panel client hanya bind `127.0.0.1`. Config lokal disimpan permission `0600`, dan password panel tidak ditampilkan ulang.
