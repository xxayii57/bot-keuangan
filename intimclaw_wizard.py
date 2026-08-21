#!/usr/bin/env python3
import os
import sys

def get_paths():
    home = os.environ.get("HOME", os.path.expanduser("~"))
    is_termux = "PREFIX" in os.environ and "com.termux" in os.environ["PREFIX"]
    
    if is_termux or os.path.exists("/data/data/com.termux/files/usr/"):
        base_dir = os.path.join(home, ".intimclaw")
        return {
            "is_termux": True,
            "config_dir": base_dir,
            "config_path": os.path.join(base_dir, "config.toml"),
            "skills_dir": os.path.join(base_dir, "skills"),
            "soul_dir": os.path.join(base_dir, "superintim"),
            "sessions_dir": os.path.join(base_dir, "sessions"),
            "memory_dir": base_dir
        }
    else:
        # Standard XDG Directory specification for Linux/macOS
        config_dir = os.environ.get("XDG_CONFIG_HOME", os.path.join(home, ".config"))
        data_dir = os.environ.get("XDG_DATA_HOME", os.path.join(home, ".local", "share"))
        
        intimclaw_config = os.path.join(config_dir, "intimclaw")
        intimclaw_data = os.path.join(data_dir, "intimclaw")
        
        return {
            "is_termux": False,
            "config_dir": intimclaw_config,
            "config_path": os.path.join(intimclaw_config, "config.toml"),
            "skills_dir": os.path.join(intimclaw_data, "skills"),
            "soul_dir": os.path.join(intimclaw_data, "superintim"),
            "sessions_dir": os.path.join(intimclaw_data, "sessions"),
            "memory_dir": intimclaw_data
        }

def check_config():
    paths = get_paths()
    config_path = paths["config_path"]
    
    if not os.path.exists(config_path):
        return False
        
    with open(config_path, "r") as f:
        content = f.read()
        
    # Check if we have a model_provider configured
    provider = ""
    for line in content.split("\n"):
        if line.strip().startswith("model_provider"):
            parts = line.split("=")
            if len(parts) == 2:
                provider = parts[1].replace('"', '').replace("'", "").strip()
                
    if not provider:
        return False
        
    # Look for [providers.provider] block and check api_key
    target_block = f"[providers.{provider}]"
    if target_block not in content:
        return False
        
    block_index = content.find(target_block)
    after_block = content[block_index:]
    
    # Find next section or end of file to isolate provider configuration
    next_section = len(content)
    for section_name in ["[providers.", "[agent]", "[skills]", "[memory]", "[security]"]:
        idx = after_block.find(section_name, len(target_block))
        if idx != -1 and idx < next_section:
            next_section = idx
            
    provider_config = after_block[:next_section]
    
    # Check if api_key has a value
    for line in provider_config.split("\n"):
        if line.strip().startswith("api_key"):
            parts = line.split("=")
            if len(parts) == 2:
                val = parts[1].replace('"', '').replace("'", "").strip()
                if val:
                    return True
                    
    return False

def run_wizard():
    print("\n=======================================================")
    print("         WELCOME TO INTIMCLAW SETUP WIZARD             ")
    print("=======================================================")
    print("Sepertinya ini pertama kalinya Anda menjalankan IntimClaw.")
    print("Silakan lakukan konfigurasi AI Provider Anda agar Agent bisa bekerja.")
    print("-------------------------------------------------------")
    
    print("\nPilih AI Provider Utama Anda:")
    print("1) Google Gemini (Sangat disarankan - Gratis & Cepat)")
    print("2) Anthropic Claude (Sangat Pintar - Butuh API Key)")
    print("3) OpenAI GPT (Standar Industri)")
    print("4) Groq (Super Cepat)")
    print("5) Custom OpenAI-Compatible (9router, Ollama, dll)")
    
    choice = input("\nMasukkan nomor pilihan Anda (1-5): ").strip()
    
    provider = ""
    default_model = ""
    api_url = ""
    
    if choice == "1":
        provider = "gemini"
        default_model = "gemini-2.5-flash"
        api_url = "https://generativelanguage.googleapis.com/v1beta"
    elif choice == "2":
        provider = "anthropic"
        default_model = "claude-3-5-sonnet-latest"
        api_url = "https://api.anthropic.com/v1"
    elif choice == "3":
        provider = "openai"
        default_model = "gpt-4o-mini"
        api_url = "https://api.openai.com/v1"
    elif choice == "4":
        provider = "groq"
        default_model = "llama-3.3-70b-versatile"
        api_url = "https://api.groq.com/openai/v1"
    else:
        provider = "custom"
        api_url = input("Masukkan Base URL API Anda (contoh: https://9router.intim.my.id/v1): ").strip()
        default_model = input("Masukkan Nama Model Default (contoh: deepseek-v3): ").strip()
        
    api_key = input(f"Masukkan API Key {provider.upper()} Anda: ").strip()
    
    if not api_key:
        print("\n[Error] API Key tidak boleh kosong! Setup dibatalkan.")
        sys.exit(1)
        
    paths = get_paths()
    
    os.makedirs(paths["config_dir"], exist_ok=True)
    os.makedirs(paths["skills_dir"], exist_ok=True)
    os.makedirs(paths["soul_dir"], exist_ok=True)
    os.makedirs(paths["sessions_dir"], exist_ok=True)
    
    config_path = paths["config_path"]
    
    # Generate config.toml
    provider_type = "anthropic" if choice == "2" else "openai-compatible"
    config_content = f"""# config.toml — IntimClaw Configuration
[agent]
name = "intimclaw"
version = "0.1.0"
model_provider = "{provider}"
model = "{default_model}"
persona = "superintim"

[providers.{provider}]
type = "{provider_type}"
base_url = "{api_url}"
api_key = "{api_key}"
models = ["{default_model}"]
"""
    
    # Add other providers for user manual setup references
    if provider != "gemini":
        config_content += """
[providers.gemini]
type = "openai-compatible"
base_url = "https://generativelanguage.googleapis.com/v1beta"
api_key = ""
models = ["gemini-2.5-flash", "gemini-2.5-pro"]
"""
    if provider != "anthropic":
        config_content += """
[providers.anthropic]
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = ""
models = ["claude-3-5-sonnet-latest"]
"""
    if provider != "openai":
        config_content += """
[providers.openai]
type = "openai-compatible"
base_url = "https://api.openai.com/v1"
api_key = ""
models = ["gpt-4o-mini"]
"""
    if provider != "groq":
        config_content += """
[providers.groq]
type = "openai-compatible"
base_url = "https://api.groq.com/openai/v1"
api_key = ""
models = ["llama-3.3-70b-versatile"]
"""

    # Add other configuration blocks
    config_content += f"""
[skills]
enabled = true
directories = ["{paths["skills_dir"]}"]

[memory]
backend = "sqlite"
semantic_search = true

[security]
risk_profile = "default"
sandbox = false
excluded_tools = ["rm", "mkfs", "dd", "shutdown"]
forbidden_paths = [".ssh", ".gnupg"]
"""

    with open(config_path, "w") as f:
        f.write(config_content)
        
    # Generate superintim/SOUL.md
    soul_path = os.path.join(paths["soul_dir"], "SOUL.md")
    if not os.path.exists(soul_path):
        soul_content = """# SOUL OF INTIMCLAW AGENTIC CODER

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
"""
        with open(soul_path, "w") as f:
            f.write(soul_content)

    print("=======================================================")
    print("🎉 SETUP SELESAI COK! Konfigurasi tersimpan di:")
    print(f"   {config_path}")
    print("-------------------------------------------------------")
    print("Mulai masuk ke IntimClaw...")
    print("=======================================================\n")

if __name__ == "__main__":
    if not check_config():
        run_wizard()
    else:
        # Config exists and valid
        pass
