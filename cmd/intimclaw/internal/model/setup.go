package model

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// providerPreset is a well-known OpenAI-compatible endpoint offered by the
// interactive setup wizard.
type providerPreset struct {
	name    string
	apiBase string
	noKey   bool   // endpoint works without an API key
	keyHint string // where to obtain an API key
}

var providerPresets = []providerPreset{
	{name: "OpenRouter", apiBase: "https://openrouter.ai/api/v1", keyHint: "https://openrouter.ai/keys"},
	{name: "OpenAI", apiBase: "https://api.openai.com/v1", keyHint: "https://platform.openai.com/api-keys"},
	{name: "DeepSeek", apiBase: "https://api.deepseek.com/v1", keyHint: "https://platform.deepseek.com/api_keys"},
	{name: "Moonshot/Kimi", apiBase: "https://api.moonshot.cn/v1", keyHint: "https://platform.moonshot.cn/console/api-keys"},
	{name: "Groq", apiBase: "https://api.groq.com/openai/v1", keyHint: "https://console.groq.com/keys"},
	{name: "Ollama (local)", apiBase: "http://localhost:11434/v1", noKey: true},
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Guided setup: choose provider, paste API key, pick a model",
		Long: `Interactive guided setup for your default AI model.

The wizard walks you through:
  1. Choosing a provider (or entering a custom OpenAI-compatible URL)
  2. Pasting your API key (input is hidden)
  3. Picking a model from the live list returned by the endpoint

The chosen model is saved and set as the default in your config.

This wizard also runs automatically at the end of 'intimclaw onboard'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RunSetupWizard(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}

// RunSetupWizard runs the interactive provider/key/model picker and saves the
// result as the default model. It is shared by 'intimclaw model setup' and
// the tail end of 'intimclaw onboard'.
func RunSetupWizard(stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)

	preset, err := choosePreset(scanner, stdout)
	if err != nil {
		return err
	}

	apiKey := "local" // placeholder accepted by Ollama and most local servers
	if !preset.noKey {
		fmt.Fprintf(stdout, "\nAPI key for %s", preset.name)
		if preset.keyHint != "" {
			fmt.Fprintf(stdout, " (%s)", preset.keyHint)
		}
		apiKey, err = readSecret(stdin, scanner, stdout)
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(stdout, "\nUsing %s without an API key.\n", preset.name)
	}

	fmt.Fprintf(stdout, "\nFetching models from %s ...\n", preset.apiBase)
	entries, err := fetchOpenAIModels(preset.apiBase, apiKey)
	if err != nil {
		return fmt.Errorf("fetch models: %w\n\nCheck your API key / network, or retry with:\n  intimclaw model setup", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no models returned by %s", preset.apiBase)
	}

	selected, err := pickWithScanner(scanner, stdout, entries)
	if err != nil {
		return err
	}

	alias := aliasFor(preset)
	return upsertModelDefault(preset.apiBase, apiKey, alias, selected, stdout)
}
func choosePreset(scanner *bufio.Scanner, stdout io.Writer) (providerPreset, error) {
	fmt.Fprintln(stdout, "\nChoose your AI provider:")
	for i, p := range providerPresets {
		line := fmt.Sprintf("  %d) %-16s %s", i+1, p.name, p.apiBase)
		if p.noKey {
			line += "   (no API key needed)"
		}
		fmt.Fprintln(stdout, line)
	}
	custom := len(providerPresets) + 1
	fmt.Fprintf(stdout, "  %d) %-16s any other OpenAI-compatible endpoint\n", custom, "Custom")

	for {
		fmt.Fprintf(stdout, "Pick a provider (1-%d): ", custom)
		if !scanner.Scan() {
			return providerPreset{}, fmt.Errorf("no selection provided")
		}
		text := strings.TrimSpace(scanner.Text())
		idx, err := strconv.Atoi(text)
		if err != nil || idx < 1 || idx > custom {
			fmt.Fprintf(stdout, "Enter a number between 1 and %d.\n", custom)
			continue
		}
		if idx < custom {
			return providerPresets[idx-1], nil
		}
		fmt.Fprint(stdout, "API base URL (e.g. https://host.example/v1): ")
		if !scanner.Scan() {
			return providerPreset{}, fmt.Errorf("no API base provided")
		}
		base := strings.TrimSpace(scanner.Text())
		if base == "" {
			fmt.Println("URL must not be empty.")
			continue
		}
		return providerPreset{name: "Custom", apiBase: base}, nil
	}
}

// readSecret reads a secret with echo disabled when stdin is an interactive
// TTY; it falls back to the shared scanner when stdin is piped (tests,
// scripted installs). The shared scanner must be used so buffered input from
// earlier prompts is not lost.
func readSecret(stdin io.Reader, scanner *bufio.Scanner, stdout io.Writer) (string, error) {
	if f, ok := stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(stdout, ": ")
		secret, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(stdout)
		if err != nil {
			return "", fmt.Errorf("reading api key: %w", err)
		}
		text := strings.TrimSpace(string(secret))
		if text == "" {
			return "", fmt.Errorf("api key must not be empty")
		}
		return text, nil
	}
	fmt.Fprintln(stdout, ":")
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		return text, nil
	}
	return "", fmt.Errorf("no api key provided")
}

func aliasFor(p providerPreset) string {
	name := strings.ToLower(p.name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	if name == "" || name == "custom" {
		return defaultAliasName
	}
	return name
}
