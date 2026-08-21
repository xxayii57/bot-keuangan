#!/usr/bin/env bash
# =================================================================
# INTIMCLAW MULTI-PLATFORM LOCAL BUILD SYSTEM
# =================================================================
set -e

# Warna Output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${BLUE}=====================================================${NC}"
echo -e "${GREEN}      MEMULAI PROSES BUILD MULTI-PLATFORM INTIMCLAW  ${NC}"
echo -e "${BLUE}=====================================================${NC}"

# Buat berkas skills.enc kosong jika belum ada agar go:embed tidak error
touch internal/agent/skills.enc

# Bersihkan direktori dist rilis sebelumnya
mkdir -p releases
rm -f releases/*

# --- KOMPILASI SILANG (CROSS COMPILATION) ---
TARGETS=(
    "linux:arm64:intimclaw-linux-arm64"
    "linux:amd64:intimclaw-linux-amd64"
    "darwin:arm64:intimclaw-darwin-arm64"
    "darwin:amd64:intimclaw-darwin-amd64"
    "windows:amd64:intimclaw-windows-amd64.exe"
    "windows:arm64:intimclaw-windows-arm64.exe"
    "android:arm64:intimclaw-android-arm64"
)

TOTAL=${#TARGETS[@]}
COUNT=0

for target in "${TARGETS[@]}"; do
    IFS=':' read -r GOOS GOARCH OUTPUT <<< "$target"
    COUNT=$((COUNT + 1))
    echo -e "${YELLOW}[$COUNT/$TOTAL] Mengompilasi $GOOS/$GOARCH → $OUTPUT...${NC}"
    GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "releases/$OUTPUT"
done

# --- GENERATE CHECKSUM SHA256 ---
echo -e "${YELLOW}Menghitung Checksum SHA256 untuk seluruh rilis...${NC}"
cd releases
sha256sum intimclaw-* > SHA256SUMS
cat SHA256SUMS
cd ..

echo -e "${GREEN}=====================================================${NC}"
echo -e "${GREEN}✓ SELURUH BINER BERHASIL DIBUILD DI DIRECTORY: releases/${NC}"
echo -e "${GREEN}=====================================================${NC}"
echo -e "Silakan upload isi folder 'releases/' ke server download Anda"
echo -e "atau buat GitHub Release dan lampirkan seluruh file di folder tersebut."
