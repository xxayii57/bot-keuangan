package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
		fmt.Printf("IntimClaw v%s\n", VERSION)
		fmt.Println("Status: core engine ready")
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\nRun 'intimclaw help' for usage.\n", args[0])
	}
}

// InputScanner reads lines from stdin.
type InputScanner struct{}

func NewInputScanner() *InputScanner {
	return &InputScanner{}
}

func (s *InputScanner) Scan() (string, error) {
	var input string
	_, err := fmt.Scanln(&input)
	return input, err
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
	fmt.Printf("  %sIntimClaw%s  %sv%s  %sAI Agent System%s\n", colorBold, colorReset, colorDim, VERSION, colorReset, colorReset)
	fmt.Printf("  %sOnline%s\n", colorGreen, colorReset)
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
	} else {
		// Interactive mode — splash + prompt
		printSplash()

		// Graceful Ctrl+C
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
			os.Exit(0)
		}()

		scanner := NewInputScanner()
		for {
			fmt.Printf("%s>%s ", colorCyan, colorReset)
			input, err := scanner.Scan()
			if err != nil {
				fmt.Printf("\n%sBye!%s\n", colorDim, colorReset)
				return
			}
			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}
			if input == "exit" || input == "quit" {
				fmt.Printf("%sBye!%s\n", colorDim, colorReset)
				return
			}

			resp, err := a.Run(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%serror: %v%s\n", colorRed, err, colorReset)
				continue
			}
			fmt.Printf("\n%s\n\n", resp)
		}
	}
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
`)
}
