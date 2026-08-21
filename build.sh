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
echo -e "${YELLOW}[1/6] Mengompilasi Linux ARM64 (aarch64)...${NC}"
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o releases/intimclaw-linux-arm64

echo -e "${YELLOW}[2/6] Mengompilasi Linux x86_64 (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o releases/intimclaw-linux-amd64

echo -e "${YELLOW}[3/6] Mengompilasi macOS ARM64 (Apple Silicon)...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o releases/intimclaw-darwin-arm64

echo -e "${YELLOW}[4/6] Mengompilasi macOS x86_64 (Intel)...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o releases/intimclaw-darwin-amd64

echo -e "${YELLOW}[5/6] Mengompilasi Windows x86_64...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o releases/intimclaw-windows-amd64.exe

echo -e "${YELLOW}[6/6] Mengompilasi Android/Termux ARM64...${NC}"
GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o releases/intimclaw-android-arm64

# --- GENERATE CHECKSUM SHA256 ---
echo -e "${YELLOW}Menghitung Checksum SHA256 untuk seluruh rilis...${NC}"
cd releases
sha256sum * > sha256.txt
cat sha256.txt
cd ..

echo -e "${GREEN}=====================================================${NC}"
echo -e "${GREEN}✓ SELURUH BINER BERHASIL DIBUILD DI DIRECTORY: releases/${NC}"
echo -e "${GREEN}=====================================================${NC}"
echo -e "Silakan upload isi folder 'releases/' ke server download Anda"
echo -e "atau buat GitHub Release dan lampirkan seluruh file di folder tersebut."
