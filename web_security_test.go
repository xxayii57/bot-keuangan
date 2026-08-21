package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xxayii/intimclaw/internal/config"
)

func TestRequireAuth_MissingToken(t *testing.T) {
	cfg = &config.Config{
		WebUI: config.WebUIConfig{APIToken: "test-secret-token-12345"},
	}

	handler := requireAuth(cfg.WebUI.APIToken, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Errorf("expected Unauthorized message, got: %s", w.Body.String())
	}
}

func TestRequireAuth_ValidHeaderToken(t *testing.T) {
	cfg = &config.Config{
		WebUI: config.WebUIConfig{APIToken: "test-secret-token-12345"},
	}

	handler := requireAuth(cfg.WebUI.APIToken, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer test-secret-token-12345")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got: %s", w.Body.String())
	}
}

func TestRequireAuth_ValidQueryToken(t *testing.T) {
	cfg = &config.Config{
		WebUI: config.WebUIConfig{APIToken: "test-secret-token-12345"},
	}

	handler := requireAuth(cfg.WebUI.APIToken, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/settings?token=test-secret-token-12345", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuth_WrongToken(t *testing.T) {
	cfg = &config.Config{
		WebUI: config.WebUIConfig{APIToken: "test-secret-token-12345"},
	}

	handler := requireAuth(cfg.WebUI.APIToken, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"short", "***"},
		{"sk-abc123xyz789long", "sk-a***long"},
		{"12345678", "***"},
		{"123456789012", "1234***9012"},
	}

	for _, tt := range tests {
		result := maskSecret(tt.input)
		if result != tt.expected {
			t.Errorf("maskSecret(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMaskSecret_SkipsShortStrings(t *testing.T) {
	result := maskSecret("abc")
	if result != "***" {
		t.Errorf("maskSecret('abc') = %q, want '***'", result)
	}
}

func TestGenerateAPIToken(t *testing.T) {
	token := generateAPIToken()
	if len(token) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("token length = %d, want 64", len(token))
	}
	// Generate another and verify uniqueness
	token2 := generateAPIToken()
	if token == token2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestHandleHealth_NoAuth(t *testing.T) {
	cfg = &config.Config{
		WebUI: config.WebUIConfig{APIToken: "secret"},
	}

	// Health endpoint should NOT require auth
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
}

func TestHandleSettings_MasksSecrets(t *testing.T) {
	cfg = &config.Config{
		Agent: config.AgentConfig{Name: "test"},
		Providers: []config.ProviderConfig{
			{Name: "openai", APIKey: "sk-supersecret123456789"},
		},
		Channels: config.ChannelsConfig{
			Telegram: config.TelegramConfig{BotToken: "bot:token123456789"},
			Discord:  config.DiscordConfig{BotToken: "discord:token987654321"},
		},
		WebUI: config.WebUIConfig{APIToken: "apikey-verysecret"},
	}

	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()
	handleSettings(w, req)

	var resp config.Config
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Check that secrets are masked
	if resp.Providers[0].APIKey == "sk-supersecret123456789" {
		t.Error("Provider API key should be masked")
	}
	if !strings.Contains(resp.Providers[0].APIKey, "***") {
		t.Errorf("Provider API key should contain ***, got: %s", resp.Providers[0].APIKey)
	}
	if resp.Channels.Telegram.BotToken == "bot:token123456789" {
		t.Error("Telegram bot token should be masked")
	}
	if !strings.Contains(resp.Channels.Telegram.BotToken, "***") {
		t.Errorf("Telegram bot token should contain ***, got: %s", resp.Channels.Telegram.BotToken)
	}
	if resp.Channels.Discord.BotToken == "discord:token987654321" {
		t.Error("Discord bot token should be masked")
	}
	if resp.WebUI.APIToken == "apikey-verysecret" {
		t.Error("WebUI API token should be masked")
	}
}

func TestHandleChannels_MasksTokens(t *testing.T) {
	cfg = &config.Config{
		Channels: config.ChannelsConfig{
			Telegram: config.TelegramConfig{
				Enabled:  true,
				BotToken: "secret-telegram-token-123456",
			},
			Discord: config.DiscordConfig{
				Enabled:  true,
				BotToken: "secret-discord-token-654321",
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/channels", nil)
	w := httptest.NewRecorder()
	handleChannels(w, req)

	var channels []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &channels)

	// Find Telegram channel
	for _, ch := range channels {
		if ch["name"] == "Telegram" {
			token, _ := ch["bot_token"].(string)
			if token == "secret-telegram-token-123456" {
				t.Error("Telegram bot token should be masked in channels response")
			}
			if !strings.Contains(token, "***") {
				t.Errorf("Telegram bot token should contain ***, got: %s", token)
			}
		}
		if ch["name"] == "Discord" {
			token, _ := ch["bot_token"].(string)
			if token == "secret-discord-token-654321" {
				t.Error("Discord bot token should be masked in channels response")
			}
		}
	}
}
