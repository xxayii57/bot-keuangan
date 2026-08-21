# Project & VPS Memory

> ⚠️ **SECURITY NOTICE (Phase 1A redaction — 2026-08-21)**
> Kredensial asli telah DIHAPUS dari file ini karena sempat terekspos di
> repository publik. Semua kredensial yang tercantum sebelumnya WAJIB dirotasi.
> Simpan kredensial di password manager / file lokal yang tidak di-commit.

## Infrastruktur

- **Source VPS:** root@47.236.50.11 (Location: `/root`, `/var/www`)
  - Passcode: `[REDACTED — ROTATE IMMEDIATELY]`
- **Target STB (CasaOS):** root@ssh.intim.my.id (via Cloudflare Tunnel)
  - STB Passcode: `[REDACTED — ROTATE IMMEDIATELY]`
- **Local STB IP:** 192.168.1.81
- **Project Backup:** Located at `/DATA/Downloads/gz/all_projects.tar.gz` on STB.
- **STB Remote Access:** Configuration stored in `STB_ACCESS.md` (Cloudflare Tunnel + SSH).
- **TV Xiaomi IP:** 192.168.1.68 (Scripts for interaction in `/root/send_text_tv.py`).

*Note: Credentials updated for convenience.* ← PRAKTIK INI TIDAK AMAN, JANGAN ULANGI.

- **Current Identity:** IRONCLAW 🔥 (Superagent V7 / v7.0 — OPENCLAW EDITION)

### Project Undangan Digital (intim.my.id)
- **Domain Utama:** `intim.my.id` & `www.intim.my.id` (pointing to STB Port 8088 Nginx)
- **Frontend Path:** `/var/www/html` on STB (Nginx serves themes & index)
- **Backend API Path:** `/var/www/undangan-api` (Port 8089 via Node.js, systemd `undangan-api.service`)
- **Nginx Config Path:** `/etc/nginx/sites-available/undangan`
- **Dynamic Routing:** Redirects `/v[1-7]/.+` to appropriate theme templates (e.g. `/themes/v5/index.html?id=...&to=...`)
- **Project History:** Transferred from VPS / `undangan.xxayii.my.id` to STB local hosting under `intim.my.id`.

### Automated Spy Bot Setup (STB)
- **Target iPhone IP:** 192.168.1.78
- **Telegram Bot:** `[REDACTED — REVOKE VIA @BotFather]` (token bot lama, sudah terekspos publik)
- **Chat ID:** 580132327
- **Script Location:** /root/spy_bot.py
- **Log Location:** /DATA/spy_log.txt
- **Status:** Active via systemd (spy-bot.service)

## Credential Rotation Checklist (Phase 1A)

| # | Credential | Lokasi lama | Status |
|---|---|---|---|
| R1 | Password root VPS 47.236.50.11 | MEMORY.md:2 | ⚠️ WAJIB ROTASI |
| R2 | Passcode root STB ssh.intim.my.id | MEMORY.md:4 | ⚠️ WAJIB ROTASI |
| R3 | Token bot Telegram "spy bot" (8774650894) | MEMORY.md:23 | ⚠️ WAJIB REVOKE via @BotFather |
| R4 | Token bot Telegram bot_keuangan (8580817124) | bot_keuangan.py:11 (+ .save files) | ⚠️ WAJIB REVOKE via @BotFather |
