package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xxayii/intimclaw/internal/agent"
	"github.com/xxayii/intimclaw/internal/config"
)

const VERSION = "0.1.0"

// ANSI colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// chatSession holds the state for an interactive chat session.
type chatSession struct {
	Name     string
	Messages []agent.Message
	Started  time.Time
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "agent" {
		runAgent(args)
		return
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Printf("IntimClaw v%s\n", VERSION)
	case "web":
		port := 18080
		host := "127.0.0.1"
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--port":
				if i+1 < len(args) {
					fmt.Sscanf(args[i+1], "%d", &port)
					i++
				}
			case "--hostname":
				if i+1 < len(args) {
					host = args[i+1]
					i++
				}
			case "--public":
				host = "0.0.0.0"
			}
		}
		startWebUI(port, host)
	case "gateway":
		if cfg.Channels.Telegram.Enabled && cfg.Channels.Telegram.BotToken != "" {
			fmt.Println("[intimclaw] Starting Telegram bot...")
			a := agent.NewFromConfig(cfg)
			bot := agent.NewTelegramBot(cfg.Channels.Telegram.BotToken, a)
			go bot.Start()
		}
		if cfg.Channels.Discord.Enabled && cfg.Channels.Discord.BotToken != "" {
			fmt.Println("[intimclaw] Starting Discord bot...")
			a := agent.NewFromConfig(cfg)
			dbot := agent.NewDiscordBot(cfg.Channels.Discord.BotToken, a)
			go dbot.Start()
		}
		fmt.Println("[intimclaw] Gateway mode")
		fmt.Println("[intimclaw] Gateway mode — coming soon")
	case "daemon":
		fmt.Println("[intimclaw] Daemon mode — coming soon")
	case "config":
		handleConfig(args[1:])
	case "status":
		printStatus()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\nRun 'intimclaw help' for usage.\n", args[0])
	}
}

func printSplash() {
	fmt.Println()
	fmt.Println(colorCyan + `     ____                          ____ _
    / ___|_      ____ _ _   _  __ _  __ _  ___ _ __ ___
   | |   \ \ /\ / / _  | | | |/ _  |/ _  |/ _ \ '_ ` + "`" + ` __
   | |___ \ V  V / (_| | |_| | (_| | (_| |  __/ | | | |
    \____| \_/\_/ \__,_|\__, |\__,_|\__, |\___|_| |_| |
                        |___/      |___/             ` + colorReset)
	fmt.Println()
	fmt.Printf("  %sIntimClaw%s %sv%s  %sAI Agent System%s\n", colorBold, colorReset, colorDim, VERSION, colorReset, colorReset)
	fmt.Printf("  %s● Online%s\n", colorGreen, colorReset)
	fmt.Println()
}

func printStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[error] config: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
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
}

func runAgent(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[error] config: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Parse flags
	message := ""
	model := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "--model":
			if i+1 < len(args) {
				model = args[i+1]
				i++
			}
		default:
			if message == "" {
				message = args[i]
			}
		}
	}

	if model != "" {
		cfg.Agent.Model = model
	}

	a := agent.New(cfg)

	if message != "" {
		// One-shot mode — no splash
		resp, err := a.Run(message)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%serror: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}
		fmt.Println(resp)
		return
	}

	// Interactive mode
	printSplash()
	printSessionHeader(cfg)

	// Graceful Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
		os.Exit(0)
	}()

	reader := bufio.NewReader(os.Stdin)
	spinner := NewBrandSpinner("intimclaw", 120*time.Millisecond)
	sess := &chatSession{Name: "default", Started: time.Now()}

	for {
		printPrompt()
		input, err := reader.ReadString('\n')
		if err != nil {
			// EOF or read error — graceful exit
			fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
			return
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Slash commands
		if strings.HasPrefix(input, "/") {
			if handleSlashCommand(input, cfg, a, sess) {
				return // /exit or /quit
			}
			continue
		}

		// Agent request
		sess.Messages = append(sess.Messages, agent.Message{Role: "user", Content: input})

		spinner.Start()
		resp, err := a.Run(input)
		spinner.Stop()

		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s✗ Request failed%s\n", colorRed, colorReset)
			fmt.Fprintf(os.Stderr, "  Provider: %s\n", cfg.Agent.Provider)
			fmt.Fprintf(os.Stderr, "  Model: %s\n", cfg.Agent.Model)
			fmt.Fprintf(os.Stderr, "  Reason: %v\n", err)
			continue
		}

		sess.Messages = append(sess.Messages, agent.Message{Role: "assistant", Content: resp})

		// Print response
		fmt.Printf("\n%s\n\n", resp)
	}
}

func printSessionHeader(cfg *config.Config) {
	provider := cfg.Agent.Provider
	if provider == "" {
		provider = "none"
	}
	model := cfg.Agent.Model
	if model == "" {
		model = "none"
	}

	fmt.Printf("  Provider : %s\n", provider)
	fmt.Printf("  Model    : %s\n", model)
	fmt.Printf("  Session  : default\n")
	fmt.Println()
}

func printPrompt() {
	fmt.Printf("%s┌─ You ─────────────────────────────────────%s\n", colorDim, colorReset)
	fmt.Printf("%s│ > %s", colorDim, colorReset)
}

// ─── Slash Commands ──────────────────────────────────────────────

// handleSlashCommand processes a slash command. Returns true if the
// session should exit (e.g. /exit, /quit).
func handleSlashCommand(input string, cfg *config.Config, a *agent.Agent, sess *chatSession) bool {
	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help", "/h":
		printSlashHelp()
	case "/model":
		printModelInfo(cfg)
	case "/models":
		printAvailableModels(cfg)
	case "/status":
		printStatus()
	case "/context":
		printContext(sess)
	case "/session":
		printSessionInfo(sess)
	case "/new":
		sess.Messages = nil
		sess.Started = time.Now()
		fmt.Printf("\n  %s✓ Session reset. History cleared.%s\n\n", colorGreen, colorReset)
	case "/clear":
		fmt.Print("\033[2J\033[H")
	case "/exit", "/quit":
		fmt.Printf("%sBye!%s\n", colorDim, colorReset)
		return true
	default:
		fmt.Printf("\n  Unknown command: %s\n", cmd)
		fmt.Printf("  Type /help for available commands.\n\n")
	}
	return false
}

func printSlashHelp() {
	fmt.Println()
	fmt.Println("  ┌─ Commands ─────────────────────────────────┐")
	fmt.Println("  │ /help       Show available commands         │")
	fmt.Println("  │ /model      Current provider & model        │")
	fmt.Println("  │ /models     Available models from config    │")
	fmt.Println("  │ /status     System status                   │")
	fmt.Println("  │ /context    Context window information      │")
	fmt.Println("  │ /session    Current session info            │")
	fmt.Println("  │ /new        Start new session               │")
	fmt.Println("  │ /clear      Clear terminal                  │")
	fmt.Println("  │ /exit       Exit                            │")
	fmt.Println("  └─────────────────────────────────────────────┘")
	fmt.Println()
}

func printModelInfo(cfg *config.Config) {
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
}

func printAvailableModels(cfg *config.Config) {
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
}

func printContext(sess *chatSession) {
	totalChars := 0
	for _, m := range sess.Messages {
		totalChars += len(m.Content)
	}
	// Rough estimate: ~4 chars per token
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
}

func printSessionInfo(sess *chatSession) {
	elapsed := time.Since(sess.Started)
	fmt.Println()
	fmt.Printf("  Session    : %s\n", sess.Name)
	fmt.Printf("  Messages   : %d\n", len(sess.Messages))
	fmt.Printf("  Started    : %s\n", sess.Started.Format("15:04:05"))
	fmt.Printf("  Duration   : %s\n", formatDuration(elapsed))
	fmt.Println()
}

func handleConfig(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: intimclaw config [list|get|set]")
		return
	}
	switch args[0] {
	case "list":
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Printf("Model: %s/%s\n", cfg.Agent.Provider, cfg.Agent.Model)
		fmt.Printf("Persona: %s\n", cfg.Agent.Persona)
		fmt.Printf("WebUI port: %d\n", cfg.WebUI.Port)
		fmt.Printf("Channels: telegram=%v discord=%v\n",
			cfg.Channels.Telegram.Enabled, cfg.Channels.Discord.Enabled)
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: intimclaw config get <key>")
			return
		}
		fmt.Printf("config get %s — coming soon\n", args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: intimclaw config set <key> <value>")
			return
		}
		fmt.Printf("config set %s=%s — coming soon\n", args[1], args[2])
	default:
		fmt.Println("Usage: intimclaw config [list|get|set]")
	}
}

func printHelp() {
	fmt.Print(`
IntimClaw — AI Agent System

Usage:
  intimclaw                     Start interactive agent (CLI)
  intimclaw agent               Start interactive agent
  intimclaw agent -m "prompt"   One-shot message
  intimclaw web                 Start WebUI
  intimclaw gateway             Start gateway
  intimclaw daemon              Start background daemon
  intimclaw config list         Show config
  intimclaw status              Show status
  intimclaw help                Show this help
  intimclaw version             Show version

Examples:
  intimclaw -m "hello"
  intimclaw agent --model gpt-4o -m "cek RAM"
  intimclaw web --port 18080 --hostname 0.0.0.0

Interactive mode commands:
  /help     Show available commands
  /model    Current provider & model
  /models   Available models
  /status   System status
  /context  Context window info
  /session  Session info
  /new      New session
  /clear    Clear terminal
  /exit     Exit
`)
}
