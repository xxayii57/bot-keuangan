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
RELEASE_JSON=$(curl -fsSL "$LATEST_RELEASE_URL" 2>/dev/null || true)

if [ -z "$RELEASE_JSON" ]; then
    echo -e "${RED}[Error] Gagal menghubungi API GitHub. Periksa koneksi internet Anda.${NC}"
    exit 1
fi

# Parse tag_name dari JSON
TAG_NAME=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"(.*)".*/\1/')

if [ -z "$TAG_NAME" ]; then
    echo -e "${RED}[Error] Belum ada binary rilis IntimClaw di repositori ini.${NC}"
    echo -e "${RED}Hubungi pengembang untuk membuat rilis baru di GitHub.${NC}"
    exit 1
fi

echo -e "${GREEN}Menemukan Rilis Terbaru: $TAG_NAME${NC}"

# URL Aset Rilis GitHub
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/$BINARY_NAME"
CHECKSUM_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/SHA256SUMS"

TEMP_BIN="$(mktemp)"
TEMP_SUM="$(mktemp)"

# --- UNDUH CHECKSUM & BINARY ---
echo -e "${YELLOW}Mengunduh SHA256SUMS...${NC}"
if ! curl -fsSL -o "$TEMP_SUM" "$CHECKSUM_URL"; then
    echo -e "${RED}[Error] Gagal mengunduh berkas SHA256SUMS dari rilis.${NC}"
    rm -f "$TEMP_BIN" "$TEMP_SUM"
    exit 1
fi

echo -e "${YELLOW}Mengunduh $BINARY_NAME...${NC}"
if ! curl -fsSL -o "$TEMP_BIN" "$DOWNLOAD_URL"; then
    echo -e "${RED}[Error] Berkas biner '$BINARY_NAME' tidak ditemukan di rilis GitHub.${NC}"
    echo -e "${RED}Pastikan rilis untuk platform ini tersedia.${NC}"
    rm -f "$TEMP_BIN" "$TEMP_SUM"
    exit 1
fi

# --- VERIFIKASI CHECKSUM ---
echo -e "${YELLOW}Memverifikasi SHA256 Checksum...${NC}"
TARGET_CHECKSUM=$(grep "$BINARY_NAME" "$TEMP_SUM" | awk '{print $1}')

if [ -z "$TARGET_CHECKSUM" ]; then
    echo -e "${RED}[Error] Checksum untuk biner '$BINARY_NAME' tidak ditemukan di berkas SHA256SUMS.${NC}"
    echo -e "${RED}Berkas SHA256SUMS mungkin tidak berisi checksum untuk platform ini.${NC}"
    rm -f "$TEMP_BIN" "$TEMP_SUM"
    exit 1
fi

VERIFIED=false
if command -v sha256sum >/dev/null 2>&1; then
    if echo "$TARGET_CHECKSUM  $TEMP_BIN" | sha256sum -c - >/dev/null 2>&1; then
        VERIFIED=true
    fi
elif command -v shasum >/dev/null 2>&1; then
    if echo "$TARGET_CHECKSUM  $TEMP_BIN" | shasum -a 256 -c - >/dev/null 2>&1; then
        VERIFIED=true
    fi
else
    # Fallback ke python3
    if python3 -c "import hashlib, sys; f=open('$TEMP_BIN','rb').read(); h=hashlib.sha256(f).hexdigest(); sys.exit(0 if h == '$TARGET_CHECKSUM' else 1)" >/dev/null 2>&1; then
        VERIFIED=true
    fi
fi

if [ "$VERIFIED" = false ]; then
    echo -e "${RED}[Error] Checksum SHA256 tidak cocok! Berkas biner tidak aman atau rusak.${NC}"
    echo -e "${RED}Diharapkan: $TARGET_CHECKSUM${NC}"
    rm -f "$TEMP_BIN" "$TEMP_SUM"
    exit 1
fi

echo -e "${GREEN}✓ Verifikasi checksum berhasil!${NC}"

# --- INSTAL BINER ---
FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"
mv "$TEMP_BIN" "$FINAL_BINARY_PATH"
chmod +x "$FINAL_BINARY_PATH"
rm -f "$TEMP_SUM"

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
curl -fsSL -o "$WIZARD_DEST" "$REPO_RAW_URL/intimclaw_wizard.py"
chmod +x "$WIZARD_DEST"

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
        *) echo "Arsitektur CPU tidak didukung."; exit 1 ;;
    esac
    
    if [ "$IS_TERMUX" = true ]; then
        BINARY_NAME="intimclaw-android-arm64"
    else
        BINARY_NAME="intimclaw-linux-$ARCH"
        [ "$OS_NAME" = "darwin" ] && BINARY_NAME="intimclaw-darwin-$ARCH"
    fi
    
    LATEST_RELEASE_URL="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    RELEASE_JSON=$(curl -fsSL "$LATEST_RELEASE_URL" 2>/dev/null || true)
    TAG_NAME=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"(.*)".*/\1/')
    
    if [ -z "$TAG_NAME" ]; then
        echo "[Error] Gagal mendeteksi rilis terbaru dari GitHub."
        exit 1
    fi
    
    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/$BINARY_NAME"
    CHECKSUM_URL="https://github.com/$GITHUB_REPO/releases/download/$TAG_NAME/SHA256SUMS"
    
    TEMP_FILE="$(mktemp)"
    TEMP_SUM="$(mktemp)"
    
    echo "Mengunduh SHA256SUMS..."
    curl -fsSL -o "$TEMP_SUM" "$CHECKSUM_URL"
    
    echo "Mengunduh biner pembaruan..."
    curl -fsSL -o "$TEMP_FILE" "$DOWNLOAD_URL"
    
    # Verifikasi Checksum
    TARGET_CHECKSUM=$(grep "$BINARY_NAME" "$TEMP_SUM" | awk '{print $1}')
    VERIFIED=false
    if command -v sha256sum >/dev/null 2>&1; then
        echo "$TARGET_CHECKSUM  $TEMP_FILE" | sha256sum -c - >/dev/null 2>&1 && VERIFIED=true
    elif command -v shasum >/dev/null 2>&1; then
        echo "$TARGET_CHECKSUM  $TEMP_FILE" | shasum -a 256 -c - >/dev/null 2>&1 && VERIFIED=true
    else
        python3 -c "import hashlib, sys; f=open('$TEMP_FILE','rb').read(); h=hashlib.sha256(f).hexdigest(); sys.exit(0 if h == '$TARGET_CHECKSUM' else 1)" >/dev/null 2>&1 && VERIFIED=true
    fi
    
    if [ "$VERIFIED" = false ]; then
        echo "[Error] Verifikasi SHA256 gagal! Berkas pembaruan dibatalkan."
        rm -f "$TEMP_FILE" "$TEMP_SUM"
        exit 1
    fi
    
    # Ganti biner lama dengan aman
    mv "$TEMP_FILE" "$FINAL_BINARY_PATH"
    chmod +x "$FINAL_BINARY_PATH"
    rm -f "$TEMP_SUM"
    
    # Update Python Wizard
    curl -fsSL -o "$WIZARD_DEST" "$REPO_RAW_URL/intimclaw_wizard.py"
    chmod +x "$WIZARD_DEST"
    
    echo "====================================================="
    echo "✓ INTIMCLAW BERHASIL DIPERBARUI KE VERSI TERBARU!"
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
python3 "$WIZARD_DEST"

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
            # Cek idempotent untuk bash/zsh
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
