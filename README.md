<div align="center">

# 🦞 IntimClaw

### Lightweight Personal AI Assistant — Your Digital Companion

<!-- PHOTO: Ganti placeholder di bawah dengan foto/creenshot lo sendiri -->
<!-- <img src="assets/screenshot.png" alt="IntimClaw Demo" width="600"> -->
<!-- <img src="assets/logo.png" alt="IntimClaw Logo" width="200"> -->

---

**Language:** [English](#english) | [Bahasa Indonesia](#bahasa-indonesia)

</div>

---

# English

## What is IntimClaw?

**IntimClaw** is a lightweight personal AI assistant that runs entirely on your own hardware. It's a fork of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com), heavily customized and rewritten under the name **IntimClaw** with a custom persona, new tools, and a personal finance focus.

It's not just a chatbot — it's a **persistent agent** that lives on your server, remembers your conversations, and can execute real actions on your systems.

### Key Features

- **Run anywhere** — Linux, macOS, Android (Termux/UserLAnd), VPS, Raspberry Pi, set-top box
- **Ultra-lightweight** — 3-15 MB RAM (Go core), runs on $10 hardware with 10MB RAM
- **Telegram bot** — chat with your AI agent directly from Telegram
- **WebUI** — browser-based dashboard for configuration and chat
- **Tool calling** — execute shell commands, read/write files, SSH to remote servers
- **MCP Bridge** — connects to 74+ Python tools via Model Context Protocol
- **Persistent memory** — SQLite JSONL session storage, never loses context
- **Self-improving** — agent can learn new skills and evolve over time

## Quick Start

### Option 1: One-Line Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/xxayii57/intimclaw/main/install.sh | bash
```

### Option 2: Manual Install

1. Download the latest release from [GitHub Releases](https://github.com/xxayii57/intimclaw/releases/latest)
2. Extract and place in your PATH:
```bash
tar xzf intimclaw_Linux_arm64.tar.gz
chmod +x intimclaw intimclaw-launcher
mv intimclaw intimclaw-launcher ~/.local/bin/
```
3. Initialize:
```bash
intimclaw onboard
```
4. Set up your AI model (interactive wizard):
```bash
intimclaw model setup
```
5. Start chatting:
```bash
intimclaw agent          # Terminal chat
intimclaw gateway        # Telegram bot + gateway
intimclaw-launcher       # WebUI at http://localhost:18080
```

## Commands

| Command | Description |
|---|---|
| `intimclaw onboard` | Initialize config + workspace |
| `intimclaw model setup` | Interactive model setup wizard |
| `intimclaw model` | Show current default model |
| `intimclaw model add` | Add model from OpenAI-compatible endpoint |
| `intimclaw agent` | Chat with agent in terminal |
| `intimclaw agent -m "msg"` | One-shot message |
| `intimclaw gateway` | Start gateway (Telegram + channels) |
| `intimclaw-launcher` | Start WebUI + gateway |
| `intimclaw status` | Show system status |
| `intimclaw sessions list` | List saved sessions |
| `intimclaw update` | Check for updates |
| `intimclaw version` | Show version info |

### Telegram Commands

| Command | Description |
|---|---|
| `/sessions list` | List all sessions (inline keyboard) |
| `/sessions switch <id>` | Switch session |
| `/sessions info` | Current session info |
| `/show model` | Current model |
| `/list models` | All available models |
| `/switch model to <name>` | Switch model |
| `/clear` | Clear chat history |
| `/context` | Token usage stats |
| `/help` | Help |

## Architecture

```
┌─────────────────────────────────────────────┐
│  intimclaw-launcher (WebUI + Gateway)       │
│  └─ Serve Web UI + Proxy WebSocket          │
│           ↓                                 │
│  intimclaw gateway (Agent Engine + Tools)   │
│  ├─ Agent Loop (ReAct)                     │
│  ├─ Tool Execution (exec, ssh, file, etc.) │
│  ├─ MCP Bridge → 74+ Python tools          │
│  ├─ Telegram Channel (long-polling)        │
│  ├─ Session Store (SQLite JSONL)           │
│  └─ Tool Approval Gate (Telegram keyboard) │
└─────────────────────────────────────────────┘
```

## Built-in Tools (17)

| Tool | Description |
|---|---|
| `exec` | Execute shell commands |
| `ssh_exec` | Remote SSH execution |
| `read_file` | Read file contents |
| `write_file` | Write file contents |
| `edit_file` | Edit file contents |
| `list_dir` | List directory contents |
| `web_search` | Search the web |
| `web_fetch` | Fetch web page content |
| `cron` | Scheduled tasks |
| `spawn` | Spawn async tasks |
| `subagent` | Sub-agent orchestration |
| `delegate` | Delegate tasks |
| `send_file` | Send files to chat |
| `send_tts` | Text-to-speech |
| `message` | Send messages |
| `find_skills` | Find installed skills |
| `install_skill` | Install skills from registry |

### MCP Tools (74+ via MicroClaw bridge)

MicroClaw tools are exposed via Model Context Protocol bridge:

| Category | Tools |
|---|---|
| **Crypto** | treasury, wallet, airdrop, rugcheck, contract_watch, backtest |
| **Security** | scam_sentinel, sybil_audit, opsec_checker, exploit_builder |
| **Content** | content, humanizer, video_pipeline, voice |
| **Research** | research_q, community_intel, alpha_radar |
| **Automation** | automation, watchdog, swarm, alerts |
| **Skills** | skill_engine, skill_forge, skill_integrity, skill_market |
| **Memory** | mem0, memory_engine |
| **Chat** | telegram_api, telegram, tv_send, tv_cast |

## Memory System

### Session Storage (SQLite JSONL)
- All messages stored permanently — never deleted
- Auto-compress when context window fills up (summarize older messages)
- Session-per-user, per-channel, per-topic
- Summary preserved for long-term context

### Long-term Memory
- **Permanent facts** stored in SQLite
- **Mem0 cloud** integration for semantic search
- Agent can save and recall facts across sessions

### Context Management
| Setting | Default | Description |
|---|---|---|
| `max_tokens` | 32768 | Output tokens per turn |
| `summarize_message_threshold` | 20 | Auto-summarize after 20 messages |
| `summarize_token_percent` | 75% | Summarize when 75% of context used |
| `max_tool_iterations` | 50 | Tool calls per turn |

## Configuration

Main config: `~/.intimclaw/config.json`

Key settings:
```json
{
  "agents": {
    "defaults": {
      "model_name": "free"
    }
  },
  "model_list": [
    {
      "model_name": "free",
      "model": "openrouter/poolside/laguna-s-2.1:free",
      "api_base": "https://your-provider.com/v1",
      "api_keys": ["your-api-key"],
      "provider": "openai"
    }
  ],
  "channels": {
    "telegram": {
      "enabled": true,
      "settings": {
        "token": "your-telegram-bot-token"
      }
    }
  },
  "tools": {
    "ssh": {
      "enabled": true,
      "allowed_hosts": ["your-server-ip"],
      "credentials": [...]
    },
    "approval": {
      "enabled": true,
      "chat_id": "your-telegram-user-id"
    }
  }
}
```

## Supported Providers

IntimClaw supports 30+ LLM providers via OpenAI-compatible API:

| Provider | Notes |
|---|---|
| OpenRouter | 100+ models, free tier available |
| OpenAI | GPT-4o, GPT-5, etc. |
| Anthropic | Claude Opus, Sonnet, Haiku |
| DeepSeek | DeepSeek-V3, DeepSeek-R1 |
| Groq | Fast inference |
| Gemini | Google AI Studio |
| Moonshot/Kimi | Chinese models |
| Ollama | Local, free |
| Custom | Any OpenAI-compatible endpoint |

## Hardware Requirements

| Component | Minimum | Recommended |
|---|---|---|
| RAM | 10 MB (gateway only) | 50 MB (with MCP tools) |
| Storage | 20 MB | 200 MB (with tools + dependencies) |
| CPU | Any | ARM64/x86_64 |
| Network | Required for LLM API | Stable connection |

## Team

| Role | Name |
|---|---|
| Creator & Maintainer | [xxayii](https://github.com/xxayii57) |

## License

**MIT License** — Copyright (c) 2026 IntimClaw contributors

This project is a fork of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com). The original PicoClaw project is licensed under MIT. IntimClaw maintains the same MIT license.

## Acknowledgements

**Special thanks to [Sipeed](https://sipeed.com) and the [PicoClaw](https://github.com/sipeed/picoclaw) contributors** for creating the original ultra-lightweight AI assistant framework that IntimClaw is built upon. PicoClaw's vision of running AI on $10 hardware with <10MB RAM inspired the entire architecture. Without their pioneering work, IntimClaw would not exist.

> *"PicoClaw is an independent open-source project initiated by Sipeed, written entirely in Go from scratch."*

---

# Bahasa Indonesia

## Apa itu IntimClaw?

**IntimClaw** adalah asisten AI pribadi ringan yang berjalan sepenuhnya di hardware lo sendiri. Ini adalah fork dari [PicoClaw](https://github.com/sipeed/picoclaw) oleh [Sipeed](https://siped.com), yang sudah dirombak total dengan nama **IntimClaw**, persona kustom, tools baru, dan fokus keuangan pribadi.

Bukan cuma chatbot — ini **agen persisten** yang hidup di server lo, ingat percakapan, dan bisa eksekusi aksi beneran di sistem lo.

### Fitur Utama

- **Jalan di mana aja** — Linux, macOS, Android (Termux/UserLAnd), VPS, Raspberry Pi, set-top box
- **Sangat ringan** — 3-15 MB RAM (core Go), bisa jalan di hardware $10 dengan 10MB RAM
- **Telegram bot** — chat langsung dari Telegram
- **WebUI** — dashboard berbasis browser buat konfigurasi dan chat
- **Tool calling** — eksekusi shell command, baca/tulis file, SSH ke server remote
- **MCP Bridge** — connect ke 74+ tools Python lewat Model Context Protocol
- **Memory persisten** — SQLite JSONL, gak pernah kehilangan konteks
- **Self-improving** — agent bisa belajar skill baru dan berkembang sendiri

## Quick Start

### Opsi 1: One-Line Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/xxayii57/intimclaw/main/install.sh | bash
```

### Opsi 2: Install Manual

1. Download release terbaru dari [GitHub Releases](https://github.com/xxayii57/intimclaw/releases/latest)
2. Extract:
```bash
tar xzf intimclaw_Linux_arm64.tar.gz
chmod +x intimclaw intimclaw-launcher
mv intimclaw intimclaw-launcher ~/.local/bin/
```
3. Inisialisasi:
```bash
intimclaw onboard
```
4. Setup model AI (wizard interaktif):
```bash
intimclaw model setup
```
5. Mulai chat:
```bash
intimclaw agent          # Chat di terminal
intimclaw gateway        # Telegram bot + gateway
intimclaw-launcher       # WebUI di http://localhost:18080
```

## Perintah

| Perintah | Keterangan |
|---|---|
| `intimclaw onboard` | Inisialisasi config + workspace |
| `intimclaw model setup` | Wizard setup model interaktif |
| `intimclaw model` | Tampilkan model default saat ini |
| `intimclaw model add` | Tambah model dari endpoint OpenAI-compatible |
| `intimclaw agent` | Chat dengan agent di terminal |
| `intimclaw agent -m "pesan"` | Kirim pesan sekali |
| `intimclaw gateway` | Jalankan gateway (Telegram + channels) |
| `intimclaw-launcher` | Jalankan WebUI di http://localhost:18080 |
| `intimclaw status` | Tampilkan status sistem |
| `intimclaw sessions list` | Daftar session tersimpan |
| `intimclaw update` | Cek update |
| `intimclaw version` | Tampilkan versi |

### Perintah Telegram

| Perintah | Keterangan |
|---|---|
| `/sessions list` | Daftar semua session (inline keyboard) |
| `/sessions switch <id>` | Ganti session |
| `/sessions info` | Info session aktif |
| `/show model` | Model saat ini |
| `/list models` | Semua model tersedia |
| `/switch model to <nama>` | Ganti model |
| `/clear` | Hapus riwayat chat |
| `/context` | Statistik token |
| `/help` | Bantuan |

## Arsitektur

```
┌─────────────────────────────────────────────┐
│  intimclaw-launcher (WebUI + Gateway)       │
│  └─ Serve Web UI + Proxy WebSocket          │
│           ↓                                 │
│  intimclaw gateway (Agent Engine + Tools)   │
│  ├─ Agent Loop (ReAct)                     │
│  ├─ Tool Execution (exec, ssh, file, dll.) │
│  ├─ MCP Bridge → 74+ tools Python          │
│  ├─ Telegram Channel (long-polling)        │
│  ├─ Session Store (SQLite JSONL)           │
│  └─ Tool Approval Gate (Telegram keyboard) │
└─────────────────────────────────────────────┘
```

## Tool Bawaan (17)

| Tool | Keterangan |
|---|---|
| `exec` | Jalankan shell command |
| `ssh_exec` | Eksekusi SSH remote |
| `read_file` | Baca isi file |
| `write_file` | Tulis file |
| `edit_file` | Edit file |
| `list_dir` | Lihat isi direktori |
| `web_search` | Cari di internet |
| `web_fetch` | Ambil konten halaman web |
| `cron` | Scheduled tasks |
| `spawn` | Spawn task async |
| `subagent` | Sub-agent orchestration |
| `delegate` | Delegasi tugas |
| `send_file` | Kirim file ke chat |
| `send_tts` | Text-to-speech |
| `message` | Kirim pesan |
| `find_skills` | Cari skill terinstall |
| `install_skill` | Install skill dari registry |

### Tool MCP (74+ via MicroClaw bridge)

| Kategori | Tools |
|---|---|
| **Crypto** | treasury, wallet, airdrop, rugcheck, contract_watch, backtest |
| **Security** | scam_sentinel, sybil_audit, opsec_checker, exploit_builder |
| **Content** | content, humanizer, video_pipeline, voice |
| **Research** | research_q, community_intel, alpha_radar |
| **Automation** | automation, watchdog, swarm, alerts |
| **Skills** | skill_engine, skill_forge, skill_integrity, skill_market |
| **Memory** | mem0, memory_engine |
| **Chat** | telegram_api, telegram, tv_send, tv_cast |

## Sistem Memory

### Session Storage (SQLite JSONL)
- Semua pesan tersimpan permanen — gak pernah dihapus
- Auto-compress kalau context window penuh (ringkas pesan lama)
- Session per-user, per-channel, per-topic
- Ringkasan tetap dipertahankan untuk konteks jangka panjang

### Long-term Memory
- **Fakta permanen** disimpan di SQLite
- **Mem0 cloud** integrasi untuk semantic search
- Agent bisa simpan dan recall fakta lintas session

### Context Management
| Setting | Default | Keterangan |
|---|---|---|
| `max_tokens` | 32768 | Output token per turn |
| `summarize_message_threshold` | 20 | Auto-ringkas setelah 20 pesan |
| `summarize_token_percent` | 75% | Ringkas saat 75% context terpakai |
| `max_tool_iterations` | 50 | Tool call per turn |

## Konfigurasi

Config utama: `~/.intimclaw/config.json`

```json
{
  "agents": {
    "defaults": {
      "model_name": "free"
    }
  },
  "model_list": [
    {
      "model_name": "free",
      "model": "openrouter/poolside/laguna-s-2.1:free",
      "api_base": "https://provider-lo.com/v1",
      "api_keys": ["api-key-lo"],
      "provider": "openai"
    }
  ],
  "channels": {
    "telegram": {
      "enabled": true,
      "settings": {
        "token": "token-bot-telegram-lo"
      }
    }
  },
  "tools": {
    "ssh": {
      "enabled": true,
      "allowed_hosts": ["ip-server-lo"],
      "credentials": [...]
    },
    "approval": {
      "enabled": true,
      "chat_id": "user-id-telegram-lo"
    }
  }
}
```

## Provider yang Didukung

IntimClaw support 30+ LLM provider lewat API OpenAI-compatible:

| Provider | Catatan |
|---|---|
| OpenRouter | 100+ model, ada free tier |
| OpenAI | GPT-4o, GPT-5, dll. |
| Anthropic | Claude Opus, Sonnet, Haiku |
| DeepSeek | DeepSeek-V3, DeepSeek-R1 |
| Groq | Inference cepat |
| Gemini | Google AI Studio |
| Moonshot/Kimi | Model Chinese |
| Ollama | Lokal, gratis |
| OpenAI-compatible lainnya | Custom endpoint |

## Spesifikasi Hardware

| Komponen | Minimum | Recommended |
|---|---|---|
| RAM | 10 MB (gateway saja) | 50 MB (dengan MCP tools) |
| Storage | 20 MB | 200 MB (tools + dependencies) |
| CPU | Apa saja | ARM64/x86_64 |
| Network | Perlu untuk LLM API | Koneksi stabil |

## Tim

| Peran | Nama |
|---|---|
| Creator & Maintainer | [xxayii](https://github.com/xxayii57) |

## Lisensi

**MIT License** — Copyright (c) 2026 IntimClaw contributors

Project ini adalah fork dari [PicoClaw](https://github.com/siped/picoclaw) oleh [Sipeed](https://siped.com). PicoClaw asli dilisensikan di bawah MIT. IntimClaw mempertahankan lisensi MIT yang sama.

## Penghargaan

**Terima kasih khusus kepada [Sipeed](https://siped.com) dan para kontributor [PicoClaw](https://github.com/siped/picoclaw)** yang telah menciptakan framework asisten AI ultraringan tempat IntimClaw dibangun. Visi PicoClaw menjalankan AI di hardware $10 dengan <10MB RAM telah menginspirasi seluruh arsitektur ini. Tanpa kerja pioneer mereka, IntimClaw tidak akan ada.

> *"PicoClaw is an independent open-source project initiated by Sipeed, written entirely in Go from scratch."*

---

<div align="center">

**Built with ❤️ by xxayii**

*Based on [PicoClaw](https://github.com/siped/picoclaw) by [Siped](https://siped.com)*

</div>
