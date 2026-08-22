package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/xxayii/intimclaw/internal/agent"
	"github.com/xxayii/intimclaw/internal/config"
)

// ─── Styles ───────────────────────────────────────────────────

const (
	colorReset = "\033[0m"
	colorCyan  = "\033[36m"
	colorDim   = "\033[2m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")) // cyan

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	styleRed = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	styleInputBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	styleInputBoxActive = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("6")).
				Padding(0, 1)

	stylePaletteBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	stylePaletteActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	stylePaletteNormal = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	styleFooter = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleResponse = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)

// ─── Model ───────────────────────────────────────────────────

type phase int

const (
	phaseInput phase = iota
	phasePalette
	phaseThinking
	phaseResponse
)

type model struct {
	input       textinput.Model
	cfg         *config.Config
	agent       *agent.Agent
	registry    *CommandRegistry
	session     *chatSession
	spinner     spinner.Model
	phase       phase
	width       int
	height      int
	lastResp    string
	lastErr     string
	startAt     time.Time
	paletteIdx  int
	paletteList []string
}

type responseMsg struct {
	resp string
	err  error
}

func initialModel(cfg *config.Config, a *agent.Agent) model {
	ti := textinput.New()
	ti.Placeholder = "Ketik pesan..."
	ti.Focus()
	ti.CharLimit = 8192
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return model{
		input:    ti,
		cfg:      cfg,
		agent:    a,
		registry: NewCommandRegistry(),
		session:  &chatSession{Name: "default", Started: time.Now()},
		spinner:  sp,
		phase:    phaseInput,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case responseMsg:
		m.lastResp = msg.resp
		m.lastErr = msg.err.Error()
		m.phase = phaseResponse
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// During thinking, ignore most keys except Ctrl+C
	if m.phase == phaseThinking {
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
		return m, nil
	}

	// During palette
	if m.phase == phasePalette {
		return m.handlePaletteKey(msg)
	}

	// During input
	return m.handleInputKey(msg)
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEnter:
		input := strings.TrimSpace(m.input.Value())
		if input == "" {
			return m, nil
		}

		// Slash command
		if strings.HasPrefix(input, "/") {
			m.input.Reset()
			return m.executeCommand(input)
		}

		// Agent request
		m.input.Reset()
		m.phase = phaseThinking
		m.startAt = time.Now()
		return m, tea.Batch(m.spinner.Tick, m.sendRequest(input))

	case tea.KeyTab:
		// Show palette on Tab
		m.showPalette("")
		return m, nil

	default:
		// Let textinput handle the key
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// Check if input now starts with "/" — show palette live
		val := m.input.Value()
		if strings.HasPrefix(val, "/") {
			m.showPalette(val)
		}

		return m, cmd
	}
}

func (m model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Close palette, return to input
		m.phase = phaseInput
		m.input.Reset()
		return m, nil

	case tea.KeyUp:
		if m.paletteIdx > 0 {
			m.paletteIdx--
		}
		return m, nil

	case tea.KeyDown:
		if m.paletteIdx < len(m.paletteList)-1 {
			m.paletteIdx++
		}
		return m, nil

	case tea.KeyEnter:
		if len(m.paletteList) > 0 && m.paletteIdx < len(m.paletteList) {
			selected := m.paletteList[m.paletteIdx]
			m.input.Reset()
			m.phase = phaseInput
			return m.executeCommand(selected)
		}
		m.phase = phaseInput
		return m, nil

	case tea.KeyCtrlC:
		return m, tea.Quit

	default:
		// Typing in palette — filter
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		val := m.input.Value()
		m.updatePaletteFilter(val)
		if len(m.paletteList) == 0 {
			// No matches — close palette
			m.phase = phaseInput
		}
		return m, cmd
	}
}

func (m *model) showPalette(filter string) {
	m.phase = phasePalette
	m.input.Reset()
	m.updatePaletteFilter(filter)
}

func (m *model) updatePaletteFilter(filter string) {
	lower := strings.ToLower(filter)
	m.paletteList = nil
	m.paletteIdx = 0
	for _, c := range m.registry.commands {
		if strings.HasPrefix(c.Name, lower) {
			m.paletteList = append(m.paletteList, c.Name)
		}
	}
}

func (m model) executeCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	// Handle built-in actions that need model state
	switch cmd {
	case "/exit", "/quit":
		return m, tea.Quit
	case "/clear":
		return m, tea.ClearScreen
	case "/new":
		m.session.Messages = nil
		m.session.Started = time.Now()
		m.lastResp = "Session reset. History cleared."
		m.phase = phaseResponse
		return m, nil
	}

	// Execute via registry
	exit := m.registry.Execute(input, m.cfg, m.session)
	if exit {
		return m, tea.Quit
	}

	// If registry printed output, we stay in input mode
	m.phase = phaseInput
	return m, nil
}

func (m model) sendRequest(input string) tea.Cmd {
	return func() tea.Msg {
		m.session.Messages = append(m.session.Messages, agent.Message{Role: "user", Content: input})
		resp, err := m.agent.Run(input)
		if err != nil {
			return responseMsg{err: err}
		}
		m.session.Messages = append(m.session.Messages, agent.Message{Role: "assistant", Content: resp})
		return responseMsg{resp: resp}
	}
}

// ─── View ─────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.viewHeader())
	b.WriteString("\n")

	// Content area
	switch m.phase {
	case phaseThinking:
		b.WriteString(m.viewThinking())
	case phaseResponse:
		b.WriteString(m.viewResponse())
		m.lastResp = ""
	case phasePalette:
		b.WriteString(m.viewPalette())
		b.WriteString(m.viewInputBox())
	case phaseInput:
		b.WriteString(m.viewInputBox())
	}

	// Footer
	b.WriteString(m.viewFooter())

	return b.String()
}

func (m model) viewHeader() string {
	provider := m.cfg.Agent.Provider
	if provider == "" {
		provider = "none"
	}
	model := m.cfg.Agent.Model
	if model == "" {
		model = "none"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styleTitle.Render("IntimClaw"))
	b.WriteString(fmt.Sprintf("  %s v%s  %s", styleDim.Render(""), VERSION, styleDim.Render("AI Agent System")))
	b.WriteString("\n")
	b.WriteString(styleGreen.Render("● Online"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Provider : %s\n", provider))
	b.WriteString(fmt.Sprintf("  Model    : %s\n", model))
	b.WriteString(fmt.Sprintf("  Session  : %s\n", m.session.Name))
	return b.String()
}

func (m model) viewInputBox() string {
	width := m.width - 4
	if width < 20 {
		width = 20
	}

	// Build the content line
	content := m.input.View()
	prompt := "> "

	// Pad content to fill width
	contentLen := lipgloss.Width(content) + lipgloss.Width(prompt)
	pad := ""
	if contentLen < width {
		pad = strings.Repeat(" ", width-contentLen)
	}

	box := fmt.Sprintf("%s%s%s%s", styleInputBoxActive.Render("│"), prompt, content, pad)
	return box + "\n"
}

func (m model) viewPalette() string {
	filter := m.input.Value()
	_ = filter // used for display

	var b strings.Builder
	b.WriteString(stylePaletteBox.Render("Commands"))
	b.WriteString("\n")

	if filter != "" {
		b.WriteString(fmt.Sprintf("  %s%s%s\n", styleDim.Render(""), filter, styleDim.Render("█")))
	}

	if len(m.paletteList) == 0 {
		b.WriteString("  (no matching commands)\n")
	} else {
		for i, name := range m.paletteList {
			desc := ""
			for _, c := range m.registry.commands {
				if c.Name == name {
					desc = c.Desc
					break
				}
			}
			line := fmt.Sprintf("  %-14s%s", name, desc)
			if i == m.paletteIdx {
				b.WriteString(stylePaletteActive.Render(line) + "\n")
			} else {
				b.WriteString(stylePaletteNormal.Render(line) + "\n")
			}
		}
	}

	return b.String()
}

func (m model) viewThinking() string {
	elapsed := time.Since(m.startAt)
	sp := m.spinner.View()
	return fmt.Sprintf("\n  %s  %s\n", sp, styleDim.Render(fmt.Sprintf("Thinking %s", formatDuration(elapsed))))
}

func (m model) viewResponse() string {
	if m.lastErr != "" {
		return fmt.Sprintf("\n%s%s%s\n\n", styleError.Render("✗ "), styleError.Render("Request failed"), styleDim.Render("\n"+m.lastErr))
	}
	if m.lastResp == "" {
		return ""
	}
	return fmt.Sprintf("\n%s\n", m.lastResp)
}

func (m model) viewFooter() string {
	parts := []string{styleFooter.Render("ctrl+p commands")}
	return "\n" + styleFooter.Render(strings.Repeat("─", m.width-2)) + "\n" +
		styleFooter.Render("  "+strings.Join(parts, "  ")) + "\n"
}

func (m model) updateWidth(w int) {
	m.width = w
}

// ─── Entry point ──────────────────────────────────────────────

func runBubbleTea(cfg *config.Config, a *agent.Agent) {
	m := initialModel(cfg, a)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%serror: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}
