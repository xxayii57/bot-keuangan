// intimclaw.ic — Built by xxayii — IntimClaw WebUI v0.1.0
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func getMemoryDBPath() string {
	return config.GetMemoryPath()
}

// generateAPIToken creates a cryptographically random hex token.
func generateAPIToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback: shouldn't fail on any sane OS
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}

// requireAuth wraps an http.HandlerFunc, requiring a valid API token.
// Token is checked via Authorization: Bearer <token> header or ?token= query param.
func requireAuth(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check Authorization header
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				if strings.TrimPrefix(authHeader, "Bearer ") == token {
					next(w, r)
					return
				}
			}
		}
		// Check query parameter
		if r.URL.Query().Get("token") == token {
			next(w, r)
			return
		}
		jsonError(w, "Unauthorized: valid API token required", 401)
	}
}

// maskSecret returns a masked version of a secret string.
// e.g. "sk-abc123xyz" → "sk-***xyz"
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

func startWebUI(port int, host string) {
	cfg, _ = config.Load()
	startTime = time.Now()
	webuiDir := getWebUIDir()

	// Generate API token if not set
	if cfg.WebUI.APIToken == "" {
		cfg.WebUI.APIToken = generateAPIToken()
		config.Save(cfg)
		// Write token to file for client discovery
		tokenPath := filepath.Join(filepath.Dir(config.GetConfigPath()), "webui.token")
		os.WriteFile(tokenPath, []byte(cfg.WebUI.APIToken), 0600)
		fmt.Printf("[intimclaw] Generated API token: %s\n", cfg.WebUI.APIToken)
		fmt.Printf("[intimclaw] Token saved to: %s\n", tokenPath)
	}
	token := cfg.WebUI.APIToken

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

	// Core — /api/health is public, all others require auth
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/chat", requireAuth(token, handleChat))
	http.HandleFunc("/ws/chat", requireAuth(token, handleWSChat))
	http.HandleFunc("/api/dashboard", requireAuth(token, handleDashboard))

	// Config
	http.HandleFunc("/api/settings", requireAuth(token, handleSettings))
	http.HandleFunc("/api/settings/save", requireAuth(token, handleSettingsSave))
	http.HandleFunc("/api/settings/raw", requireAuth(token, handleSettingsRaw))
	http.HandleFunc("/api/settings/raw/save", requireAuth(token, handleSettingsRawSave))
	http.HandleFunc("/api/settings/default", requireAuth(token, handleSettingsDefault))

	// Providers
	http.HandleFunc("/api/providers", requireAuth(token, handleProviders))
	http.HandleFunc("/api/providers/test", requireAuth(token, handleProviderTest))

	// Channels
	http.HandleFunc("/api/channels", requireAuth(token, handleChannels))
	http.HandleFunc("/api/channels/toggle", requireAuth(token, handleChannelToggle))
	http.HandleFunc("/api/channels/save", requireAuth(token, handleChannelSave))

	// Skills
	http.HandleFunc("/api/skills", requireAuth(token, handleSkills))
	http.HandleFunc("/api/skills/toggle", requireAuth(token, handleSkillToggle))

	// Memory
	http.HandleFunc("/api/memory", requireAuth(token, handleMemory))
	http.HandleFunc("/api/memory/delete", requireAuth(token, handleMemoryDelete))
	http.HandleFunc("/api/memory/add", requireAuth(token, handleMemoryAdd))
	http.HandleFunc("/api/memory/toggle-pin", requireAuth(token, handleMemoryTogglePin))
	http.HandleFunc("/api/learning", requireAuth(token, handleLearning))

	// MCP
	http.HandleFunc("/api/mcp", requireAuth(token, handleMCP))
	http.HandleFunc("/api/mcp/add", requireAuth(token, handleMCPAdd))
	http.HandleFunc("/api/mcp/remove", requireAuth(token, handleMCPRemove))

	// Sessions
	http.HandleFunc("/api/sessions", requireAuth(token, handleSessions))
	http.HandleFunc("/api/sessions/", requireAuth(token, handleSessionMessages))

	// Logs
	http.HandleFunc("/api/logs", requireAuth(token, handleLogs))

	fmt.Printf("[intimclaw] intimclaw.ic WebUI: http://%s:%d\n", host, port)
	fmt.Printf("[intimclaw] API token required for all /api/* endpoints (except /api/health)\n")
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
		// Allow same-origin connections and localhost
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Non-browser clients (curl, etc.)
		}
		// Allow localhost origins
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			return true
		}
		// Allow the configured host
		if cfg != nil && cfg.WebUI.Host != "" && cfg.WebUI.Host != "0.0.0.0" {
			allowed := fmt.Sprintf("http://%s:%d", cfg.WebUI.Host, cfg.WebUI.Port)
			if origin == allowed || origin == "https://"+strings.TrimPrefix(allowed, "http://") {
				return true
			}
		}
		// If bound to 0.0.0.0, allow all (user explicitly chose public mode)
		if cfg != nil && cfg.WebUI.Host == "0.0.0.0" {
			return true
		}
		return false
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
	// Create a masked copy of the config to avoid leaking secrets
	masked := *cfg
	masked.Providers = make([]config.ProviderConfig, len(cfg.Providers))
	for i, p := range cfg.Providers {
		masked.Providers[i] = p
		masked.Providers[i].APIKey = maskSecret(p.APIKey)
	}
	masked.Channels.Telegram.BotToken = maskSecret(cfg.Channels.Telegram.BotToken)
	masked.Channels.Discord.BotToken = maskSecret(cfg.Channels.Discord.BotToken)
	masked.Channels.Email.ImapPass = maskSecret(cfg.Channels.Email.ImapPass)
	masked.WebUI.APIToken = maskSecret(cfg.WebUI.APIToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&masked)
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

	// Preserve API token — never allow it to be cleared via this endpoint
	if newCfg.WebUI.APIToken == "" || strings.Contains(newCfg.WebUI.APIToken, "***") {
		newCfg.WebUI.APIToken = cfg.WebUI.APIToken
	}

	// Preserve provider API keys if masked
	for i, p := range newCfg.Providers {
		if strings.Contains(p.APIKey, "***") && i < len(cfg.Providers) {
			// Find matching provider by name and restore real key
			for _, orig := range cfg.Providers {
				if orig.Name == p.Name {
					newCfg.Providers[i].APIKey = orig.APIKey
					break
				}
			}
		}
	}

	// Preserve channel tokens if masked
	if strings.Contains(newCfg.Channels.Telegram.BotToken, "***") {
		newCfg.Channels.Telegram.BotToken = cfg.Channels.Telegram.BotToken
	}
	if strings.Contains(newCfg.Channels.Discord.BotToken, "***") {
		newCfg.Channels.Discord.BotToken = cfg.Channels.Discord.BotToken
	}

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

	// Parse, mask secrets, re-encode to TOML
	var cfgCopy config.Config
	if err := toml.Unmarshal(data, &cfgCopy); err != nil {
		// If parse fails, return raw (shouldn't happen with valid config)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
		return
	}

	// Mask all secrets
	for i := range cfgCopy.Providers {
		cfgCopy.Providers[i].APIKey = maskSecret(cfgCopy.Providers[i].APIKey)
	}
	cfgCopy.Channels.Telegram.BotToken = maskSecret(cfgCopy.Channels.Telegram.BotToken)
	cfgCopy.Channels.Discord.BotToken = maskSecret(cfgCopy.Channels.Discord.BotToken)
	cfgCopy.Channels.Email.ImapPass = maskSecret(cfgCopy.Channels.Email.ImapPass)
	cfgCopy.WebUI.APIToken = maskSecret(cfgCopy.WebUI.APIToken)

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(&cfgCopy); err != nil {
		jsonError(w, "Failed to encode config: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(buf.Bytes())
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

	// Validate that the body is valid TOML before writing
	var testCfg config.Config
	if err := toml.Unmarshal(body, &testCfg); err != nil {
		jsonError(w, "Invalid TOML: "+err.Error(), 400)
		return
	}

	// Reject if API token would be overwritten to empty
	if testCfg.WebUI.APIToken == "" && cfg.WebUI.APIToken != "" {
		testCfg.WebUI.APIToken = cfg.WebUI.APIToken
	}

	path := config.GetConfigPath()
	if err := os.WriteFile(path, body, 0644); err != nil {
		jsonError(w, "Failed to save raw config: "+err.Error(), 500)
		return
	}

	// Update in-memory config
	cfg = &testCfg

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
		// Reject masked API keys
		if strings.Contains(newProv.APIKey, "***") {
			jsonError(w, "API key appears to be masked. Send the real key.", 400)
			return
		}
		// Validate base URL is not empty for non-ollama providers
		if newProv.BaseURL == "" && newProv.Type != "ollama" {
			jsonError(w, "base_url is required", 400)
			return
		}
		if newProv.Type == "" {
			newProv.Type = "openai-compatible"
		}
		// Prevent duplicate provider names
		for _, p := range cfg.Providers {
			if p.Name == newProv.Name {
				jsonError(w, "Provider '"+newProv.Name+"' already exists. Use a different name.", 400)
				return
			}
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
			"bot_token":    maskSecret(cfg.Channels.Telegram.BotToken),
			"owner_id":     cfg.Channels.Telegram.OwnerID,
			"mention_only": cfg.Channels.Telegram.MentionOnly,
		},
		{
			"name":      "Discord",
			"enabled":   cfg.Channels.Discord.Enabled,
			"type":      "websocket",
			"desc":      "WebSocket gateway · slash commands",
			"bot_token": maskSecret(cfg.Channels.Discord.BotToken),
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
		jsonError(w, "Invalid data: "+err.Error(), 400)
		return
	}
	switch req.Name {
	case "Telegram":
		cfg.Channels.Telegram.Enabled = req.Enabled
	case "Discord":
		cfg.Channels.Discord.Enabled = req.Enabled
	case "WebChat":
		cfg.Channels.WebChat.Enabled = req.Enabled
	default:
		jsonError(w, "Unsupported channel: "+req.Name, 400)
		return
	}
	if err := config.Save(cfg); err != nil {
		jsonError(w, "Failed to save config: "+err.Error(), 500)
		return
	}
	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}
	jsonOK(w, req.Name+" channel "+status)
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

	// Reject masked secrets — client must send the real value or omit the field
	if strings.Contains(req.BotToken, "***") {
		jsonError(w, "Bot token appears to be masked. Send the real token or omit the field.", 400)
		return
	}

	switch req.Name {
	case "Telegram":
		cfg.Channels.Telegram.Enabled = req.Enabled
		if req.BotToken != "" {
			cfg.Channels.Telegram.BotToken = req.BotToken
		}
		cfg.Channels.Telegram.OwnerID = req.OwnerID
		cfg.Channels.Telegram.MentionOnly = req.MentionOnly
	case "Discord":
		cfg.Channels.Discord.Enabled = req.Enabled
		if req.BotToken != "" {
			cfg.Channels.Discord.BotToken = req.BotToken
		}
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
	script := `import sys, sqlite3, json
try:
    conn = sqlite3.connect(sys.argv[1])
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
	out, err := runMemoryPythonArgs(script, getMemoryDBPath())
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
conn = sqlite3.connect(sys.argv[1])
cursor = conn.cursor()
cursor.execute('DELETE FROM memory_facts WHERE id = ?', (sys.argv[1],))
cursor.execute('DELETE FROM memory_evidence WHERE fact_id = ?', (sys.argv[1],))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, getMemoryDBPath(), req.ID)
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
conn = sqlite3.connect(sys.argv[1])
cursor = conn.cursor()
new_id = str(uuid.uuid4())
cursor.execute('''
INSERT INTO memory_facts (id, content, category, confidence, status, pinned, created_at, updated_at)
VALUES (?, ?, ?, 100, 'active', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
''', (new_id, sys.argv[1], sys.argv[2]))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, getMemoryDBPath(), content, category)
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
conn = sqlite3.connect(sys.argv[1])
cursor = conn.cursor()
pinned_val = 1 if sys.argv[2] == 'true' else 0
cursor.execute('UPDATE memory_facts SET pinned = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?', (pinned_val, sys.argv[1]))
conn.commit()
print("ok")
`
	_, err := runMemoryPythonArgs(script, getMemoryDBPath(), req.ID, fmt.Sprintf("%t", req.Pinned))
	if err != nil {
		jsonError(w, "Failed to toggle pin: "+err.Error(), 500)
		return
	}

	jsonOK(w, "Pin toggled successfully")
}

func handleLearning(w http.ResponseWriter, r *http.Request) {
	script := `import sys, sqlite3, json
try:
    conn = sqlite3.connect(sys.argv[1])
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
	out, err := runMemoryPythonArgs(script, getMemoryDBPath())
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
type SessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Session struct {
	ID       string           `json:"id"`
	Title    string           `json:"title"`
	Model    string           `json:"model"`
	Created  time.Time        `json:"created"`
	Messages []SessionMessage `json:"messages,omitempty"`
}

func sessionDataPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".intimclaw", "data", "sessions.json")
}

func loadSessions() []Session {
	data, err := os.ReadFile(sessionDataPath())
	if err != nil {
		return []Session{}
	}
	var sessions []Session
	json.Unmarshal(data, &sessions)
	return sessions
}

func saveSessions(sessions []Session) {
	out, _ := json.MarshalIndent(sessions, "", "  ")
	os.MkdirAll(filepath.Dir(sessionDataPath()), 0755)
	os.WriteFile(sessionDataPath(), out, 0644)
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sessions := loadSessions()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	case "POST":
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Title string `json:"title"`
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			jsonError(w, "Invalid data", 400)
			return
		}
		sessions := loadSessions()
		sess := Session{
			ID:      fmt.Sprintf("ses_%d", time.Now().UnixNano()),
			Title:   req.Title,
			Model:   req.Model,
			Created: time.Now(),
		}
		if sess.Title == "" {
			sess.Title = "New session"
		}
		if sess.Model == "" {
			sess.Model = cfg.Agent.Model
		}
		sessions = append(sessions, sess)
		saveSessions(sessions)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sess)
	default:
		jsonError(w, "Method not allowed", 405)
	}
}

func handleSessionMessages(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if id == "" {
		jsonError(w, "Session ID required", 400)
		return
	}
	sessions := loadSessions()
	for _, s := range sessions {
		if s.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.Messages)
			return
		}
	}
	jsonError(w, "Session not found", 404)
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
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		filepath.Join(dir, "webui"),
		filepath.Join(home, ".intimclaw", "webui"),
		"webui",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "webui"
}



func formatUptime(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}
