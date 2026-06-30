# Project & VPS Memory
- **Source VPS:** root@47.236.50.11 (Location: `/root`, `/var/www`, Passcode: `[REDACTED-ROTATE-PASSCODE]`)
- **Target STB (CasaOS):** root@ssh.xxayii.my.id (via Cloudflare Tunnel)
- **STB Passcode:** [REDACTED-ROTATE-PASSCODE]
- **Local STB IP:** 192.168.1.81
- **Project Backup:** Located at `/DATA/Downloads/gz/all_projects.tar.gz` on STB.
- **STB Remote Access:** Configuration stored in `STB_ACCESS.md` (Cloudflare Tunnel + SSH).
- **TV Xiaomi IP:** 192.168.1.68 (Scripts for interaction in `/root/send_text_tv.py`).

*Note: Credentials updated for convenience.*
- **Current Identity:** SUPERAGENT 🔥 (v4.2 OPENCLAW EDITION)

### Project Undangan Digital (intim.my.id)
- **Domain Utama:** `intim.my.id` & `www.intim.my.id` (pointing to STB Port 8088 Nginx)
- **Frontend Path:** `/var/www/html` on STB (Nginx serves themes & index)
- **Backend API Path:** `/var/www/undangan-api` (Port 8089 via Node.js, systemd `undangan-api.service`)
- **Nginx Config Path:** `/etc/nginx/sites-available/undangan`
- **Dynamic Routing:** Redirects `/v[1-7]/.+` to appropriate theme templates (e.g. `/themes/v5/index.html?id=...&to=...`)
- **Project History:** Transferred from VPS / `undangan.xxayii.my.id` to STB local hosting under `intim.my.id`.

### Automated Spy Bot Setup (STB)
- **Target iPhone IP:** 192.168.1.78
- **Telegram Bot:** [REDACTED-TOKEN-SPY-BOT]
- **Chat ID:** 580132327
- **Script Location:** /root/spy_bot.py
- **Log Location:** /DATA/spy_log.txt
- **Status:** Active via systemd (spy-bot.service)
