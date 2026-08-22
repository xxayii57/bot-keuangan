package main

import (
	"fmt"
)

// CommandPalette shows a filtered list of commands.
type CommandPalette struct {
	registry *CommandRegistry
	width    int
}

// NewCommandPalette creates a palette bound to the given registry.
func NewCommandPalette(reg *CommandRegistry, width int) *CommandPalette {
	return &CommandPalette{registry: reg, width: width}
}

// Show displays the palette, optionally filtered by a prefix.
func (p *CommandPalette) Show(filter string) {
	matches := p.registry.Match(filter)

	fmt.Print("\033[2K") // clear line

	if filter == "" {
		fmt.Printf("\r%s┌─ Commands ─%s\n", colorDim, colorReset)
	} else {
		fmt.Printf("\r%s┌─ Commands ───────────────────────%s\n", colorDim, colorReset)
	}

	if filter != "" {
		fmt.Printf("%s│%s %s%s%s\n", colorDim, colorReset, colorCyan, filter, colorReset)
		fmt.Printf("%s├──────────────────────────────────%s\n", colorDim, colorReset)
	}

	if len(matches) == 0 {
		fmt.Printf("%s│%s   (no matching commands)%s\n", colorDim, colorReset, colorReset)
	} else {
		for _, c := range matches {
			line := fmt.Sprintf("  %-14s%s", c.Name, c.Desc)
			if p.width > 0 && len(line) > p.width-4 {
				line = line[:p.width-4]
			}
			fmt.Printf("%s│%s %s%s\n", colorDim, colorReset, line, colorReset)
		}
	}

	fmt.Printf("%s└──────────────────────────────────%s\n", colorDim, colorReset)
}

// ShowAll displays all available commands.
func (p *CommandPalette) ShowAll() {
	p.Show("")
}
