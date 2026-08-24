// Package approval implements a Telegram human-in-the-loop tool approval
// gate. When a sensitive tool is about to execute, an inline keyboard with
// Approve/Reject buttons is sent to the configured Telegram chat; execution
// blocks until the operator answers or the timeout expires (default deny).
package approval

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"

	"github.com/xxayii57/bot-keuangan/pkg/logger"
)

// Options configures the Telegram approval gate.
type Options struct {
	Bot      *telego.Bot
	ChatID   int64
	Patterns []string
	Timeout  time.Duration
}

// Gate asks a human operator on Telegram before executing a tool.
type Gate struct {
	bot      *telego.Bot
	chatID   int64
	patterns []string
	timeout  time.Duration

	mu      sync.Mutex
	waiting map[string]chan bool
}

// NewGate creates a Telegram approval gate.
func NewGate(opts Options) *Gate {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Gate{
		bot:      opts.Bot,
		chatID:   opts.ChatID,
		patterns: opts.Patterns,
		timeout:  timeout,
		waiting:  make(map[string]chan bool),
	}
}

// NeedsApproval reports whether the given tool call matches any configured pattern.
func (g *Gate) NeedsApproval(tool string, args map[string]any) bool {
	if g == nil || len(g.patterns) == 0 {
		return false
	}
	summary := summarizeCall(tool, args)
	for _, p := range g.patterns {
		if p != "" && strings.Contains(summary, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// PromptAndWait sends an inline-keyboard approval prompt and blocks until
// the operator taps a button, timeout expires, or context is cancelled.
func (g *Gate) PromptAndWait(ctx context.Context, tool string, args map[string]any, chatID string) (bool, error) {
	summary := summarizeCall(tool, args)

	cbData := fmt.Sprintf("icaw:%d", time.Now().UnixNano())
	ch := make(chan bool, 1)
	g.mu.Lock()
	g.waiting[cbData] = ch
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		delete(g.waiting, cbData)
		g.mu.Unlock()
	}()

	text := fmt.Sprintf("⚠️ *Persetujuan Tool*\n\n`%s`\n\nJalankan?", truncate(summary, 900))
	keyboard := telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			telego.InlineKeyboardButton{Text: "✅ Setujui", CallbackData: cbData + ":y"},
			telego.InlineKeyboardButton{Text: "❌ Tolak", CallbackData: cbData + ":n"},
		),
	)

	target := g.chatID
	if chatID != "" {
		if cid, err := parseChatID(chatID); err == nil {
			target = cid
		}
	}

	msg := telegoutil.Message(telegoutil.ID(target), text).WithReplyMarkup(keyboard).WithParseMode("Markdown")
	sent, err := g.bot.SendMessage(ctx, msg)
	if err != nil {
		return false, fmt.Errorf("send approval prompt: %w", err)
	}

	select {
	case approved := <-ch:
		if sent != nil {
			editCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			answer := "✅ Disetujui"
			if !approved {
				answer = "❌ Ditolak"
			}
			_ = g.editResult(editCtx, target, fmt.Sprintf("%d", sent.MessageID), answer, truncate(summary, 300))
		}
		return approved, nil
	case <-time.After(g.timeout):
		logger.InfoC("approval", "approval prompt timed out; denying")
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// HandleCallback resolves a Telegram callback query to a waiting approval.
func (g *Gate) HandleCallback(data string) bool {
	if g == nil || !strings.HasPrefix(data, "icaw:") {
		return false
	}
	trimmed := strings.TrimSuffix(data, ":y")
	trimmed = strings.TrimSuffix(trimmed, ":n")
	approved := strings.HasSuffix(data, ":y")

	g.mu.Lock()
	ch, ok := g.waiting[trimmed]
	g.mu.Unlock()
	if !ok {
		return true
	}
	ch <- approved
	return true
}

func summarizeCall(tool string, args map[string]any) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(tool))
	b.WriteString(" ")
	for k, v := range args {
		fmt.Fprintf(&b, "%s=%v ", k, v)
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func parseChatID(s string) (int64, error) {
	var id int64
	if _, err := fmt.Sscanf(s, "%d", &id); err != nil {
		return 0, err
	}
	return id, nil
}

func (g *Gate) editResult(ctx context.Context, chatID int64, messageIDStr, answer, summary string) error {
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("%s\n\n`%s`", answer, summary)
	edit := telegoutil.EditMessageText(telegoutil.ID(chatID), messageID, text).WithParseMode("Markdown")
	_, err = g.bot.EditMessageText(ctx, edit)
	return err
}
