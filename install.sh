#!/usr/bin/env bash

# =================================================================
# INTIMCLAW UNIVERSAL BOOTSTRAP INSTALLER
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

# Menentukan Target Biner
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

# --- UNDUH & VERIFIKASI CHECKSUM ---
LATEST_RELEASE_URL="https://api.github.com/repos/xxayii57/bot-keuangan/releases/latest"
TAG_NAME=$(curl -fsSL "$LATEST_RELEASE_URL" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"(.*)".*/\1/')

if [ -z "$TAG_NAME" ]; then
    echo -e "${YELLOW}GitHub Release tidak dapat diakses (Private Repo / Belum ada Release). Menggunakan fallback server intimpos...${NC}"
    DOWNLOAD_URL="http://intimpos.xxayii.my.id/downloads/$BINARY_NAME"
    CHECKSUM_URL="http://intimpos.xxayii.my.id/downloads/sha256.txt"
else
    echo -e "${GREEN}Menemukan Rilis Terbaru di GitHub: $TAG_NAME${NC}"
    DOWNLOAD_URL="https://github.com/xxayii57/bot-keuangan/releases/download/$TAG_NAME/$BINARY_NAME"
    CHECKSUM_URL="https://github.com/xxayii57/bot-keuangan/releases/download/$TAG_NAME/sha256.txt"
fi

TEMP_FILE="$(mktemp)"

echo -e "${YELLOW}Mengunduh berkas checksum...${NC}"
if ! CHECKSUMS_DATA=$(curl -fsSL "$CHECKSUM_URL"); then
    echo -e "${RED}[Error] Gagal mengunduh berkas checksum dari $CHECKSUM_URL.${NC}"
    rm -f "$TEMP_FILE"
    exit 1
fi
TARGET_CHECKSUM=$(echo "$CHECKSUMS_DATA" | grep "$BINARY_NAME" | awk '{print $1}')

if [ -z "$TARGET_CHECKSUM" ]; then
    echo -e "${RED}[Error] Checksum untuk biner '$BINARY_NAME' tidak ditemukan di database rilis.${NC}"
    rm -f "$TEMP_FILE"
    exit 1
fi

echo -e "${YELLOW}Mengunduh biner $BINARY_NAME dari sumber aman...${NC}"
if ! curl -fsSL -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
    echo -e "${YELLOW}Mencoba fallback lokal dari repo assets...${NC}"
    REPO_RAW_URL="https://raw.githubusercontent.com/xxayii57/bot-keuangan/main"
    if ! curl -fsSL -o "$TEMP_FILE" "$REPO_RAW_URL/intimclaw-android-project/app/src/main/assets/intimclaw-android"; then
        echo -e "${RED}[Error] Gagal mengunduh biner dari semua sumber rilis.${NC}"
        rm -f "$TEMP_FILE"
        exit 1
    fi
fi

# Verifikasi Checksum SHA256
echo -e "${YELLOW}Memverifikasi integritas berkas (SHA256 Checksum)...${NC}"
VERIFIED=false

if command -v sha256sum >/dev/null 2>&1; then
    if echo "$TARGET_CHECKSUM  $TEMP_FILE" | sha256sum -c - >/dev/null 2>&1; then
        VERIFIED=true
    fi
elif command -v shasum >/dev/null 2>&1; then
    if echo "$TARGET_CHECKSUM  $TEMP_FILE" | shasum -a 256 -c - >/dev/null 2>&1; then
        VERIFIED=true
    fi
else
    # Fallback to python3 hash verification
    if python3 -c "import hashlib, sys; f=open('$TEMP_FILE','rb').read(); h=hashlib.sha256(f).hexdigest(); sys.exit(0 if h == '$TARGET_CHECKSUM' else 1)" >/dev/null 2>&1; then
        VERIFIED=true
    fi
fi

if [ "$VERIFIED" = false ]; then
    echo -e "${RED}[Error] Checksum SHA256 tidak cocok! Pembajakan berkas atau unduhan rusak dideteksi.${NC}"
    echo -e "${RED}Instalasi dibatalkan demi keamanan.${NC}"
    rm -f "$TEMP_FILE"
    exit 1
fi

echo -e "${GREEN}✓ Verifikasi checksum berhasil. Berkas biner aman!${NC}"

# --- INSTAL BINER ---
FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"
mv "$TEMP_FILE" "$FINAL_BINARY_PATH"
chmod +x "$FINAL_BINARY_PATH"

# --- SETUP DIREKTORI DATA & CONFIG ---
CONFIG_DIR=""
WIZARD_DEST=""

if [ "$IS_TERMUX" = true ]; then
    CONFIG_DIR="$HOME/.intimclaw"
    WIZARD_DEST="$HOME/intimclaw_wizard.py"
else
    CONFIG_DIR="$HOME/.config/intimclaw"
    mkdir -p "$HOME/.local/share/intimclaw/skills"
    mkdir -p "$HOME/.local/share/intimclaw/superintim"
    mkdir -p "$HOME/.local/share/intimclaw/sessions"
    WIZARD_DEST="$CONFIG_DIR/intimclaw_wizard.py"
fi

mkdir -p "$CONFIG_DIR"

# Download Python Wizard File
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
    CONFIG_DIR="$HOME/.intimclaw"
    WIZARD_DEST="$HOME/intimclaw_wizard.py"
    FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"
else
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/intimclaw"
    WIZARD_DEST="$CONFIG_DIR/intimclaw_wizard.py"
    FINAL_BINARY_PATH="$INSTALL_DIR/intimclaw-bin"
fi

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
    
    DOWNLOAD_URL="$REPO_RAW_URL/releases/$BINARY_NAME"
    CHECKSUM_URL="$REPO_RAW_URL/docs/release/sha256.txt"
    TEMP_FILE="$(mktemp)"
    
    echo "Mengecek rilis terbaru & checksum..."
    CHECKSUMS_DATA=$(curl -fsSL "$CHECKSUM_URL")
    TARGET_CHECKSUM=$(echo "$CHECKSUMS_DATA" | grep "$BINARY_NAME" | awk '{print $1}')
    
    if [ -z "$TARGET_CHECKSUM" ]; then
        echo "[Error] Checksum untuk $BINARY_NAME tidak ditemukan."
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    echo "Mengunduh biner pembaruan..."
    if ! curl -fsSL -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
        echo "Mencoba fallback lokal dari repo assets..."
        curl -fsSL -o "$TEMP_FILE" "$REPO_RAW_URL/intimclaw-android-project/app/src/main/assets/intimclaw-android"
    fi
    
    # Verifikasi Checksum
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
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    # Ganti biner lama dengan aman
    mv "$TEMP_FILE" "$FINAL_BINARY_PATH"
    chmod +x "$FINAL_BINARY_PATH"
    
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
        [ "$IS_TERMUX" = false ] && rm -rf "$HOME/.local/share/intimclaw"
        [ "$IS_TERMUX" = false ] && rm -rf "$HOME/.cache/intimclaw"
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
echo -e "${GREEN}       INTIMCLAW CLI BERHASIL DIINSTAL COK!          ${NC}"
echo -e "${GREEN}=====================================================${NC}"
echo -e "Silakan ketik ${BLUE}source ~/.bashrc${NC} (atau zshrc jika zsh) untuk mereload PATH."
echo -e "Ketik ${BLUE}intimclaw${NC} di terminal lo untuk memulai setup wizard!"
echo -e "${BLUE}=====================================================${NC}"
