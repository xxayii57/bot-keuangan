package cliui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PrintOnboardComplete prints the post-onboard “ready” message and next steps.
func PrintOnboardComplete(logo string, encrypt bool, configPath string) {
	if !UseFancyLayout() {
		printOnboardPlain(logo, encrypt, configPath)
		return
	}
	printOnboardFancy(logo, encrypt, configPath)
}

func printOnboardPlain(logo string, encrypt bool, configPath string) {
	fmt.Printf("\n%s intimclaw is ready!\n", logo)
	fmt.Println("\nNext steps:")
	if encrypt {
		fmt.Println("  1. Set your encryption passphrase before starting intimclaw:")
		fmt.Println("       export INTIMCLAW_KEY_PASSPHRASE=<your-passphrase>   # Linux/macOS")
		fmt.Println("       set INTIMCLAW_KEY_PASSPHRASE=<your-passphrase>      # Windows cmd")
		fmt.Println("")
		fmt.Println("  2. Configure your AI model with the guided wizard:")
		fmt.Println("       intimclaw model setup")
	} else {
		fmt.Println("  1. Configure your AI model with the guided wizard:")
		fmt.Println("       intimclaw model setup")
	}
	fmt.Println("")
	fmt.Println("     It walks you through choosing a provider, pasting your API key,")
	fmt.Println("     and picking a model interactively.")
	fmt.Println("")
	fmt.Println("     Recommended: OpenRouter (https://openrouter.ai/keys, access 100+")
	fmt.Println("     models) or Ollama (https://ollama.com, local & free, no key).")
	fmt.Println("")
	if encrypt {
		fmt.Println("  3. Chat: intimclaw agent")
	} else {
		fmt.Println("  2. Chat: intimclaw agent")
	}
}

func printOnboardFancy(logo string, encrypt bool, configPath string) {
	inner := InnerWidth()
	box := borderStyle().MaxWidth(inner + 8)

	ready := titleBarStyle().Render(logo+" intimclaw is ready!") + "\n"
	fmt.Println()
	fmt.Println(box.Width(inner).Render(strings.TrimSpace(ready)))
	fmt.Println()

	steps := buildOnboardingSteps(encrypt, configPath)
	rec := recommendedBlock()
	chat := chatStep(encrypt)

	if UseColumnLayout() {
		leftW := min(inner/2-2, 52)
		rightW := inner - leftW - 4
		if rightW < 36 {
			rightW = 36
		}
		leftBlock := borderStyle().MaxWidth(leftW + 8).Width(leftW).
			Render(titleBarStyle().Render("Next steps") + "\n\n" + bodyStyle().Width(leftW).Render(steps))
		rightBlock := borderStyle().MaxWidth(rightW + 8).Width(rightW).
			Render(mutedStyle().Bold(true).Render("Recommended") + "\n\n" + bodyStyle().Width(rightW).Render(rec))
		gap := strings.Repeat(" ", 2)
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gap, rightBlock))
		fmt.Println()
		full := borderStyle().Width(inner).Render(bodyStyle().Width(inner - 4).Render(chat))
		fmt.Println(full)
		return
	}

	// Same order as plain output: numbered steps → recommended → chat line.
	next := titleBarStyle().Render("Next steps") + "\n\n" +
		bodyStyle().Width(inner-4).Render(steps+"\n\n"+rec+"\n\n"+chat)
	fmt.Println(borderStyle().Width(inner).Render(next))
}

func buildOnboardingSteps(encrypt bool, configPath string) string {
	var b strings.Builder
	if encrypt {
		b.WriteString("1. Set your encryption passphrase before starting intimclaw:\n")
		b.WriteString("   export INTIMCLAW_KEY_PASSPHRASE=<your-passphrase>   # Linux/macOS\n")
		b.WriteString("   set INTIMCLAW_KEY_PASSPHRASE=<your-passphrase>      # Windows cmd\n\n")
		b.WriteString("2. Configure your AI model:\n   intimclaw model setup\n")
	} else {
		b.WriteString("1. Configure your AI model:\n   intimclaw model setup\n")
	}
	return b.String()
}

func recommendedBlock() string {
	return "The wizard asks for a provider, an API key, and lets you\n" +
		"pick a model from the live list.\n\n" +
		"• OpenRouter: https://openrouter.ai/keys\n  (access 100+ models)\n\n" +
		"• Ollama: https://ollama.com\n  (local, free, no key needed)"
}

func chatStep(encrypt bool) string {
	if encrypt {
		return "3. Chat:\n   intimclaw agent"
	}
	return "2. Chat:\n   intimclaw agent"
}
