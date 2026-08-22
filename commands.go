package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/xxayii/intimclaw/internal/config"
)

// CommandDef defines a single slash command.
type CommandDef struct {
	Name        string
	Aliases     []string
	Desc        string
	Handler     func(cfg *config.Config, sess *chatSession) (exit bool, err error)
}

// CommandRegistry is the single source of truth for all slash commands.
type CommandRegistry struct {
	commands []CommandDef
}

// NewCommandRegistry creates and populates the registry.
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{}
	r.commands = []CommandDef{
		{Name: "/help", Aliases: []string{"/h"}, Desc: "Show available commands", Handler: cmdHelp},
		{Name: "/model", Desc: "Current provider & model", Handler: cmdModel},
		{Name: "/models", Desc: "Available models from config", Handler: cmdModels},
		{Name: "/status", Desc: "System status", Handler: cmdStatus},
		{Name: "/context", Desc: "Context window information", Handler: cmdContext},
		{Name: "/session", Desc: "Current session info", Handler: cmdSession},
		{Name: "/new", Desc: "Start new session", Handler: cmdNew},
		{Name: "/clear", Desc: "Clear terminal screen", Handler: cmdClear},
		{Name: "/exit", Desc: "Exit IntimClaw", Handler: cmdExit},
		{Name: "/quit", Desc: "Exit IntimClaw", Handler: cmdExit},
	}
	return r
}

// Match finds commands matching the given input prefix.
func (r *CommandRegistry) Match(input string) []CommandDef {
	lower := strings.ToLower(input)
	var matches []CommandDef
	for _, c := range r.commands {
		if strings.HasPrefix(c.Name, lower) || strings.HasPrefix(lower, c.Name) {
			matches = append(matches, c)
			continue
		}
		for _, alias := range c.Aliases {
			if strings.HasPrefix(alias, lower) {
				matches = append(matches, c)
				break
			}
		}
	}
	return matches
}

// Execute runs a command by name. Returns true if session should exit.
func (r *CommandRegistry) Execute(input string, cfg *config.Config, sess *chatSession) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}
	cmd := strings.ToLower(parts[0])
	for _, c := range r.commands {
		if cmd == c.Name {
			exit, _ := c.Handler(cfg, sess)
			return exit
		}
		for _, alias := range c.Aliases {
			if cmd == alias {
				exit, _ := c.Handler(cfg, sess)
				return exit
			}
		}
	}
	fmt.Printf("\n  Unknown command: %s\n", cmd)
	fmt.Printf("  Type /help for available commands.\n\n")
	return false
}

// ─── Command Handlers ─────────────────────────────────────────

func cmdHelp(cfg *config.Config, sess *chatSession) (bool, error) {
	fmt.Println()
	fmt.Println("  ┌─ Commands ─────────────────────────────────────┐")
	fmt.Println("  │ /help       Show available commands            │")
	fmt.Println("  │ /model      Current provider & model           │")
	fmt.Println("  │ /models     Available models from config       │")
	fmt.Println("  │ /status     System status                      │")
	fmt.Println("  │ /context    Context window information         │")
	fmt.Println("  │ /session    Current session info               │")
	fmt.Println("  │ /new        Start new session                  │")
	fmt.Println("  │ /clear      Clear terminal screen              │")
	fmt.Println("  │ /exit       Exit IntimClaw                     │")
	fmt.Println("  └────────────────────────────────────────────────┘")
	fmt.Println()
	return false, nil
}

func cmdModel(cfg *config.Config, sess *chatSession) (bool, error) {
	provider := cfg.Agent.Provider
	if provider == "" {
		provider = "none"
	}
	model := cfg.Agent.Model
	if model == "" {
		model = "none"
	}
	fmt.Println()
	fmt.Printf("  Provider : %s\n", provider)
	fmt.Printf("  Model    : %s\n", model)
	fmt.Println()
	return false, nil
}

func cmdModels(cfg *config.Config, sess *chatSession) (bool, error) {
	fmt.Println()
	fmt.Println("  Available models:")
	for _, p := range cfg.Providers {
		for _, m := range p.Models {
			marker := " "
			if p.Name == cfg.Agent.Provider && m == cfg.Agent.Model {
				marker = "*"
			}
			fmt.Printf("  %s %s/%s\n", marker, p.Name, m)
		}
	}
	fmt.Println("  (* = active)")
	fmt.Println()
	return false, nil
}

func cmdStatus(cfg *config.Config, sess *chatSession) (bool, error) {
	provider := cfg.Agent.Provider
	if provider == "" {
		provider = "none"
	}
	model := cfg.Agent.Model
	if model == "" {
		model = "none"
	}

	fmt.Println()
	fmt.Println("  IntimClaw Status")
	fmt.Println("  ────────────────────────────")
	fmt.Printf("  Version    : v%s\n", VERSION)
	fmt.Printf("  Provider   : %s\n", provider)
	fmt.Printf("  Model      : %s\n", model)
	fmt.Printf("  Status     : %sOnline%s\n", colorGreen, colorReset)
	fmt.Printf("  WebUI      : %s:%d\n", cfg.WebUI.Host, cfg.WebUI.Port)
	fmt.Println("  ────────────────────────────")
	fmt.Println()
	return false, nil
}

func cmdContext(cfg *config.Config, sess *chatSession) (bool, error) {
	totalChars := 0
	for _, m := range sess.Messages {
		totalChars += len(m.Content)
	}
	estimatedTokens := totalChars / 4
	messages := len(sess.Messages)

	fmt.Println()
	fmt.Println("  Context Window")
	fmt.Println("  ────────────────────────────")
	fmt.Printf("  Messages   : %d\n", messages)
	fmt.Printf("  Characters : %d\n", totalChars)
	fmt.Printf("  Est. Tokens: ~%d\n", estimatedTokens)
	fmt.Println("  (estimated — actual usage depends on tokenizer)")
	fmt.Println("  ────────────────────────────")
	fmt.Println()
	return false, nil
}

func cmdSession(cfg *config.Config, sess *chatSession) (bool, error) {
	elapsed := time.Since(sess.Started)
	fmt.Println()
	fmt.Printf("  Session    : %s\n", sess.Name)
	fmt.Printf("  Messages   : %d\n", len(sess.Messages))
	fmt.Printf("  Started    : %s\n", sess.Started.Format("15:04:05"))
	fmt.Printf("  Duration   : %s\n", formatDuration(elapsed))
	fmt.Println()
	return false, nil
}

func cmdNew(cfg *config.Config, sess *chatSession) (bool, error) {
	sess.Messages = nil
	sess.Started = time.Now()
	fmt.Printf("\n  %s✓ Session reset. History cleared.%s\n\n", colorGreen, colorReset)
	return false, nil
}

func cmdClear(cfg *config.Config, sess *chatSession) (bool, error) {
	fmt.Print("\033[2J\033[H")
	return false, nil
}

func cmdExit(cfg *config.Config, sess *chatSession) (bool, error) {
	fmt.Printf("%sBye!%s\n", colorDim, colorReset)
	return true, nil
}
