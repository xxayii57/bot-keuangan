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

// ─── Colors ──────────────────────────────────────────────────

const (
	colorReset = "\033[0m"
	colorCyan  = "\033[36m"
	colorDim   = "\033[2m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

// ─── Lipgloss Styles ─────────────────────────────────────────

var (
	styleLogo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("51")).
			MarginBottom(1)

	styleOnline = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	styleDimText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleInputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	stylePaletteBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	stylePaletteActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	stylePaletteNormal = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	styleFooter = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)

// ─── Phases ──────────────────────────────────────────────────

type phase int

const (
	phaseInput    phase = iota
	phasePalette
	phaseThinking
	phaseResponse
)

// ─── Model ───────────────────────────────────────────────────

type model struct {
	textinput   textinput.Model
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

func clampMin(v, min int) int {
	if v < min {
		return min
	}
	return v
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
		textinput: ti,
		cfg:       cfg,
		agent:     a,
		registry:  NewCommandRegistry(),
		session:   &chatSession{Name: "default", Started: time.Now()},
		spinner:   sp,
		phase:     phaseInput,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

// ─── Update ──────────────────────────────────────────────────

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
		if msg.err != nil {
			m.lastErr = msg.err.Error()
		} else {
			m.lastErr = ""
		}
		// KEY FIX: transition back to input phase and refocus textinput
		m.phase = phaseInput
		m.textinput.Focus()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ─── Key Handling ────────────────────────────────────────────

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.phase == phaseThinking {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.phase == phasePalette {
		return m.handlePaletteKey(msg)
	}

	return m.handleInputKey(msg)
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEnter:
		input := strings.TrimSpace(m.textinput.Value())
		if input == "" {
			return m, nil
		}
		m.textinput.Reset()

		if strings.HasPrefix(input, "/") {
			return m.executeCommand(input)
		}

		// Agent request
		m.phase = phaseThinking
		m.startAt = time.Now()
		return m, tea.Batch(m.spinner.Tick, m.sendRequest(input))

	case tea.KeyTab:
		m.showPalette("")
		return m, nil

	default:
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		val := m.textinput.Value()
		if strings.HasPrefix(val, "/") {
			m.showPalette(val)
		}
		return m, cmd
	}
}

func (m model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.phase = phaseInput
		m.textinput.Reset()
		m.textinput.Focus()
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
			m.textinput.Reset()
			m.phase = phaseInput
			return m.executeCommand(selected)
		}
		m.phase = phaseInput
		m.textinput.Reset()
		m.textinput.Focus()
		return m, nil

	case tea.KeyCtrlC:
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		val := m.textinput.Value()
		m.updatePaletteFilter(val)
		if len(m.paletteList) == 0 {
			m.phase = phaseInput
			m.textinput.Focus()
		}
		return m, cmd
	}
}

func (m *model) showPalette(filter string) {
	m.phase = phasePalette
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
	if len(parts) == 0 {
		m.phase = phaseInput
		m.textinput.Focus()
		return m, nil
	}
	cmd := strings.ToLower(parts[0])

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
		m.textinput.Focus()
		return m, nil
	}

	exit := m.registry.Execute(input, m.cfg, m.session)
	if exit {
		return m, tea.Quit
	}

	// After command execution, return to input
	m.phase = phaseInput
	m.textinput.Focus()
	return m, nil
}

func (m model) sendRequest(input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.agent.Run(input)
		return responseMsg{resp: resp, err: err}
	}
}

// ─── View ────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	// Always show logo + session info at top
	b.WriteString(m.viewLogo())
	b.WriteString("\n")
	b.WriteString(m.viewSessionInfo())
	b.WriteString("\n")

	// Content area
	switch m.phase {
	case phaseThinking:
		b.WriteString(m.viewThinking())
	case phaseResponse:
		b.WriteString(m.viewResponse())
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

func (m model) viewLogo() string {
	return styleLogo.Render("INTIMCLAW") + " " + styleDimText.Render("v"+VERSION+"  AI Agent System")
}

func (m model) viewSessionInfo() string {
	provider := m.cfg.Agent.Provider
	if provider == "" {
		provider = "none"
	}
	modelName := m.cfg.Agent.Model
	if modelName == "" {
		modelName = "none"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s● Online%s\n\n", styleOnline.Render(""), ""))
	b.WriteString(fmt.Sprintf("  Provider : %s\n", provider))
	b.WriteString(fmt.Sprintf("  Model    : %s\n", modelName))
	b.WriteString(fmt.Sprintf("  Session  : %s\n", m.session.Name))
	return b.String()
}

func (m model) viewInputBox() string {
	return styleInputBox.Width(clampMin(m.width-2, 20)).Render("> " + m.textinput.View())
}

func (m model) viewPalette() string {
	var b strings.Builder
	b.WriteString(stylePaletteBox.Render("Commands"))
	b.WriteString("\n")

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
	return fmt.Sprintf("\n  %s  %s\n", sp, styleDimText.Render(fmt.Sprintf("Thinking %s", formatDuration(elapsed))))
}

func (m model) viewResponse() string {
	if m.lastErr != "" {
		return fmt.Sprintf("\n%s%s%s\n\n", styleError.Render("✗ "), styleError.Render("Request failed"), styleDimText.Render("\n"+m.lastErr))
	}
	if m.lastResp == "" {
		return ""
	}
	return fmt.Sprintf("\n%s\n", m.lastResp)
}

func (m model) viewFooter() string {
	sepLen := clampMin(m.width-2, 0)
	sep := strings.Repeat("─", sepLen)
	return "\n" + styleFooter.Render(sep) + "\n" +
		styleFooter.Render("  ctrl+p commands") + "\n"
}

// ─── Entry Point ─────────────────────────────────────────────

func runBubbleTea(cfg *config.Config, a *agent.Agent) {
	m := initialModel(cfg, a)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%serror: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}
