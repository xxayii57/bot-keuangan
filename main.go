package main

import (
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
	runInteractive(cfg, a)
}

func runInteractive(cfg *config.Config, a *agent.Agent) {
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

	registry := NewCommandRegistry()
	palette := NewCommandPalette(registry, 70)
	spinner := NewBrandSpinner("intimclaw", 120*time.Millisecond)
	sess := &chatSession{Name: "default", Started: time.Now()}
	editor := NewInputEditor(os.Stdin)
	editor.EnableRawMode()
	defer editor.DisableRawMode()

	printFooter()

	for {
		input, exited := editor.ReadInput()
		if exited {
			fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
			return
		}
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Slash commands
		if strings.HasPrefix(input, "/") {
			if input == "/" {
				// Show full palette, then read command
				palette.ShowAll()
				fmt.Printf("%s> %s", colorDim, colorReset)
				cmdInput, exited := editor.ReadInput()
				if exited {
					fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
					return
				}
				cmdInput = strings.TrimSpace(cmdInput)
				fmt.Printf("%s└───────────────────────────────────────────────%s\n", colorDim, colorReset)
				if cmdInput == "" {
					fmt.Println()
					continue
				}
				if registry.Execute(cmdInput, cfg, sess) {
					return
				}
				fmt.Println()
				continue
			}
			// Check if it's a valid command
			matches := registry.Match(input)
			if len(matches) == 1 {
				// Exact match — execute
				if registry.Execute(input, cfg, sess) {
					return
				}
				fmt.Println()
			} else if len(matches) > 1 {
				// Multiple matches — show palette
				palette.Show(input)
				fmt.Printf("%s> %s", colorDim, colorReset)
				cmdInput, exited := editor.ReadInput()
				if exited {
					fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
					return
				}
				cmdInput = strings.TrimSpace(cmdInput)
				fmt.Printf("%s└───────────────────────────────────────────────%s\n", colorDim, colorReset)
				if cmdInput == "" {
					fmt.Println()
					continue
				}
				if registry.Execute(cmdInput, cfg, sess) {
					return
				}
				fmt.Println()
			} else {
				// No matches — show help
				registry.Execute(input, cfg, sess)
				fmt.Println()
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
			fmt.Println()
			printFooter()
			continue
		}

		sess.Messages = append(sess.Messages, agent.Message{Role: "assistant", Content: resp})

		fmt.Printf("\n%s\n\n", resp)
		printFooter()
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

func printFooter() {
	fmt.Printf("%s────────────────────────────────────────────────────%s\n", colorDim, colorReset)
	fmt.Printf("%s  ctrl+p commands%s\n", colorDim, colorReset)
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
