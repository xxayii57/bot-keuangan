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
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "superintim") | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $ConfigDir "sessions") | Out-Null
}

# --- DOWNLOAD & VERIFIKASI ---
$RepoUrl = "https://raw.githubusercontent.com/xxayii57/bot-keuangan/main"
$BinaryUrl = "$RepoUrl/releases/$BinaryName"
$ChecksumUrl = "$RepoUrl/docs/release/sha256.txt"

$TempExe = New-TemporaryFile
$TempExePath = $TempExe.FullName
Rename-Item -Path $TempExePath -NewName "$($TempExe.Name).exe" -Force
$TempExePath = "$TempExePath.exe"

Write-Host "Mengunduh berkas checksum..." -ForegroundColor Yellow
$WebClient = New-Object System.Net.WebClient
$ChecksumsData = $WebClient.DownloadString($ChecksumUrl)

# Temukan target checksum
$TargetChecksum = ""
$Lines = $ChecksumsData -split "`n"
foreach ($Line in $Lines) {
    if ($Line -match $BinaryName) {
        $TargetChecksum = ($Line -split "\s+")[0].Trim()
        break
    }
}

if ([string]::IsNullOrEmpty($TargetChecksum)) {
    Write-Error "Checksum untuk biner '$BinaryName' tidak ditemukan di database rilis."
    Exit 1
}

Write-Host "Mengunduh biner $BinaryName dari sumber aman..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $BinaryUrl -OutFile $TempExePath -UseBasicParsing
} catch {
    Write-Host "Mencoba fallback lokal dari repo assets..." -ForegroundColor Yellow
    $FallbackUrl = "$RepoUrl/intimclaw-android-project/app/src/main/assets/intimclaw-android"
    Invoke-WebRequest -Uri $FallbackUrl -OutFile $TempExePath -UseBasicParsing
}

# Verifikasi Checksum SHA256
Write-Host "Memverifikasi integritas berkas (SHA256 Checksum)..." -ForegroundColor Yellow
$FileHash = (Get-FileHash -Path $TempExePath -Algorithm SHA256).Hash.ToLower()

if ($FileHash -ne $TargetChecksum.ToLower()) {
    Write-Error "Checksum SHA256 tidak cocok! Unduhan rusak atau berkas tidak aman dideteksi."
    Remove-Item $TempExePath -Force
    Exit 1
}

Write-Host "✓ Verifikasi checksum berhasil. Berkas biner aman!" -ForegroundColor Green

# --- INSTALASI BINER ---
$FinalBinaryPath = Join-Path $UserLocalBin "intimclaw.exe"
Move-Item -Path $TempExePath -Destination $FinalBinaryPath -Force

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
model_provider = ""
model = ""
persona = "superintim"

[providers.gemini]
type = "openai-compatible"
base_url = "https://generativelanguage.googleapis.com/v1beta"
api_key = ""
models = ["gemini-2.5-flash", "gemini-2.5-pro"]

[providers.anthropic]
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = ""
models = ["claude-3-5-sonnet-latest"]

[providers.openai]
type = "openai-compatible"
base_url = "https://api.openai.com/v1"
api_key = ""
models = ["gpt-4o-mini"]

[providers.groq]
type = "openai-compatible"
base_url = "https://api.groq.com/openai/v1"
api_key = ""
models = ["llama-3.3-70b-versatile"]

[skills]
enabled = true
directories = ["$($ConfigDir.Replace('\', '\\'))\\skills"]

[memory]
backend = "sqlite"
semantic_search = true

[security]
risk_profile = "default"
sandbox = false
excluded_tools = ["rm", "mkfs", "dd", "shutdown"]
forbidden_paths = [".ssh", ".gnupg"]
"@
    Set-Content -Path $ConfigPath -Value $ConfigContent
}

# Buat SOUL.md jika belum ada
$SoulPath = Join-Path $ConfigDir "superintim\SOUL.md"
if (-not (Test-Path $SoulPath)) {
    $SoulContent = @"
# SOUL OF INTIMCLAW AGENTIC CODER

Anda adalah **IntimClaw**, AI Agent coding mandiri yang tangguh, dikembangkan oleh xxayii (intim.my.id).

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
