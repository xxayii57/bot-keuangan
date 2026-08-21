#!/data/data/com.termux/files/usr/bin/bash

# =================================================================
# INTIMCLAW CLI AUTO-INSTALLER FOR ANDROID/TERMUX
# =================================================================
set -e

# Warna output terminal
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=====================================================${NC}"
echo -e "${GREEN}       MEMULAI INSTALASI AUTOMATIS INTIMCLAW CLI     ${NC}"
echo -e "${BLUE}=====================================================${NC}"

# 1. Update paket & Install dependencies dasar
echo -e "${YELLOW}[1/4] Menginstal dependensi Termux (curl, git, nodejs)...${NC}"
pkg update -y
pkg install -y curl git nodejs

# 2. Membuat direktori konfigurasi & workspace lokal di HP user baru
echo -e "${YELLOW}[2/4] Membuat direktori workspace ~/.intimclaw...${NC}"
mkdir -p ~/.intimclaw
mkdir -p ~/.intimclaw/skills
mkdir -p ~/.intimclaw/superintim
mkdir -p ~/.intimclaw/sessions

# 3. Mengunduh binary intimclaw untuk Android (ARM64)
echo -e "${YELLOW}[3/4] Mengunduh binary intimclaw ARM64...${NC}"
# Kita download langsung biner compiled dari server intimpos STB lo
curl -L -o $PREFIX/bin/intimclaw "http://intimpos.xxayii.my.id/downloads/intimclaw"
chmod +x $PREFIX/bin/intimclaw

# 4. Mengunduh berkas aset pendukung (Prompt & config template)
echo -e "${YELLOW}[4/4] Mendownload template konfigurasi & kepribadian AI...${NC}"
# Jika gagal konek ke web server, kita buat template dasar secara otomatis
cat << 'EOF' > ~/.intimclaw/config.toml
# config.toml — IntimClaw Default Configuration
[agent]
name = "intimclaw"
version = "0.1.0"
model_provider = ""
model = ""
persona = "superintim"

[providers.gemini]
type = "openai-compatible"
base_url = "https://generativelanguage.googleapis.com/v1beta"
api_key = ""
models = ["gemini-2.5-flash", "gemini-2.5-pro", "gemini-1.5-flash"]

[providers.anthropic]
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = ""
models = ["claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"]

[providers.openai]
type = "openai-compatible"
base_url = "https://api.openai.com/v1"
api_key = ""
models = ["gpt-4o", "gpt-4o-mini"]

[providers.groq]
type = "openai-compatible"
base_url = "https://api.groq.com/openai/v1"
api_key = ""
models = ["llama-3.3-70b-versatile", "mixtral-8x7b-32768"]

[skills]
enabled = true
directories = ["/data/data/com.termux/files/home/.intimclaw/skills"]

[memory]
backend = "sqlite"
semantic_search = true

[security]
risk_profile = "default"
sandbox = false
excluded_tools = ["rm", "mkfs", "dd", "shutdown"]
forbidden_paths = [".ssh", ".gnupg"]
EOF

# Membuat template SOUL.md kepribadian coding agent
cat << 'EOF' > ~/.intimclaw/superintim/SOUL.md
# SOUL OF INTIMCLAW AGENTIC CODER

Anda adalah **IntimClaw**, AI Agent coding mandiri yang tangguh, dikembangkan oleh xxayii (intim.my.id).

## Gaya Berkomunikasi & Nada:
- Gunakan bahasa Indonesia santai (casual lo/gue).
- Nada bicara santai, cerdas, efisien, dan berorientasi pada eksekusi.
- Berbicaralah seperti software engineer profesional yang asyik.

## Aturan Utama Eksekusi Coding:
1. Anda bukan asisten teks biasa. Anda adalah Agent Mandiri yang bisa memodifikasi kode.
2. Gunakan `list_dir` dan `file_read` untuk memahami codebase proyek sebelum melakukan perubahan.
3. Selalu gunakan `file_write` atau `file_edit` untuk membuat perubahan kode secara langsung dan rapi.
4. Setelah melakukan edit file, jalankan perintah pengujian menggunakan tool `exec` (misal `npm run build`, `python3 script.py`) untuk memverifikasi bahwa perubahan bekerja sempurna dan tidak menimbulkan error baru.
5. Jika ada error, analisis log-nya, perbaiki langsung, dan jangan menyerah sebelum target tercapai.
EOF

echo -e "${GREEN}=====================================================${NC}"
echo -e "${GREEN}       INTIMCLAW CLI BERHASIL DIINSTAL COK!          ${NC}"
echo -e "${GREEN}=====================================================${NC}"
echo -e "Ketik ${BLUE}intimclaw${NC} di terminal lo untuk memulai setup wizard!"
echo -e "Hubungi xxayii di ${YELLOW}intim.my.id${NC} jika butuh bantuan."
echo -e "${BLUE}=====================================================${NC}"
