#!/usr/bin/env bash
set -euo pipefail

REPO="${VLESSBANK_CLIENT_REPO:-jhopan/Pan-VlessBank-Client}"
RELEASE="${VLESSBANK_CLIENT_RELEASE:-latest}"
APP_DIR="${VLESSBANK_CLIENT_DIR:-$HOME/pan-vlessbank-client}"

say() { printf '%s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
platform() {
  case "$(uname -s)/$(uname -m)" in
    Linux/x86_64|Linux/amd64) printf 'linux_amd64' ;;
    Linux/aarch64|Linux/arm64) printf 'linux_arm64' ;;
    Darwin/x86_64|Darwin/amd64) printf 'darwin_amd64' ;;
    Darwin/arm64) printf 'darwin_arm64' ;;
    *) die "platform tidak didukung: $(uname -s) $(uname -m)" ;;
  esac
}
command -v curl >/dev/null || die 'curl belum terpasang'
ASSET="pan-vlessbank-client_$(platform)"
BASE_URL="https://github.com/$REPO/releases/download/$RELEASE"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
curl -fsSL --retry 3 -o "$TMP_DIR/$ASSET" "$BASE_URL/$ASSET" || die "download gagal: $BASE_URL/$ASSET"
curl -fsSL --retry 3 -o "$TMP_DIR/checksums.txt" "$BASE_URL/checksums.txt" || die 'download checksums gagal'
EXPECTED=$(awk -v asset="$ASSET" '$2==asset {print $1}' "$TMP_DIR/checksums.txt")
[ -n "$EXPECTED" ] || die "checksum $ASSET tidak ditemukan"
if command -v sha256sum >/dev/null; then ACTUAL=$(sha256sum "$TMP_DIR/$ASSET" | awk '{print $1}'); else ACTUAL=$(shasum -a 256 "$TMP_DIR/$ASSET" | awk '{print $1}'); fi
[ "$EXPECTED" = "$ACTUAL" ] || die 'checksum binary tidak cocok'
mkdir -p "$APP_DIR"
install -m 0755 "$TMP_DIR/$ASSET" "$APP_DIR/pan-vlessbank-client"
say "Installed: $APP_DIR/pan-vlessbank-client"
say "Release: $RELEASE"
say 'client.settings.json dipertahankan. Install cloudflared sendiri atau isi path di panel client.'
say "Jalankan: $APP_DIR/pan-vlessbank-client"
