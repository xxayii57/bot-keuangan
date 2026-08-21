# =================================================================
# INTIMCLAW WINDOWS NATIVE POWERSHELL INSTALLER
# =================================================================

$ErrorActionPreference = 'Stop'

Write-Host "=====================================================" -ForegroundColor Blue
Write-Host "       MEMULAI INSTALASI NATIVE INTIMCLAW WINDOWS     " -ForegroundColor Green
Write-Host "=====================================================" -ForegroundColor Blue

# --- DETEKSI ARSITEKTUR ---
$Arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
$BinaryName = ""

switch ($Arch) {
    "X64" {
        $BinaryName = "intimclaw-windows-amd64.exe"
        Write-Host "Deteksi Platform: Windows x64" -ForegroundColor Yellow
    }
    "Arm64" {
        $BinaryName = "intimclaw-windows-arm64.exe"
        Write-Host "Deteksi Platform: Windows ARM64" -ForegroundColor Yellow
    }
    Default {
        Write-Error "Arsitektur CPU '$Arch' tidak didukung."
        Exit 1
    }
}

# --- DIREKTORI INSTALASI USER ---
$UserLocalBin = Join-Path $HOME ".local\bin"
if (-not (Test-Path $UserLocalBin)) {
    New-Item -ItemType Directory -Force -Path $UserLocalBin | Out-Null
}

$ConfigDir = Join-Path $HOME ".intimclaw"
if (-not (Test-Path $ConfigDir)) {
    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "skills") | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "persona") | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "sessions") | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "memory") | Out-Null
}

# --- DETEKSI GITHUB RELEASE TERBARU ---
$GITHUB_REPO = "xxayii57/bot-keuangan"
$LatestReleaseUrl = "https://api.github.com/repos/$GITHUB_REPO/releases/latest"

Write-Host "Menghubungi API GitHub untuk mencari rilis terbaru..." -ForegroundColor Yellow
try {
    $ReleaseJson = Invoke-RestMethod -Uri $LatestReleaseUrl -UseBasicParsing
    $TagName = $ReleaseJson.tag_name
} catch {
    Write-Error "Gagal menghubungi API GitHub. Periksa koneksi internet Anda."
    Exit 1
}

if ([string]::IsNullOrEmpty($TagName)) {
    Write-Error "Belum ada binary rilis IntimClaw di repositori ini."
    Exit 1
}

Write-Host "Menemukan Rilis Terbaru: $TagName" -ForegroundColor Green

# --- DOWNLOAD & VERIFIKASI ---
$DownloadUrl = "https://github.com/$GITHUB_REPO/releases/download/$TagName/$BinaryName"
$ChecksumUrl = "https://github.com/$GITHUB_REPO/releases/download/$TagName/SHA256SUMS"

$TempExe = Join-Path $env:TEMP "intimclaw_download.exe"
$TempSum = Join-Path $env:TEMP "intimclaw_sha256.txt"

# Unduh checksum
Write-Host "Mengunduh SHA256SUMS..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $TempSum -UseBasicParsing
} catch {
    Write-Error "Gagal mengunduh berkas SHA256SUMS dari rilis."
    Remove-Item $TempExe, $TempSum -Force -ErrorAction SilentlyContinue
    Exit 1
}

# Temukan target checksum
$TargetChecksum = ""
$ChecksumContent = Get-Content $TempSum -Raw
$Lines = $ChecksumContent -split "`n"
foreach ($Line in $Lines) {
    if ($Line -match $BinaryName) {
        $TargetChecksum = ($Line -split "\s+")[0].Trim()
        break
    }
}

if ([string]::IsNullOrEmpty($TargetChecksum)) {
    Write-Error "Checksum untuk biner '$BinaryName' tidak ditemukan di berkas SHA256SUMS."
    Remove-Item $TempExe, $TempSum -Force -ErrorAction SilentlyContinue
    Exit 1
}

# Unduh biner
Write-Host "Mengunduh biner $BinaryName..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempExe -UseBasicParsing
} catch {
    Write-Error "Gagal mengunduh biner '$BinaryName' dari rilis GitHub."
    Remove-Item $TempExe, $TempSum -Force -ErrorAction SilentlyContinue
    Exit 1
}

# Verifikasi Checksum SHA256
Write-Host "Memverifikasi integritas berkas (SHA256 Checksum)..." -ForegroundColor Yellow
$FileHash = (Get-FileHash -Path $TempExe -Algorithm SHA256).Hash.ToLower()

if ($FileHash -ne $TargetChecksum.ToLower()) {
    Write-Error "Checksum SHA256 tidak cocok! Unduhan rusak atau berkas tidak aman."
    Write-Host "Diharapkan: $TargetChecksum" -ForegroundColor Red
    Write-Host "Ditemukan:  $FileHash" -ForegroundColor Red
    Remove-Item $TempExe, $TempSum -Force -ErrorAction SilentlyContinue
    Exit 1
}

Write-Host "✓ Verifikasi checksum berhasil. Berkas biner aman!" -ForegroundColor Green

# --- INSTALASI BINER ---
$FinalBinaryPath = Join-Path $UserLocalBin "intimclaw.exe"
Move-Item -Path $TempExe -Destination $FinalBinaryPath -Force
Remove-Item $TempSum -Force -ErrorAction SilentlyContinue

# --- PENGATURAN USER PATH ENVIRONMENT (IDEMPOTENT) ---
Write-Host "Mengonfigurasi System PATH..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$UserLocalBin*") {
    $NewUserPath = "$UserPath;$UserLocalBin"
    [Environment]::SetEnvironmentVariable("PATH", $NewUserPath, "User")
    $env:Path = "$env:Path;$UserLocalBin"
    Write-Host "✓ Direktori $UserLocalBin ditambahkan ke PATH." -ForegroundColor Green
} else {
    Write-Host "✓ Direktori $UserLocalBin sudah terdaftar di PATH." -ForegroundColor Green
}

# --- MEMBUAT CONFIG & SOUL FILE ---
$ConfigPath = Join-Path $ConfigDir "config.toml"
if (-not (Test-Path $ConfigPath)) {
    $ConfigContent = @"
# config.toml — IntimClaw Default Configuration
[agent]
name = "intimclaw"
version = "0.1.0"
provider = ""
model = ""
persona = "default"

[[providers]]
name = "openai"
type = "openai-compatible"
base_url = "https://api.openai.com/v1"
api_key = ""
models = ["gpt-4o-mini", "gpt-4o"]

[[providers]]
name = "anthropic"
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = ""
models = ["claude-3-5-sonnet-latest"]

[[providers]]
name = "ollama"
type = "openai-compatible"
base_url = "http://localhost:11434/v1"
api_key = ""
models = ["llama3.2", "codellama"]

[[providers]]
name = "groq"
type = "openai-compatible"
base_url = "https://api.groq.com/openai/v1"
api_key = ""
models = ["llama-3.3-70b-versatile"]

[channels.telegram]
enabled = false
bot_token = ""
owner_id = 0
mention_only = false

[channels.discord]
enabled = false
bot_token = ""

[channels.webchat]
enabled = true
port = 18080
host = "127.0.0.1"

[mcp]
enabled = false

[skills]
enabled = true
directories = ["$($ConfigDir.Replace('\', '\\'))\\skills"]

[memory]
backend = "sqlite"
semantic_search = true
decay_days = 30

[security]
risk_profile = "default"
sandbox = false
excluded_tools = ["rm", "mkfs", "dd", "shutdown", "poweroff"]
forbidden_paths = [".ssh", ".gnupg", ".aws"]

[webui]
enabled = true
port = 18080
host = "127.0.0.1"
theme = "intimclaw"

[cron]
enabled = true
max_jobs = 50
"@
    Set-Content -Path $ConfigPath -Value $ConfigContent
}

# Buat SOUL.md jika belum ada
$SoulPath = Join-Path $ConfigDir "persona\SOUL.md"
if (-not (Test-Path $SoulPath)) {
    $SoulContent = @"
# SOUL OF INTIMCLAW AGENTIC CODER

Anda adalah **IntimClaw**, AI Agent coding mandiri yang tangguh dan efisien.

## Gaya Berkomunikasi & Nada:
- Gunakan bahasa Indonesia santai (casual lo/gue).
- Nada bicara santai, cerdas, efisien, dan berorientasi pada eksekusi.
- Berbicaralah seperti software engineer profesional yang asyik.

## Aturan Utama Eksekusi Coding:
1. Anda bukan asisten teks biasa. Anda adalah Agent Mandiri yang bisa memodifikasi kode.
2. Gunakan ``list_dir`` dan ``file_read`` untuk memahami codebase proyek sebelum melakukan perubahan.
3. Selalu gunakan ``file_write`` atau ``file_edit`` untuk membuat perubahan kode secara langsung dan rapi.
4. Setelah melakukan edit file, jalankan perintah pengujian menggunakan tool ``exec`` (misal ``npm run build``, ``python3 script.py``) untuk memverifikasi bahwa perubahan bekerja sempurna dan tidak menimbulkan error baru.
5. Jika ada error, analisis log-nya, perbaiki langsung, dan jangan menyerah sebelum target tercapai.
"@
    Set-Content -Path $SoulPath -Value $SoulContent
}

Write-Host "=====================================================" -ForegroundColor Green
Write-Host "       INTIMCLAW CLI BERHASIL DIINSTAL COK!          " -ForegroundColor Green
Write-Host "=====================================================" -ForegroundColor Blue
Write-Host "Buka PowerShell baru dan ketik 'intimclaw' untuk memulai setup!" -ForegroundColor Yellow
