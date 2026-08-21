// intimclaw.ic — Built by xxayii — IntimClaw WebUI v0.1.0
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xxayii/intimclaw/internal/agent"
	"github.com/xxayii/intimclaw/internal/config"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/websocket"
)

type ChatRequest struct {
	Message string `json:"message"`
	Mode    string `json:"mode,omitempty"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

var (
	cfg        *config.Config
	startTime  = time.Now()
	queryCount = 0
)

func startWebUI(port int, host string) {
	cfg, _ = config.Load()
	startTime = time.Now()
	webuiDir := getWebUIDir()

	// Start Telegram bot if enabled
	if cfg.Channels.Telegram.Enabled && cfg.Channels.Telegram.BotToken != "" {
		fmt.Println("[intimclaw] Starting Telegram bot inside WebUI process...")
		a := agent.NewFromConfig(cfg)
		bot := agent.NewTelegramBot(cfg.Channels.Telegram.BotToken, a)
		go bot.Start()
	}

	// Start Discord bot if enabled
	if cfg.Channels.Discord.Enabled && cfg.Channels.Discord.BotToken != "" {
		fmt.Println("[intimclaw] Starting Discord bot inside WebUI process...")
		a := agent.NewFromConfig(cfg)
		dbot := agent.NewDiscordBot(cfg.Channels.Discord.BotToken, a)
		go dbot.Start()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webuiDir, "index.html"))
	})

	// Core
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/ws/chat", handleWSChat)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/dashboard", handleDashboard)

	// Config
	http.HandleFunc("/api/settings", handleSettings)
	http.HandleFunc("/api/settings/save", handleSettingsSave)
	http.HandleFunc("/api/settings/raw", handleSettingsRaw)
	http.HandleFunc("/api/settings/raw/save", handleSettingsRawSave)
	http.HandleFunc("/api/settings/default", handleSettingsDefault)

	// Providers
	http.HandleFunc("/api/providers", handleProviders)
	http.HandleFunc("/api/providers/test", handleProviderTest)

	// Channels
	http.HandleFunc("/api/channels", handleChannels)
	http.HandleFunc("/api/channels/toggle", handleChannelToggle)
	http.HandleFunc("/api/channels/save", handleChannelSave)

	// Skills
	http.HandleFunc("/api/skills", handleSkills)
	http.HandleFunc("/api/skills/toggle", handleSkillToggle)

	// Memory
	http.HandleFunc("/api/memory", handleMemory)
	http.HandleFunc("/api/memory/delete", handleMemoryDelete)
	http.HandleFunc("/api/memory/add", handleMemoryAdd)
	http.HandleFunc("/api/memory/toggle-pin", handleMemoryTogglePin)
	http.HandleFunc("/api/learning", handleLearning)

	// MCP
	http.HandleFunc("/api/mcp", handleMCP)
	http.HandleFunc("/api/mcp/add", handleMCPAdd)
	http.HandleFunc("/api/mcp/remove", handleMCPRemove)

	// Sessions
	http.HandleFunc("/api/sessions", handleSessions)
	http.HandleFunc("/api/sessions/", handleSessionMessages)

	// Logs
	http.HandleFunc("/api/logs", handleLogs)

	fmt.Printf("[intimclaw] intimclaw.ic WebUI: http://%s:%d\n", host, port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "[intimclaw] web error: %v\n", err)
	}
}

// ═══════════════════════════════════════════════════════════════
// Health
// ═══════════════════════════════════════════════════════════════
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"version": "0.1.0",
		"name":    "IntimClaw",
		"ic":      "intimclaw.ic",
	})
}

// ═══════════════════════════════════════════════════════════════
// Dashboard — real system data
// ═══════════════════════════════════════════════════════════════
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	uptime := int(time.Since(startTime).Seconds())
	disk := getDiskUsage()
	ram := getRAMUsage()
	
	a := agent.NewFromConfig(cfg)
	toolsCount := a.GetToolsCount()
	skillsCount := a.GetSkillsCount()

	data := map[string]interface{}{
		"ic":            "intimclaw.ic",
		"version":       "0.1.0",
		"uptime":        uptime,
		"uptime_human":  formatUptime(uptime),
		"model":         cfg.Agent.Provider + "/" + cfg.Agent.Model,
		"provider":      cfg.Agent.Provider,
		"persona":       cfg.Agent.Persona,
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS + "/" + runtime.GOARCH,
		"disk_usage":    disk,
		"ram_usage":     ram,
		"tools_count":   toolsCount,
		"skills_count":  skillsCount,
		"providers":     len(cfg.Providers),
		"channels": map[string]bool{
			"telegram": cfg.Channels.Telegram.Enabled,
			"discord":  cfg.Channels.Discord.Enabled,
			"webchat":  cfg.Channels.WebChat.Enabled,
		},
		"memory_backend":  cfg.Memory.Backend,
		"semantic_search": cfg.Memory.SemanticSearch,
		"risk_profile":    cfg.Security.RiskProfile,
		"sandbox":         cfg.Security.Sandbox,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ═══════════════════════════════════════════════════════════════
// Chat
// ═══════════════════════════════════════════════════════════════
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req ChatRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Message == "" {
		jsonError(w, "Message is required", 400)
		return
	}

	// Add mode prefix if specified
	msg := req.Message
	if req.Mode == "build" {
		msg = "[MODE:BUILD] " + msg
	} else if req.Mode == "plan" {
		msg = "[MODE:PLAN] " + msg
	}

	queryCount++
	a := agent.NewFromConfig(cfg)
	resp, err := a.Run(msg)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Response: resp})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWSChat(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[intimclaw] WS upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req struct {
			Message   string `json:"message"`
			Model     string `json:"model"`
			Mode      string `json:"mode"`
			SessionID string `json:"session_id,omitempty"`
		}
		if err := json.Unmarshal(msgData, &req); err != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"content": "Invalid request format",
			})
			continue
		}

		msg := req.Message
		if req.Mode == "build" {
			msg = "[MODE:BUILD] " + msg
		} else if req.Mode == "plan" {
			msg = "[MODE:PLAN] " + msg
		}

		queryCount++
		a := agent.NewFromConfig(cfg)
		a.SetModel(req.Model)

		finalAnswer, err := a.RunStream(msg, func(event agent.AgentEvent) {
			conn.WriteJSON(event)
		})

		if err != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"content": err.Error(),
			})
		} else {
			conn.WriteJSON(map[string]interface{}{
				"type":    "message",
				"content": finalAnswer,
			})
			if req.SessionID != "" {
				go a.ExtractMemory(req.SessionID, req.Message, finalAnswer)
			}
		}
	}
}

// ═══════════════════════════════════════════════════════════════
// Settings
// ═══════════════════════════════════════════════════════════════
func handleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func handleSettingsSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var newCfg config.Config
	if err := json.Unmarshal(body, &newCfg); err != nil {
		jsonError(w, "Invalid config", 400)
		return
	}
	// Save to TOML
	cfg = &newCfg
	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save settings: "+err.Error(), 500)
		return
	}
	jsonOK(w, "Settings saved")
}

func handleSettingsRaw(w http.ResponseWriter, r *http.Request) {
	path := config.GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		jsonError(w, "Failed to read raw config: "+err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func handleSettingsRawSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "Failed to read request body: "+err.Error(), 400)
		return
	}

	path := config.GetConfigPath()
	if err := os.WriteFile(path, body, 0644); err != nil {
		jsonError(w, "Failed to save raw config: "+err.Error(), 500)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		exec.Command("systemctl", "--user", "restart", "intimclaw").Run()
	}()

	jsonOK(w, "config.toml updated successfully. Restarting service to apply changes...")
}

func handleSettingsDefault(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(config.DefaultConfig()); err != nil {
		jsonError(w, "Failed to encode default config: "+err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(buf.Bytes())
}

// ═══════════════════════════════════════════════════════════════
// Providers
// ═══════════════════════════════════════════════════════════════
func handleProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		providers := make([]map[string]interface{}, 0)
		for _, p := range cfg.Providers {
			keyLen := 0
			if p.APIKey != "" {
				keyLen = len(p.APIKey)
			}
			providers = append(providers, map[string]interface{}{
				"name":     p.Name,
				"type":     p.Type,
				"base_url": p.BaseURL,
				"api_key":  keyLen > 0,
				"models":   p.Models,
				"primary":  p.Name == cfg.Agent.Provider,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(providers)
	case "POST":
		body, _ := io.ReadAll(r.Body)
		var newProv struct {
			Name    string   `json:"name"`
			Type    string   `json:"type"`
			BaseURL string   `json:"base_url"`
			APIKey  string   `json:"api_key"`
			Models  []string `json:"models"`
		}
		if err := json.Unmarshal(body, &newProv); err != nil || newProv.Name == "" {
			jsonError(w, "Invalid provider data", 400)
			return
		}
		if newProv.Type == "" {
			newProv.Type = "openai-compatible"
		}
		cfg.Providers = append(cfg.Providers, config.ProviderConfig{
			Name:    newProv.Name,
			Type:    newProv.Type,
			BaseURL: newProv.BaseURL,
			APIKey:  newProv.APIKey,
			Models:  newProv.Models,
		})
		if err := config.Save(cfg); err != nil {
			jsonError(w, "Failed to save provider: "+err.Error(), 500)
			return
		}
		jsonOK(w, "Provider added: "+newProv.Name)
	case "DELETE":
		body, _ := io.ReadAll(r.Body)
		var del struct{ Name string `json:"name"` }
		json.Unmarshal(body, &del)
		var newCfg []config.ProviderConfig
		for _, p := range cfg.Providers {
			if p.Name != del.Name {
				newCfg = append(newCfg, p)
			}
		}
		cfg.Providers = newCfg
		if err := config.Save(cfg); err != nil {
			jsonError(w, "Failed to delete provider: "+err.Error(), 500)
			return
		}
		jsonOK(w, "Provider deleted: "+del.Name)
	default:
		jsonError(w, "Method not allowed", 405)
	}
}

func handleProviderTest(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.Header().Set("Content-Type", "application/json")
	
	var prov config.ProviderConfig
	found := false
	for _, p := range cfg.Providers {
		if p.Name == name {
			prov = p
			found = true
			break
		}
	}
	if !found {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "provider not found"})
		return
	}

	url := prov.BaseURL
	isAnthropic := prov.Type == "anthropic" || strings.Contains(strings.ToLower(prov.Name), "anthropic")
	if isAnthropic {
		if url == "" {
			url = "https://api.anthropic.com"
		}
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": err.Error()})
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": err.Error()})
		return
	}
	resp.Body.Close()
	
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "ic": "intimclaw.ic"})
}

// ═══════════════════════════════════════════════════════════════
// Channels
// ═══════════════════════════════════════════════════════════════
func handleChannels(w http.ResponseWriter, r *http.Request) {
	channels := []map[string]interface{}{
		{
			"name":         "Telegram",
			"enabled":      cfg.Channels.Telegram.Enabled,
			"type":         "long-polling",
			"desc":         "Inline keyboard approval · vision support",
			"bot_token":    cfg.Channels.Telegram.BotToken,
			"owner_id":     cfg.Channels.Telegram.OwnerID,
			"mention_only": cfg.Channels.Telegram.MentionOnly,
		},
		{
			"name":      "Discord",
			"enabled":   cfg.Channels.Discord.Enabled,
			"type":      "websocket",
			"desc":      "WebSocket gateway · slash commands",
			"bot_token": cfg.Channels.Discord.BotToken,
		},
		{
			"name":    "WhatsApp",
			"enabled": false,
			"type":    "wa-webjs",
			"desc":    "wa-webjs bridge",
		},
		{
			"name":    "WebChat",
			"enabled": cfg.Channels.WebChat.Enabled,
			"type":    "websocket",
			"desc":    "Built-in WebSocket chat · port 18080",
			"port":    cfg.Channels.WebChat.Port,
			"host":    cfg.Channels.WebChat.Host,
		},
		{
			"name":    "Email",
			"enabled": false,
			"type":    "imap-smtp",
			"desc":    "IMAP / SMTP",
		},
		{
			"name":    "Matrix",
			"enabled": false,
			"type":    "matrix-sdk",
			"desc":    "Matrix SDK",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

func handleChannelToggle(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, "Channel toggled")
}

func handleChannelSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	
	var req struct {
		Name        string `json:"name"`
		Enabled     bool   `json:"enabled"`
		BotToken    string `json:"bot_token,omitempty"`
		OwnerID     int64  `json:"owner_id,omitempty"`
		MentionOnly bool   `json:"mention_only,omitempty"`
		Port        int    `json:"port,omitempty"`
		Host        string `json:"host,omitempty"`
	}
	
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "Invalid data: "+err.Error(), 400)
		return
	}

	switch req.Name {
	case "Telegram":
		cfg.Channels.Telegram.Enabled = req.Enabled
		cfg.Channels.Telegram.BotToken = req.BotToken
		cfg.Channels.Telegram.OwnerID = req.OwnerID
		cfg.Channels.Telegram.MentionOnly = req.MentionOnly
	case "Discord":
		cfg.Channels.Discord.Enabled = req.Enabled
		cfg.Channels.Discord.BotToken = req.BotToken
	case "WebChat":
		cfg.Channels.WebChat.Enabled = req.Enabled
		if req.Port > 0 {
			cfg.Channels.WebChat.Port = req.Port
		}
		if req.Host != "" {
			cfg.Channels.WebChat.Host = req.Host
		}
	default:
		jsonError(w, "Unsupported channel: "+req.Name, 400)
		return
	}

	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save config: "+err.Error(), 500)
		return
	}

	// Trigger a background self-restart via systemd user unit to apply changes
	go func() {
		time.Sleep(1 * time.Second)
		exec.Command("systemctl", "--user", "restart", "intimclaw").Run()
	}()

	jsonOK(w, "Channel settings saved. Restarting service to apply changes...")
}

// ═══════════════════════════════════════════════════════════════
// Skills
// ═══════════════════════════════════════════════════════════════
func getSkillCategory(name string) string {
	name = strings.ToLower(name)
	
	if strings.Contains(name, "hermes") {
		return "Web3 & Crypto"
	}
	
	if strings.Contains(name, "ctf") || 
		name == "sk32" || name == "m32" || 
		name == "sk51" || strings.Contains(name, "exploit") || 
		strings.Contains(name, "redteam") || strings.Contains(name, "offensive") {
		return "Security & CTF"
	}
	
	if name == "sk10" || name == "sk13" || name == "sk31" || 
		name == "sk33" || name == "sk34" || name == "sk36" || 
		name == "sk37" || name == "sk38" || name == "sk39" || 
		strings.Contains(name, "airdrop") || strings.Contains(name, "swap") || 
		strings.Contains(name, "bridge") || strings.Contains(name, "tokenomics") {
		return "Web3 & Crypto"
	}
	
	if strings.HasPrefix(name, "x") || 
		name == "sk0" || name == "m0" || 
		name == "sk52" || name == "sk53" || name == "sk54" || 
		name == "sk55" || name == "sk56" || name == "sk57" || 
		name == "sk58" || name == "sk59" || 
		strings.Contains(name, "audit") || strings.Contains(name, "strategy") || 
		strings.Contains(name, "debug") || strings.Contains(name, "reflection") {
		return "Audit & Strategy"
	}
	
	if name == "sk1" || name == "sk3" || name == "m3" || 
		name == "sk14" || name == "m14" || name == "sk15" || name == "m15" || 
		name == "sk18" || name == "sk20" || name == "sk22" || name == "sk23" || 
		name == "sk26" || name == "sk27" || name == "sk28" || name == "m28" || 
		name == "sk29" || name == "sk30" || name == "sk35" || name == "sk40" || 
		name == "sk41" || name == "sk42" || name == "sk50" || 
		strings.Contains(name, "monetize") || strings.Contains(name, "content") || 
		strings.Contains(name, "copywriting") || strings.Contains(name, "writing") || 
		strings.Contains(name, "media") || strings.Contains(name, "video") {
		return "Content & Business"
	}
	
	return "Automation & Code"
}

func handleSkills(w http.ResponseWriter, r *http.Request) {
	a := agent.NewFromConfig(cfg)
	skills := a.GetSkills()

	type SkillResp struct {
		Name    string `json:"name"`
		Desc    string `json:"desc"`
		Tag     string `json:"tag"`
		Source  string `json:"source"`
		Enabled bool   `json:"enabled"`
	}

	disabledMap := make(map[string]bool)
	for _, name := range cfg.Skills.DisabledSkills {
		disabledMap[name] = true
	}

	var res []SkillResp
	for _, s := range skills {
		if strings.HasPrefix(strings.ToLower(s.Name), "sk") {
			continue // Hide core brain skills from UI
		}
		if s.Description == "" {
			a.LoadSkillContent(s.Name)
		}

		tag := getSkillCategory(s.Name)

		res = append(res, SkillResp{
			Name:    s.Name,
			Desc:    s.Description,
			Tag:     tag,
			Source:  s.Source,
			Enabled: !disabledMap[s.Name],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handleSkillToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "Invalid data", 400)
		return
	}

	if strings.HasPrefix(strings.ToLower(req.Name), "sk") {
		jsonError(w, "Cannot modify core brain skills", 403)
		return
	}

	var newDisabled []string
	found := false
	for _, name := range cfg.Skills.DisabledSkills {
		if name == req.Name {
			found = true
			if req.Enabled {
				continue
			}
		}
		newDisabled = append(newDisabled, name)
	}
	if !req.Enabled && !found {
		newDisabled = append(newDisabled, req.Name)
	}
	cfg.Skills.DisabledSkills = newDisabled

	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save skill toggle: "+err.Error(), 500)
		return
	}

	jsonOK(w, "Skill toggle saved")
}

// ═══════════════════════════════════════════════════════════════
// Memory
// ═══════════════════════════════════════════════════════════════
func handleMemory(w http.ResponseWriter, r *http.Request) {
	script := `import sqlite3, json
try:
    conn = sqlite3.connect('/root/microclaw/microclaw.db')
    cursor = conn.cursor()
    # Initialize tables
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS memory_facts (
        id TEXT PRIMARY KEY,
        content TEXT,
        category TEXT,
        confidence INTEGER,
        status TEXT,
        pinned INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
    ''')
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS memory_evidence (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        fact_id TEXT,
        session_id TEXT,
        quote TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(fact_id) REFERENCES memory_facts(id) ON DELETE CASCADE
    )
    ''')
    
    cursor.execute("SELECT id, content, category, confidence, status, pinned FROM memory_facts ORDER BY status ASC, confidence DESC")
    facts = cursor.fetchall()
    res = []
    for f in facts:
        f_id, content, cat, conf, status, pinned = f
        cursor.execute("SELECT quote, timestamp FROM memory_evidence WHERE fact_id = ?", (f_id,))
        evs = cursor.fetchall()
        evidence_list = [{"quote": ev[0], "timestamp": ev[1]} for ev in evs]
        res.append({
            "id": f_id,
            "content": content,
            "category": cat,
            "confidence": conf,
            "status": status,
            "pinned": bool(pinned),
            "evidence": evidence_list
        })
    conn.close()
    print(json.dumps(res))
except Exception as e:
    print(json.dumps([]))
`
	out, err := runMemoryPythonArgs(script)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(out))
}

func handleMemoryDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ID == "" {
		jsonError(w, "Invalid request", 400)
		return
	}

	script := `import sys, sqlite3
conn = sqlite3.connect('/root/microclaw/microclaw.db')
cursor = conn.cursor()
cursor.execute('DELETE FROM memory_facts WHERE id = ?', (sys.argv[1],))
cursor.execute('DELETE FROM memory_evidence WHERE fact_id = ?', (sys.argv[1],))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, req.ID)
	if err != nil {
		jsonError(w, "Failed to delete fact: "+err.Error(), 500)
		return
	}

	jsonOK(w, "Memory deleted")
}

func handleMemoryAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Key      string `json:"key,omitempty"`
		Value    string `json:"value,omitempty"`
		Content  string `json:"content,omitempty"`
		Category string `json:"category,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "Invalid data", 400)
		return
	}

	content := req.Content
	if content == "" && req.Key != "" {
		content = req.Key
		if req.Value != "" {
			content += ": " + req.Value
		}
	}
	if content == "" {
		jsonError(w, "Content is required", 400)
		return
	}

	category := req.Category
	if category == "" {
		category = "general"
	}

	script := `import sys, sqlite3, uuid
conn = sqlite3.connect('/root/microclaw/microclaw.db')
cursor = conn.cursor()
new_id = str(uuid.uuid4())
cursor.execute('''
INSERT INTO memory_facts (id, content, category, confidence, status, pinned, created_at, updated_at)
VALUES (?, ?, ?, 100, 'active', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
''', (new_id, sys.argv[1], sys.argv[2]))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, content, category)
	if err != nil {
		jsonError(w, "Failed to add fact: "+err.Error(), 500)
		return
	}

	jsonOK(w, "Memory added")
}

func handleMemoryTogglePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		ID     string `json:"id"`
		Pinned bool   `json:"pinned"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ID == "" {
		jsonError(w, "Invalid request", 400)
		return
	}

	script := `import sys, sqlite3
conn = sqlite3.connect('/root/microclaw/microclaw.db')
cursor = conn.cursor()
pinned_val = 1 if sys.argv[2] == 'true' else 0
cursor.execute('UPDATE memory_facts SET pinned = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?', (pinned_val, sys.argv[1]))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, req.ID, fmt.Sprintf("%t", req.Pinned))
	if err != nil {
		jsonError(w, "Failed to toggle pin: "+err.Error(), 500)
		return
	}

	jsonOK(w, "Pin toggled successfully")
}

func handleLearning(w http.ResponseWriter, r *http.Request) {
	script := `import sqlite3, json
try:
    conn = sqlite3.connect('/root/microclaw/microclaw.db')
    cursor = conn.cursor()
    
    cursor.execute("CREATE TABLE IF NOT EXISTS experiences (id TEXT, task TEXT, actions TEXT, errors TEXT, solution TEXT, result TEXT, timestamp DATETIME)")
    cursor.execute("CREATE TABLE IF NOT EXISTS lessons (id TEXT, title TEXT, content TEXT, category TEXT, confidence REAL, used_count INTEGER, success_count INTEGER, created_at DATETIME, updated_at DATETIME)")

    cursor.execute("SELECT COUNT(*) FROM experiences")
    total_tasks = cursor.fetchone()[0]
    cursor.execute("SELECT COUNT(*) FROM experiences WHERE result = 'success'")
    success_tasks = cursor.fetchone()[0]
    failed_tasks = total_tasks - success_tasks

    cursor.execute("SELECT id, title, content, category, confidence, used_count, success_count FROM lessons ORDER BY confidence DESC")
    rows = cursor.fetchall()
    lessons_list = []
    for r in rows:
        lessons_list.append({
            "id": r[0],
            "title": r[1],
            "content": r[2],
            "category": r[3],
            "confidence": round(r[4] * 100, 1),
            "used": r[5],
            "success": r[6]
        })
        
    conn.close()
    
    res = {
        "total_tasks": total_tasks,
        "success_tasks": success_tasks,
        "failed_tasks": failed_tasks,
        "lessons": lessons_list
    }
    print(json.dumps(res))
except Exception as e:
    print(json.dumps({
        "total_tasks": 0,
        "success_tasks": 0,
        "failed_tasks": 0,
        "lessons": []
    }))
`
	out, err := runMemoryPythonArgs(script)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(out))
}

// ═══════════════════════════════════════════════════════════════
// MCP
// ═══════════════════════════════════════════════════════════════
func handleMCP(w http.ResponseWriter, r *http.Request) {
	a := agent.NewFromConfig(cfg)
	servers := a.GetMCPServers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func handleMCPAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	
	var req struct {
		Name    string   `json:"name"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Name == "" || req.Command == "" {
		jsonError(w, "Invalid MCP server data", 400)
		return
	}

	cfg.MCP.Enabled = true
	cfg.MCP.Servers = append(cfg.MCP.Servers, config.MCPServer{
		Name:    req.Name,
		Command: req.Command,
		Args:    req.Args,
	})

	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save MCP config: "+err.Error(), 500)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		exec.Command("systemctl", "--user", "restart", "intimclaw").Run()
	}()

	jsonOK(w, "MCP server added. Restarting service to apply changes...")
}

func handleMCPRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	
	var req struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Name == "" {
		jsonError(w, "Invalid request data", 400)
		return
	}

	var newServers []config.MCPServer
	for _, s := range cfg.MCP.Servers {
		if s.Name != req.Name {
			newServers = append(newServers, s)
		}
	}
	cfg.MCP.Servers = newServers

	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save MCP config: "+err.Error(), 500)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		exec.Command("systemctl", "--user", "restart", "intimclaw").Run()
	}()

	jsonOK(w, "MCP server removed. Restarting service...")
}

// ═══════════════════════════════════════════════════════════════
// Sessions
// ═══════════════════════════════════════════════════════════════
func handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []map[string]interface{}{
		{"id": "ses_001", "title": "IntimClaw setup", "model": cfg.Agent.Model, "created": "now"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func handleSessionMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

// ═══════════════════════════════════════════════════════════════
// Logs
// ═══════════════════════════════════════════════════════════════
func handleLogs(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("/var/log/syslog")
	var logLines []map[string]interface{}
	
	if err != nil {
		logLines = []map[string]interface{}{
			{"time": time.Now().Format("15:04:05"), "level": "info", "msg": "Syslog unreadable: " + err.Error()},
			{"time": time.Now().Format("15:04:05"), "level": "info", "msg": fmt.Sprintf("Model: %s/%s", cfg.Agent.Provider, cfg.Agent.Model)},
		}
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		var tempLines []string
		
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), "intimclaw") {
				tempLines = append(tempLines, line)
			}
		}
		
		start := len(tempLines) - 50
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(tempLines); i++ {
			line := tempLines[i]
			parts := strings.SplitN(line, "intimclaw", 2)
			msg := line
			logTime := time.Now().Format("15:04:05")
			level := "info"
			
			if len(parts) > 1 {
				msg = parts[1]
				subParts := strings.SplitN(msg, ":", 2)
				if len(subParts) > 1 {
					msg = strings.TrimSpace(subParts[1])
				}
				
				timePart := strings.Fields(parts[0])
				if len(timePart) > 0 {
					t, err := time.Parse(time.RFC3339, timePart[0])
					if err == nil {
						logTime = t.Format("15:04:05")
					} else {
						if len(timePart) >= 3 {
							logTime = strings.Join(timePart[1:3], " ")
						}
					}
				}
			}
			
			if strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "fail") {
				level = "error"
			} else if strings.Contains(strings.ToLower(msg), "warn") {
				level = "warn"
			}
			
			logLines = append(logLines, map[string]interface{}{
				"time":  logTime,
				"level": level,
				"msg":   msg,
			})
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logLines)
}

// ═══════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════
func runMemoryPythonArgs(script string, args ...string) (string, error) {
	cmdArgs := append([]string{"-c", script}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(out))
	}
	return string(out), nil
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ChatResponse{Error: msg})
}

func jsonOK(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "message": msg, "ic": "intimclaw.ic"})
}

func getWebUIDir() string {
	ex, _ := os.Executable()
	dir := filepath.Dir(ex)
	for _, p := range []string{
		filepath.Join(dir, "webui"),
		"webui",
		"/root/intimclaw/webui",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "webui"
}

func getDiskUsage() string {
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return "N/A"
	}
	
	freeBytes := stat.Bavail * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.1fGB free / %.1fGB total", freeGB, totalGB)
}

func getRAMUsage() string {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "N/A"
	}
	defer file.Close()

	var memTotal, memAvailable, memFree uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch name {
		case "MemTotal":
			memTotal = val
		case "MemAvailable":
			memAvailable = val
		case "MemFree":
			memFree = val
		}
	}
	
	if memAvailable == 0 {
		memAvailable = memFree
	}

	totalMB := float64(memTotal) / 1024
	availableMB := float64(memAvailable) / 1024
	
	if totalMB > 1024 {
		return fmt.Sprintf("%.1fGB available / %.1fGB total", availableMB/1024, totalMB/1024)
	}
	return fmt.Sprintf("%.0fMB available / %.0fMB total", availableMB, totalMB)
}

func formatUptime(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}
