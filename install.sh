#!/usr/bin/env bash

# =================================================================
# INTIMCLAW UNIVERSAL BOOTSTRAP INSTALLER (PRODUCTION READY)
# =================================================================
set -e

# Warna output terminal
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=====================================================${NC}"
echo -e "${GREEN}       MEMULAI INSTALASI UNIVERSAL INTIMCLAW CLI     ${NC}"
echo -e "${BLUE}=====================================================${NC}"

# --- DETEKSI OS & ARSITEKTUR ---
OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_NAME="$(uname -m)"

# Normalisasi Nama Arsitektur
case "$ARCH_NAME" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}[Error] Arsitektur CPU '$ARCH_NAME' tidak didukung.${NC}"
        exit 1
        ;;
esac

# Deteksi Termux / Android
IS_TERMUX=false
if [ -n "$PREFIX" ] && [[ "$PREFIX" == *"com.termux"* ]]; then
    IS_TERMUX=true
fi

# Deteksi Windows WSL
IS_WSL=false
if grep -qE "(Microsoft|WSL)" /proc/version 2>/dev/null; then
    IS_WSL=true
fi

# Menentukan Target Biner & Direktori Instalasi
BINARY_NAME=""
INSTALL_DIR=""

if [ "$IS_TERMUX" = true ]; then
    echo -e "${YELLOW}Deteksi Platform: Android ARM64 (Termux)${NC}"
    BINARY_NAME="intimclaw-android-arm64"
    INSTALL_DIR="$PREFIX/bin"
else
    case "$OS_NAME" in
        linux)
            if [ "$IS_WSL" = true ]; then
                echo -e "${YELLOW}Deteksi Platform: Windows WSL ($ARCH)${NC}"
            else
                echo -e "${YELLOW}Deteksi Platform: Linux ($ARCH)${NC}"
            fi
            BINARY_NAME="intimclaw-linux-$ARCH"
            INSTALL_DIR="$HOME/.local/bin"
            ;;
        darwin)
            echo -e "${YELLOW}Deteksi Platform: macOS ($ARCH)${NC}"
            BINARY_NAME="intimclaw-darwin-$ARCH"
            INSTALL_DIR="$HOME/.local/bin"
            ;;
        *)
            echo -e "${RED}[Error] Sistem Operasi '$OS_NAME' tidak didukung oleh installer Linux/Unix.${NC}"
            echo -e "${YELLOW}Untuk Windows, silakan jalankan installer PowerShell (install.ps1).${NC}"
            exit 1
            ;;
    esac
fi

# Buat direktori instalasi jika belum ada
mkdir -p "$INSTALL_DIR"

# --- DETEKSI GITHUB RELEASE TERBARU ---
GITHUB_REPO="xxayii57/bot-keuangan"
LATEST_RELEASE_URL="https://api.github.com/repos/$GITHUB_REPO/releases/latest"

echo -e "${YELLOW}Menghubungi API GitHub untuk mencari rilis terbaru...${NC}"

HTTP_CODE=$(curl -s -o /tmp/intimclaw_release.json -w "%{http_code}" "$LATEST_RELEASE_URL" 2>/dev/null)
CURL_EXIT=$?

if [ $CURL_EXIT -ne 0 ]; then
    echo -e "${RED}[Error] Gagal menghubungi API GitHub (curl error).${NC}"
    echo -e "${RED}  → Periksa koneksi internet Anda.${NC}"
    echo -e "${YELLOW}  → Jika di jaringan korporat, pastikan proxy mengizinkan akses ke api.github.com.${NC}"
    rm -f /tmp/intimclaw_release.json
    exit 1
fi

case "$HTTP_CODE" in
    200)
        # Success - lanjut parse
        ;;
    404)
        echo -e "${RED}[Error] No published IntimClaw release found for this repository yet.${NC}"
        echo -e "${YELLOW}  → Repository: $GITHUB_REPO${NC}"
        echo -e "${YELLOW}  → Belum ada release yang dipublikasikan.${NC}"
        echo -e "${YELLOW}  → Hubungi pengembang untuk membuat rilis baru di GitHub:${NC}"
        echo -e "${YELLOW}    https://github.com/$GITHUB_REPO/releases/new${NC}"
        rm -f /tmp/intimclaw_release.json
        exit 1
        ;;
    403)
        echo -e "${RED}[Error] GitHub API rate limit exceeded (HTTP 403).${NC}"
        echo -e "${YELLOW}  → Tunggu beberapa menit atau login ke GitHub untuk meningkatkan limit.${NC}"
        echo -e "${YELLOW}  → Login: https://github.com/settings/tokens${NC}"
        rm -f /tmp/intimclaw_release.json
        exit 1
        ;;
    429)
        echo -e "${RED}[Error] GitHub API rate limit exceeded (HTTP 429).${NC}"
        echo -e "${YELLOW}  → Tunggu beberapa menit sebelum mencoba lagi.${NC}"
        rm -f /tmp/intimclaw_release.json
        exit 1
        ;;
    *)
        echo -e "${RED}[Error] GitHub API mengembalikan status HTTP $HTTP_CODE.${NC}"
        echo -e "${YELLOW}  → Response: $(cat /tmp/intimclaw_release.json | head -c 200)${NC}"
        rm -f /tmp/intimclaw_release.json
        exit 1
        ;;
esac

RELEASE_JSON=$(cat /tmp/intimclaw_release.json)
rm -f /tmp/intimclaw_release.json

# Parse tag_name dari JSON
TAG_NAME=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"(.*)".*/\1/')

if [ -z "$TAG_NAME" ]; then
    echo -e "${RED}[Error] No published IntimClaw release found for this repository yet.${NC}"
    echo -e "${YELLOW}  → Response dari GitHub tidak berisi tag release yang valid.${NC}"
    echo -e "${YELLOW}  → Repository: $GITHUB_REPO${NC}"
    echo -e "${YELLOW}  → Hubungi pengembang untuk membuat rilis baru di GitHub:${NC}"
    echo -e "${YELLOW}    https://github.com/$GITHUB_REPO/releases/new${NC}"
    exit 1
fi

echo -e "${GREEN}Menemukan Rilis Terbaru: $TAG_NAME${NC}"

# URL Aset Rilis GitHub
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/$BINARY_NAME"
CHECKSUM_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/SHA256SUMS"

TEMP_BIN="$(mktemp)"
TEMP_SUM="$(mktemp)"
cleanup() { rm -f "$TEMP_BIN" "$TEMP_SUM"; }
trap cleanup EXIT

# --- UNDUH CHECKSUM ---
echo -e "${YELLOW}Mengunduh SHA256SUMS...${NC}"
HTTP_CODE=$(curl -s -o "$TEMP_SUM" -w "%{http_code}" "$CHECKSUM_URL" 2>/dev/null)
CURL_EXIT=$?

if [ $CURL_EXIT -ne 0 ]; then
    echo -e "${RED}[Error] Gagal mengunduh SHA256SUMS (curl error).${NC}"
    echo -e "${YELLOW}  → Periksa koneksi internet Anda.${NC}"
    exit 1
fi

case "$HTTP_CODE" in
    200)
        # Success
        ;;
    404)
        echo -e "${RED}[Error] SHA256SUMS tidak ditemukan di rilis (HTTP 404).${NC}"
        echo -e "${YELLOW}  → Rilis $TAG_NAME tidak memiliki berkas SHA256SUMS.${NC}"
        echo -e "${YELLOW}  → Hubungi pengembang untuk memperbaiki rilis.$NC"
        exit 1
        ;;
    403)
        echo -e "${RED}[Error] GitHub rate limit saat mengunduh SHA256SUMS (HTTP 403).${NC}"
        echo -e "${YELLOW}  → Tunggu beberapa menit atau login ke GitHub.${NC}"
        exit 1
        ;;
    *)
        echo -e "${RED}[Error] Gagal mengunduh SHA256SUMS (HTTP $HTTP_CODE).${NC}"
        exit 1
        ;;
esac

# --- UNDUH BINARY ---
echo -e "${YELLOW}Mengunduh $BINARY_NAME...${NC}"
HTTP_CODE=$(curl -s -o "$TEMP_BIN" -w "%{http_code}" "$DOWNLOAD_URL" 2>/dev/null)
CURL_EXIT=$?

if [ $CURL_EXIT -ne 0 ]; then
    echo -e "${RED}[Error] Gagal mengunduh biner (curl error).${NC}"
    echo -e "${YELLOW}  → Periksa koneksi internet Anda.${NC}"
    exit 1
fi

case "$HTTP_CODE" in
    200)
        # Success
        ;;
    404)
        echo -e "${RED}[Error] Binary '$BINARY_NAME' tidak ditemukan di rilis $TAG_NAME (HTTP 404).${NC}"
        echo -e "${YELLOW}  → Rilis tersedia tapi tidak menyediakan biner untuk platform ini.${NC}"
        echo -e "${YELLOW}  → Platform: $(uname -s) $(uname -m)${NC}"
        echo -e "${YELLOW}  → Nama biner yang dicari: $BINARY_NAME${NC}"
        echo -e "${YELLOW}  → Hubungi pengembang untuk menambahkan build untuk platform ini.${NC}"
        exit 1
        ;;
    403)
        echo -e "${RED}[Error] GitHub rate limit saat mengunduh biner (HTTP 403).${NC}"
        echo -e "${YELLOW}  → Tunggu beberapa menit atau login ke GitHub.${NC}"
        exit 1
        ;;
    *)
        echo -e "${RED}[Error] Gagal mengunduh biner (HTTP $HTTP_CODE).${NC}"
        exit 1
        ;;
esac

# --- VERIFIKASI CHECKSUM ---
echo -e "${YELLOW}Memverifikasi SHA256 Checksum...${NC}"
TARGET_CHECKSUM=$(grep "$BINARY_NAME" "$TEMP_SUM" | awk '{print $1}')

if [ -z "$TARGET_CHECKSUM" ]; then
    echo -e "${RED}[Error] Checksum untuk biner '$BINARY_NAME' tidak ditemukan di SHA256SUMS.${NC}"
    echo -e "${YELLOW}  → Isi SHA256SUMS:${NC}"
    cat "$TEMP_SUM" | head -5
    echo -e "${YELLOW}  → Biner yang dicari: $BINARY_NAME${NC}"
    echo -e "${YELLOW}  → Kemungkinan rilis belum menyertakan checksum untuk platform ini.${NC}"
    exit 1
fi

ACTUAL_HASH=$(sha256sum "$TEMP_BIN" | awk '{print $1}')
if [ "$ACTUAL_HASH" != "$TARGET_CHECKSUM" ]; then
    echo -e "${RED}[Error] Checksum SHA256 TIDAK COCOK — file mungkin rusak atau terinfeksi.${NC}"
    echo -e "${RED}  → Diharapkan: $TARGET_CHECKSUM${NC}"
    echo -e "${RED}  → Ditemukan:  $ACTUAL_HASH${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Verifikasi checksum berhasil!${NC}"

# --- INSTAL BINER ---
FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"
mv "$TEMP_BIN" "$FINAL_BINARY_PATH"
chmod +x "$FINAL_BINARY_PATH"

# --- SETUP DIREKTORI DATA & CONFIG ---
CONFIG_DIR="$HOME/.intimclaw"
WIZARD_DEST="$CONFIG_DIR/intimclaw_wizard.py"

mkdir -p "$CONFIG_DIR"
mkdir -p "$CONFIG_DIR/skills"
mkdir -p "$CONFIG_DIR/persona"
mkdir -p "$CONFIG_DIR/sessions"
mkdir -p "$CONFIG_DIR/memory"

# Download Python Wizard File dari Github
REPO_RAW_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main"
HTTP_CODE=$(curl -s -o "$WIZARD_DEST" -w "%{http_code}" "$REPO_RAW_URL/intimclaw_wizard.py" 2>/dev/null)
if [ "$HTTP_CODE" = "200" ]; then
    chmod +x "$WIZARD_DEST"
else
    echo -e "${YELLOW}[Warning] Gagal mengunduh wizard (HTTP $HTTP_CODE). Setup wizard mungkin tidak tersedia.${NC}"
fi

# Buat wrapper bash/script di path utama
echo -e "${YELLOW}Mengonfigurasi tautan eksekusi...${NC}"
cat << 'EOF' > "$INSTALL_DIR/intimclaw"
#!/usr/bin/env bash

# Paths Configuration
IS_TERMUX=false
if [ -n "$PREFIX" ] && [[ "$PREFIX" == *"com.termux"* ]]; then
    IS_TERMUX=true
fi

if [ "$IS_TERMUX" = true ]; then
    INSTALL_DIR="$PREFIX/bin"
else
    INSTALL_DIR="$HOME/.local/bin"
fi

CONFIG_DIR="$HOME/.intimclaw"
WIZARD_DEST="$CONFIG_DIR/intimclaw_wizard.py"
FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"

GITHUB_REPO="xxayii57/bot-keuangan"
REPO_RAW_URL="https://raw.githubusercontent.com/xxayii57/bot-keuangan/main"

# Handle CLI Update Command
if [ "$1" = "update" ]; then
    echo "====================================================="
    echo "            MEMULAI PEMBARUAN INTIMCLAW              "
    echo "====================================================="

    # Deteksi OS & Arch
    OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH_NAME="$(uname -m)"

    case "$ARCH_NAME" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "[Error] Arsitektur CPU tidak didukung."; exit 1 ;;
    esac

    if [ "$IS_TERMUX" = true ]; then
        BINARY_NAME="intimclaw-android-arm64"
    else
        BINARY_NAME="intimclaw-linux-$ARCH"
        [ "$OS_NAME" = "darwin" ] && BINARY_NAME="intimclaw-darwin-$ARCH"
    fi

    LATEST_RELEASE_URL="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    HTTP_CODE=$(curl -s -o /tmp/intimclaw_release.json -w "%{http_code}" "$LATEST_RELEASE_URL" 2>/dev/null)
    CURL_EXIT=$?

    if [ $CURL_EXIT -ne 0 ]; then
        echo "[Error] Gagal menghubungi API GitHub (curl error)."
        exit 1
    fi

    if [ "$HTTP_CODE" = "404" ]; then
        echo "[Error] No published IntimClaw release found yet."
        rm -f /tmp/intimclaw_release.json
        exit 1
    fi

    if [ "$HTTP_CODE" != "200" ]; then
        echo "[Error] GitHub API returned HTTP $HTTP_CODE."
        rm -f /tmp/intimclaw_release.json
        exit 1
    fi

    RELEASE_JSON=$(cat /tmp/intimclaw_release.json)
    rm -f /tmp/intimclaw_release.json
    TAG_NAME=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"(.*)".*/\1/')

    if [ -z "$TAG_NAME" ]; then
        echo "[Error] No valid release tag found."
        exit 1
    fi

    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/$BINARY_NAME"
    CHECKSUM_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/SHA256SUMS"

    TEMP_FILE="$(mktemp)"
    TEMP_SUM="$(mktemp)"

    echo "Mengunduh SHA256SUMS..."
    HTTP_CODE=$(curl -s -o "$TEMP_SUM" -w "%{http_code}" "$CHECKSUM_URL" 2>/dev/null)
    if [ "$HTTP_CODE" != "200" ]; then
        echo "[Error] SHA256SUMS tidak tersedia (HTTP $HTTP_CODE)."
        rm -f "$TEMP_FILE" "$TEMP_SUM"
        exit 1
    fi

    echo "Mengunduh biner pembaruan..."
    HTTP_CODE=$(curl -s -o "$TEMP_FILE" -w "%{http_code}" "$DOWNLOAD_URL" 2>/dev/null)
    if [ "$HTTP_CODE" != "200" ]; then
        echo "[Error] Binary '$BINARY_NAME' tidak ditemukan di rilis $TAG_NAME (HTTP $HTTP_CODE)."
        rm -f "$TEMP_FILE" "$TEMP_SUM"
        exit 1
    fi

    # Verifikasi Checksum
    TARGET_CHECKSUM=$(grep "$BINARY_NAME" "$TEMP_SUM" | awk '{print $1}')
    if [ -z "$TARGET_CHECKSUM" ]; then
        echo "[Error] Checksum untuk '$BINARY_NAME' tidak ditemukan di SHA256SUMS."
        rm -f "$TEMP_FILE" "$TEMP_SUM"
        exit 1
    fi

    ACTUAL_HASH=$(sha256sum "$TEMP_FILE" | awk '{print $1}')
    if [ "$ACTUAL_HASH" != "$TARGET_CHECKSUM" ]; then
        echo "[Error] Checksum SHA256 tidak cocok! File mungkin rusak."
        echo "  Diharapkan: $TARGET_CHECKSUM"
        echo "  Ditemukan:  $ACTUAL_HASH"
        rm -f "$TEMP_FILE" "$TEMP_SUM"
        exit 1
    fi

    # Ganti biner lama dengan aman
    mv "$TEMP_FILE" "$FINAL_BINARY_PATH"
    chmod +x "$FINAL_BINARY_PATH"

    # Update Python Wizard
    curl -fsSL -o "$WIZARD_DEST" "$REPO_RAW_URL/intimclaw_wizard.py" 2>/dev/null
    chmod +x "$WIZARD_DEST"

    echo "====================================================="
    echo "✓ INTIMCLAW BERHASIL DIPERBARUI KE VERSI $TAG_NAME!"
    echo "====================================================="
    exit 0
fi

# Handle CLI Uninstall Command
if [ "$1" = "uninstall" ]; then
    echo "====================================================="
    echo "            UNINSTALL INTIMCLAW CLI                  "
    echo "====================================================="

    PURGE=false
    if [ "$2" = "--purge" ]; then
        PURGE=true
    fi

    echo "Menghapus biner aplikasi..."
    rm -f "$FINAL_BINARY_PATH"
    rm -f "$INSTALL_DIR/intimclaw"

    if [ "$PURGE" = true ]; then
        echo "Melakukan pembersihan total (--purge)..."
        rm -rf "$CONFIG_DIR"
        echo "✓ Semua data personal, memori, skills, dan konfigurasi telah dihapus."
    else
        echo "✓ Biner aplikasi berhasil dihapus."
        echo "ℹ️ Data personal Anda (konfigurasi, memori, skills) di $CONFIG_DIR tetap dipertahankan."
        echo "ℹ️ Untuk menghapus data personal secara permanen, jalankan kembali dengan opsi: --purge"
    fi
    echo "====================================================="
    exit 0
fi

# Check and run Setup Wizard if unconfigured
if [ -f "$WIZARD_DEST" ]; then
    python3 "$WIZARD_DEST"
fi

# Execute the real Go binary passing all arguments
exec "$FINAL_BINARY_PATH" "$@"
EOF
chmod +x "$INSTALL_DIR/intimclaw"

# --- SINKRONISASI PATH IDEMPOTENT ---
if [ "$IS_TERMUX" = false ]; then
    SHELL_CONFIGS=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.config/fish/config.fish")
    PATH_LINE='export PATH="$HOME/.local/bin:$PATH"'
    FISH_PATH_LINE='set -gx PATH $HOME/.local/bin $PATH'

    for config in "${SHELL_CONFIGS[@]}"; do
        if [ -f "$config" ]; then
            if [[ "$config" == *".fish" ]]; then
                if ! grep -q 'set -gx PATH $HOME/.local/bin' "$config"; then
                    echo "" >> "$config"
                    echo "$FISH_PATH_LINE" >> "$config"
                    echo -e "${GREEN}✓ Menambahkan PATH ke $config${NC}"
                fi
            else
                if ! grep -q 'export PATH="$HOME/.local/bin' "$config"; then
                    echo "" >> "$config"
                    echo "$PATH_LINE" >> "$config"
                    echo -e "${GREEN}✓ Menambahkan PATH ke $config${NC}"
                fi
            fi
        fi
    done
fi

echo -e "${GREEN}=====================================================${NC}"
echo -e "${GREEN}✓ INTIMCLAW CLI BERHASIL DIINSTAL!                  ${NC}"
echo -e "${GREEN}=====================================================${NC}"
echo -e "Silakan ketik ${BLUE}source ~/.bashrc${NC} (atau zshrc jika zsh) untuk mereload PATH."
echo -e "Ketik ${BLUE}intimclaw --version${NC} untuk menguji biner."
echo -e "${BLUE}=====================================================${NC}"
