# IntimClaw Installation & Distribution

Panduan instalasi universal untuk mendistribusikan **IntimClaw** lintas platform secara aman dan production-ready.

## 🚀 Instalasi Satu Baris (Linux, macOS, WSL, Termux)

Jalankan perintah ini langsung di terminal Anda untuk mengunduh, memverifikasi, dan mengonfigurasi IntimClaw secara otomatis:

```bash
curl -fsSL https://raw.githubusercontent.com/xxayii57/bot-keuangan/main/install.sh | bash
```

## 🪟 Instalasi Windows (PowerShell Native)

Buka PowerShell Anda (sebagai User Biasa atau Administrator) lalu jalankan perintah instalasi berikut:

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12; iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/xxayii57/bot-keuangan/main/install.ps1'))
```

---

## 📂 Lokasi Direktori Konfigurasi & Data

Untuk menjaga keamanan dan kebersihan sistem, IntimClaw memisahkan lokasi biner aplikasi dari data pribadi pengguna (XDG Standard):

### 1. Sistem Unix-Like (Linux, macOS, WSL, UserLAnd)
* **Executable Binary:** `~/.local/bin/intimclaw` (Ditambahkan ke `$PATH` otomatis secara idempotent).
* **Configuration:** `~/.config/intimclaw/config.toml` (Berisi pengaturan API key & provider).
* **Data Workspace:** `~/.local/share/intimclaw/` (Menyimpan `memory`, `skills`, `sessions`, `workspace`).
* **Cache & Logs:** `~/.cache/intimclaw/` (Tempat log dan cache).

### 2. Android (Termux)
* **Executable Binary:** `$PREFIX/bin/intimclaw` (Standard binary path Termux).
* **Workspace & Data:** `~/.intimclaw/` (Satu folder terisolasi untuk kemudahan akses di mobile).

### 3. Windows Native
* **Executable Binary:** `%USERPROFILE%\AppData\Local\Microsoft\WindowsApps\intimclaw.exe` atau `%USERPROFILE%\.local\bin\intimclaw.exe`.
* **Configuration:** `%USERPROFILE%\.intimclaw\config.toml` atau `%APPDATA%\intimclaw\config.toml`.
* **Data Workspace:** `%USERPROFILE%\.intimclaw\` atau `%LOCALAPPDATA%\intimclaw\`.

---

## 🛠️ Perawatan & Pemeliharaan (Update & Uninstall)

### Memperbarui IntimClaw
Cukup jalankan perintah pembaruan berikut untuk mencari rilis terbaru, memverifikasi checksum, dan mengganti biner tanpa menghapus data workspace lo:
```bash
intimclaw update
```

### Menghapus IntimClaw (Uninstall)
Untuk menghapus instalasi biner aplikasi tanpa menyentuh data personal (skills, memory, session):
```bash
# Menghapus biner saja
intimclaw uninstall --binary

# Menghapus biner dan seluruh data konfigurasi/memori (pembersihan total)
intimclaw uninstall --purge
```
