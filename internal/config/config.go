// intimclaw.ic — Built by xxayii — IntimClaw Config v0.1.0
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Agent     AgentConfig     `toml:"agent"`
	Providers []ProviderConfig `toml:"providers"`
	Channels  ChannelsConfig  `toml:"channels"`
	MCP       MCPConfig       `toml:"mcp"`
	Skills    SkillsConfig    `toml:"skills"`
	Memory    MemoryConfig    `toml:"memory"`
	Security  SecurityConfig  `toml:"security"`
	WebUI     WebUIConfig     `toml:"webui"`
	Cron      CronConfig      `toml:"cron"`
}

type AgentConfig struct {
	Name     string `toml:"name"`
	Version  string `toml:"version"`
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	Persona  string `toml:"persona"`
}

type ProviderConfig struct {
	Name    string            `toml:"name"`
	Type    string            `toml:"type"`
	BaseURL string            `toml:"base_url"`
	APIKey  string            `toml:"api_key"`
	Models  []string          `toml:"models"`
	Headers map[string]string `toml:"headers,omitempty"`
}

type ChannelsConfig struct {
	Telegram TelegramConfig `toml:"telegram"`
	Discord  DiscordConfig  `toml:"discord"`
	WebChat  WebChatConfig  `toml:"webchat"`
	WhatsApp WhatsAppConfig `toml:"whatsapp"`
	Email    EmailConfig    `toml:"email"`
}

type TelegramConfig struct {
	Enabled     bool   `toml:"enabled"`
	BotToken    string `toml:"bot_token"`
	OwnerID     int64  `toml:"owner_id"`
	MentionOnly bool   `toml:"mention_only"`
}

type DiscordConfig struct {
	Enabled  bool   `toml:"enabled"`
	BotToken string `toml:"bot_token"`
}

type WebChatConfig struct {
	Enabled bool   `toml:"enabled"`
	Port    int    `toml:"port"`
	Host    string `toml:"host"`
}

type WhatsAppConfig struct {
	Enabled bool `toml:"enabled"`
}

type EmailConfig struct {
	Enabled  bool   `toml:"enabled"`
	ImapHost string `toml:"imap_host"`
	ImapUser string `toml:"imap_user"`
	ImapPass string `toml:"imap_pass"`
	SmtpHost string `toml:"smtp_host"`
	SmtpPort int    `toml:"smtp_port"`
}

type MCPConfig struct {
	Enabled bool        `toml:"enabled"`
	Servers []MCPServer `toml:"servers"`
}

type MCPServer struct {
	Name    string   `toml:"name"`
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

type SkillsConfig struct {
	Enabled        bool     `toml:"enabled"`
	Directories    []string `toml:"directories"`
	DisabledSkills []string `toml:"disabled_skills,omitempty"`
}

type MemoryConfig struct {
	Backend        string `toml:"backend"`
	SemanticSearch bool   `toml:"semantic_search"`
	DecayDays      int    `toml:"decay_days"`
}

type SecurityConfig struct {
	RiskProfile    string   `toml:"risk_profile"`
	Sandbox        bool     `toml:"sandbox"`
	ExcludedTools  []string `toml:"excluded_tools"`
	ForbiddenPaths []string `toml:"forbidden_paths"`
}

type WebUIConfig struct {
	Enabled  bool   `toml:"enabled"`
	Port     int    `toml:"port"`
	Host     string `toml:"host"`
	Theme    string `toml:"theme"`
	APIToken string `toml:"api_token,omitempty"`
}

type CronConfig struct {
	Enabled bool `toml:"enabled"`
	MaxJobs int  `toml:"max_jobs"`
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}
	return filepath.Join(home, ".intimclaw", "config.toml")
}

func GetMemoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}
	return filepath.Join(home, ".intimclaw", "memory.db")
}

func GetSkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}
	return filepath.Join(home, ".intimclaw", "skills")
}

func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Name:     "intimclaw",
			Version:  "0.1.0",
			Provider: "openai",
			Model:    "gpt-4o-mini",
			Persona:  "default",
		},
		Providers: []ProviderConfig{
			{
				Name:    "openai",
				Type:    "openai-compatible",
				BaseURL: "https://api.openai.com/v1",
				Models: []string{
					"gpt-4o-mini",
					"gpt-4o",
				},
			},
			{
				Name:    "anthropic",
				Type:    "anthropic",
				BaseURL: "https://api.anthropic.com/v1",
				Models: []string{
					"claude-3-5-sonnet-latest",
				},
			},
			{
				Name:    "ollama",
				Type:    "openai-compatible",
				BaseURL: "http://localhost:11434/v1",
				Models: []string{
					"llama3.2",
					"codellama",
				},
			},
		},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{Enabled: false},
			WebChat:  WebChatConfig{Enabled: true, Port: 18080, Host: "127.0.0.1"},
		},
		Skills: SkillsConfig{
			Enabled:     true,
			Directories: []string{GetSkillsDir()},
		},
		Memory: MemoryConfig{
			Backend:        "sqlite",
			SemanticSearch: true,
			DecayDays:      30,
		},
		Security: SecurityConfig{
			RiskProfile:    "default",
			ExcludedTools:  []string{"rm", "mkfs", "dd", "shutdown", "poweroff"},
			ForbiddenPaths: []string{".ssh", ".gnupg", ".aws"},
		},
		WebUI: WebUIConfig{
			Enabled: true,
			Port:    18080,
			Host:    "127.0.0.1",
			Theme:   "intimclaw",
		},
		Cron: CronConfig{
			Enabled: true,
			MaxJobs: 50,
		},
	}
}

func Load() (*Config, error) {
	path := GetConfigPath()
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		os.MkdirAll(filepath.Dir(path), 0755)
		out, _ := os.Create(path)
		enc := toml.NewEncoder(out)
		enc.Encode(cfg)
		out.Close()
		return cfg, nil
	}

	_, err = toml.Decode(string(data), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path := GetConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := toml.NewEncoder(out)
	return enc.Encode(cfg)
}
