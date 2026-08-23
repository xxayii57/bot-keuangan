#!/usr/bin/env bash
# =================================================================
# INTIMCLAW INSTALLER (Linux / macOS)
# Installs the latest stable release from GitHub.
#
#   curl -fsSL https://raw.githubusercontent.com/xxayii57/bot-keuangan/main/install.sh | bash
# =================================================================
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[Error]${NC} $1" >&2; exit 1; }

GITHUB_REPO="xxayii57/bot-keuangan"
DOWNLOAD_BASE="https://github.com/${GITHUB_REPO}/releases/latest/download"

echo -e "${BLUE}=====================================================${NC}"
echo -e "${GREEN}       INSTALASI INTIMCLAW (Linux/macOS)             ${NC}"
echo -e "${BLUE}=====================================================${NC}"

command -v curl >/dev/null 2>&1 || fail "curl tidak ditemukan. Install curl terlebih dahulu."
command -v tar >/dev/null 2>&1 || fail "tar tidak ditemukan. Install tar terlebih dahulu."

# --- OS & ARCH DETECTION ---
OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_NAME="$(uname -m)"

case "$ARCH_NAME" in
    x86_64|amd64)   ARCH="x86_64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    i386|i686)      ARCH="i386" ;;
    riscv64)        ARCH="riscv64" ;;
    loongarch64)    ARCH="loong64" ;;
    s390x)          ARCH="s390x" ;;
    mips64el*)      fail "Arsitektur '$ARCH_NAME' tidak didukung (yang didukung: mipsle)."
                    exit 1 ;;
    armv6l)         ARCH="armv6" ;;
    armv7l|armv8l)  ARCH="armv7" ;;
    arm64*32*)      fail "Arsitektur ILP32 tidak didukung."; exit 1 ;;
    *)              fail "Arsitektur CPU '$ARCH_NAME' tidak didukung." ;;
esac

IS_TERMUX=false
if [ -n "${PREFIX:-}" ] && [[ "${PREFIX}" == *"com.termux"* ]]; then
    IS_TERMUX=true
fi

case "$OS_NAME" in
    linux)
        if [ "$IS_TERMUX" = true ]; then
            warn "Deteksi Platform: Termux/Android."
            warn "Binary Linux di bawah ini berjalan di dalam proot-distro, BUKAN Termux native."
            warn "Jika ingin native Android, unduh bundle: https://github.com/${GITHUB_REPO}/releases"
        fi
        OS_TITLE="Linux"
        ;;
    darwin)
        OS_TITLE="Darwin"
        ;;
    *)
        fail "Sistem Operasi '$OS_NAME' tidak didukung installer ini. Unduh manual dari https://github.com/${GITHUB_REPO}/releases"
        ;;
esac

ASSET="intimclaw_${OS_TITLE}_${ARCH}.tar.gz"
CHECKSUMS="checksums.txt"

# --- INSTALL DIRECTORY ---
if [ "$IS_TERMUX" = true ]; then
    INSTALL_DIR="${PREFIX}/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
fi
mkdir -p "$INSTALL_DIR"

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

# --- DOWNLOAD ---
info "Mengunduh ${ASSET}..."
HTTP_CODE=$(curl -fsSL -o "${TMP_DIR}/${ASSET}" -w "%{http_code}" "${DOWNLOAD_BASE}/${ASSET}") \
    || fail "Gagal mengunduh ${ASSET} (HTTP ${HTTP_CODE:-000}). Kemungkinan belum ada release stabil. Cek: https://github.com/${GITHUB_REPO}/releases"

info "Memverifikasi checksum..."
if curl -fsSL -o "${TMP_DIR}/${CHECKSUMS}" "${DOWNLOAD_BASE}/${CHECKSUMS}" 2>/dev/null; then
    EXPECTED=$(grep " ${ASSET}\$" "${TMP_DIR}/${CHECKSUMS}" | awk '{print $1}')
    if [ -n "$EXPECTED" ]; then
        ACTUAL=$(sha256sum "${TMP_DIR}/${ASSET}" | awk '{print $1}')
        if [ "$EXPECTED" != "$ACTUAL" ]; then
            fail "Checksum mismatch! Expected ${EXPECTED}, got ${ACTUAL}."
        fi
        ok "Checksum valid."
    else
        warn "Aset tidak ditemukan di checksums.txt — lewati verifikasi."
    fi
else
    warn "checksums.txt tidak tersedia — lewati verifikasi."
fi

# --- EXTRACT & INSTALL ---
info "Ekstrak dan instal ke ${INSTALL_DIR}..."
tar -xzf "${TMP_DIR}/${ASSET}" -C "$TMP_DIR"

for BIN in intimclaw intimclaw-launcher; do
    if [ -f "${TMP_DIR}/${BIN}" ]; then
        install -m 0755 "${TMP_DIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
        ok "${BIN} terpasang."
    fi
done

[ -x "${INSTALL_DIR}/intimclaw" ] || fail "Binary 'intimclaw' tidak ditemukan di dalam arsip."

echo ""
echo -e "${BLUE}=====================================================${NC}"
ok  "Instalasi selesai!"
echo -e "${BLUE}=====================================================${NC}"
echo ""

case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        warn "${INSTALL_DIR} belum ada di PATH Anda."
        echo "  Tambahkan baris ini ke ~/.bashrc atau ~/.zshrc:"
        echo ""
        echo "    export PATH=\"\${HOME}/.local/bin:\$PATH\""
        echo ""
        ;;
esac

info "Langkah selanjutnya:"
echo "  1. intimclaw onboard    # inisialisasi konfigurasi & workspace"
echo "  2. intimclaw model      # atur default model LLM + API key"
echo "  3. intimclaw gateway    # jalankan gateway"
echo "  4. intimclaw-launcher   # (opsional) Web console di http://localhost:18800"
