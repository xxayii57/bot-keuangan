package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

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
			Foreground(lipgloss.Color("51"))

	styleOnline = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	styleDimText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleUserMsg = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1)

	styleBotMsg = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

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

	styleThinking = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
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
	viewport    viewport.Model
	cfg         *config.Config
	agent       *agent.Agent
	registry    *CommandRegistry
	session     *chatSession
	spinner     spinner.Model
	phase       phase
	width       int
	height      int
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
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	vp := viewport.New(80, 20)

	return model{
		textinput: ti,
		viewport:  vp,
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case responseMsg:
		if msg.err != nil {
			errText := styleError.Render("✗ Error: " + msg.err.Error())
			m.appendMessage(errText)
			m.lastErr = msg.err.Error()
		} else {
			m.appendMessage(styleBotMsg.Render(msg.resp))
			m.lastErr = ""
		}
		m.phase = phaseInput
		m.textinput.Focus()
		m.updateLayout()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) appendMessage(msg string) {
	// Update viewport content with word-wrapped message
	m.updateLayout()
	existing := m.viewport.View()
	if existing == "" {
		m.viewport.SetContent(msg)
	} else {
		m.viewport.SetContent(existing + "\n\n" + msg)
	}
	m.viewport.GotoBottom()
}

func (m *model) updateLayout() {
	headerHeight := 6 // logo + session info
	footerHeight := 3 // separator + footer line + input box

	if m.phase == phasePalette {
		footerHeight += 5 // palette + input
	}
	if m.phase == phaseThinking {
		footerHeight += 1 // thinking line
	}

	vpHeight := m.height - headerHeight - footerHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.Width = m.width
	m.viewport.Height = vpHeight
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

		// Wrap user message and add to viewport
		wrapped := wordwrap.String(input, clampMin(m.width-8, 20))
		userLine := styleUserMsg.Render(wrapped)
		m.appendMessage(userLine)

		m.textinput.Reset()

		if strings.HasPrefix(input, "/") {
			return m.executeCommand(input)
		}

		m.phase = phaseThinking
		m.startAt = time.Now()
		m.updateLayout()
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
		m.updateLayout()
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
			m.updateLayout()
			return m.executeCommand(selected)
		}
		m.phase = phaseInput
		m.textinput.Reset()
		m.textinput.Focus()
		m.updateLayout()
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
			m.updateLayout()
		}
		return m, cmd
	}
}

func (m *model) showPalette(filter string) {
	m.phase = phasePalette
	m.updatePaletteFilter(filter)
	m.updateLayout()
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
		m.updateLayout()
		return m, nil
	}
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/exit", "/quit":
		return m, tea.Quit
	case "/clear":
		m.viewport.SetContent("")
		return m, nil
	case "/new":
		m.session.Messages = nil
		m.session.Started = time.Now()
		m.viewport.SetContent("")
		m.phase = phaseInput
		m.textinput.Focus()
		m.updateLayout()
		return m, nil
	}

	exit := m.registry.Execute(input, m.cfg, m.session)
	if exit {
		return m, tea.Quit
	}

	m.phase = phaseInput
	m.textinput.Focus()
	m.updateLayout()
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
	var sections []string

	// Header (fixed height)
	header := m.viewLogo() + "\n" + m.viewSessionInfo()
	sections = append(sections, header)

	// Chat viewport (scrollable)
	sections = append(sections, m.viewport.View())

	// Thinking indicator (inline, only when active)
	if m.phase == phaseThinking {
		elapsed := time.Since(m.startAt)
		sp := m.spinner.View()
		sections = append(sections, styleThinking.Render(fmt.Sprintf("  %s  Thinking %s", sp, formatDuration(elapsed))))
	}

	// Palette (if visible)
	if m.phase == phasePalette {
		sections = append(sections, m.viewPalette())
	}

	// Input box (always visible, anchored near bottom)
	sections = append(sections, m.viewInputBox())

	// Footer
	sections = append(sections, m.viewFooter())

	return strings.Join(sections, "\n")
}

func (m model) viewLogo() string {
	return styleLogo.Render("INTIMCLAW") + " " + styleDimText.Render("v"+VERSION+"  AI Agent System") + "  " + styleOnline.Render("● Online")
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
	return fmt.Sprintf("  Provider : %s   Model : %s   Session : %s", provider, modelName, m.session.Name)
}

func (m model) viewInputBox() string {
	w := clampMin(m.width-2, 20)
	return styleInputBox.Width(w).Render("> " + m.textinput.View())
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

func (m model) viewFooter() string {
	sepLen := clampMin(m.width-2, 0)
	sep := strings.Repeat("─", sepLen)
	return styleFooter.Render(sep) + "\n" +
		styleFooter.Render("  ctrl+p commands")
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
