# Pan-VlessBank Client

Client Cloudflare Access TCP lintas Windows, Linux, dan macOS untuk proxy yang di-host Pan-VlessBank. Client ini berjalan di mesin Camoufox, bukan server proxy.

## Cara kerja

```text
Camoufox → 127.0.0.1:7001 → cloudflared access tcp → Cloudflare → Pan-VlessBank/Xray
```

Cloudflare hanya transport. IP exit tetap ditentukan akun VLESS upstream.

## Jalankan

Install `cloudflared` di mesin client, lalu jalankan binary:

```bash
./pan-vlessbank-client --domain proxy.example.com --accounts 2
```

Buka `http://127.0.0.1:9900`. Simpan domain proxy lalu klik **Start Listener**. Cloudflared akan meminta login Cloudflare Access di mesin ini jika belum ada sesi.

Port lokal default:

| Akun | SOCKS | HTTP |
|---:|---:|---:|
| 1 | `127.0.0.1:7001` | `127.0.0.1:8001` |
| 2 | `127.0.0.1:7002` | `127.0.0.1:8002` |

## Camoufox

```python
proxy = {
    "server": "socks5://127.0.0.1:7001",
    "username": "USER",
    "password": "PASSWORD",
}
```

HTTP:

```python
proxy = {
    "server": "http://127.0.0.1:8001",
    "username": "USER",
    "password": "PASSWORD",
}
```

Client panel hanya bind `127.0.0.1`, config lokal disimpan permission `0600`, dan client tidak memerlukan API token Cloudflare atau link VLESS.
