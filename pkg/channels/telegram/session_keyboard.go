package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

type SessionKeyboard struct {
	bot             *telego.Bot
	mu              sync.Mutex
	sessions        []string
	current         string
	pendingAction   *pendingAction
	RenameSession   func(oldKey, newKey string) error
	DeleteSession   func(key string) error
	RefreshSessions func() []string
}

type pendingAction struct {
	action string
	index  int
}

func NewSessionKeyboard(bot *telego.Bot) *SessionKeyboard {
	return &SessionKeyboard{bot: bot}
}

func (sk *SessionKeyboard) UpdateSessions(sessions []string, current string) {
	sk.mu.Lock()
	defer sk.mu.Unlock()
	sk.sessions = append([]string(nil), sessions...)
	sk.current = current
}

func (sk *SessionKeyboard) HandleCommand(ctx context.Context, chatID int64, text string) bool {
	if !strings.HasPrefix(text, "/sessions") {
		return false
	}
	sk.refreshAndSend(ctx, chatID)
	return true
}

func (sk *SessionKeyboard) refreshAndSend(ctx context.Context, chatID int64) {
	if sk.RefreshSessions != nil {
		sk.UpdateSessions(sk.RefreshSessions(), "")
	}
	sk.SendSessionList(ctx, chatID)
}

func (sk *SessionKeyboard) SendSessionList(ctx context.Context, chatID int64) error {
	sk.mu.Lock()
	sessions := append([]string(nil), sk.sessions...)
	current := sk.current
	sk.mu.Unlock()

	if len(sessions) == 0 {
		_, err := sk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID), "Tidak ada session tersimpan."))
		return err
	}

	text := fmt.Sprintf("*Sessions*\n\nCurrent: `%s`\nPilih nomor untuk aksi:", current)

	var rows [][]telego.InlineKeyboardButton
	for i, s := range sessions {
		name := truncateSessionName(s, 20)
		if s == current {
			name = name + " (aktif)"
		}
		rows = append(rows, []telego.InlineKeyboardButton{
			{Text: fmt.Sprintf("%d) %s", i+1, name), CallbackData: fmt.Sprintf("sess:sel:%d", i+1)},
			{Text: "\u270f\ufe0f", CallbackData: fmt.Sprintf("sess:ren:%d", i+1)},
			{Text: "\U0001f5d1", CallbackData: fmt.Sprintf("sess:del:%d", i+1)},
		})
	}
	rows = append(rows, []telego.InlineKeyboardButton{
		{Text: "Refresh", CallbackData: "sess:refresh"},
	})

	msg := telegoutil.Message(telegoutil.ID(chatID), text).
		WithParseMode("Markdown").
		WithReplyMarkup(telegoutil.InlineKeyboard(rows...))

	_, err := sk.bot.SendMessage(ctx, msg)
	return err
}

func (sk *SessionKeyboard) HandleCallback(ctx context.Context, query telego.CallbackQuery) bool {
	data := query.Data
	if !strings.HasPrefix(data, "sess:") {
		return false
	}

	chatID := query.Message.GetChat().ID
	parts := strings.SplitN(data, ":", 3)
	if len(parts) < 3 {
		return true
	}
	action := parts[1]
	idx, _ := strconv.Atoi(parts[2])
	answer := func(text string) {
		sk.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID, Text: text,
		})
	}
	send := func(text string) {
		sk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID), text))
	}
	sendKb := func(text string, kb *telego.InlineKeyboardMarkup) {
		sk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID), text).WithReplyMarkup(kb))
	}

	switch action {
	case "sel":
		answer(fmt.Sprintf("Session %d dipilih.", idx))

	case "ren":
		sk.mu.Lock()
		sk.pendingAction = &pendingAction{action: "rename", index: idx}
		sk.mu.Unlock()
		send(fmt.Sprintf("Kirim nama baru untuk session %d:", idx))
		answer("")

	case "del":
		sk.mu.Lock()
		sk.pendingAction = &pendingAction{action: "delete", index: idx}
		sk.mu.Unlock()
		kb := telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(
			telego.InlineKeyboardButton{Text: "Ya, hapus", CallbackData: fmt.Sprintf("sess:cdel:%d", idx)},
			telego.InlineKeyboardButton{Text: "Batal", CallbackData: "sess:cancel"},
		))
		sendKb(fmt.Sprintf("Hapus session %d?", idx), kb)
		answer("")

	case "cdel":
		// Confirmed delete — actually delete
		sk.mu.Lock()
		sessions := append([]string(nil), sk.sessions...)
		sk.mu.Unlock()
		if idx > 0 && idx <= len(sessions) {
			key := sessions[idx-1]
			if sk.DeleteSession != nil {
				if err := sk.DeleteSession(key); err != nil {
					answer(fmt.Sprintf("Error: %v", err))
					return true
				}
			}
			answer(fmt.Sprintf("Session %d dihapus.", idx))
			sk.refreshAndSend(ctx, chatID)
			return true
		}
		answer("Session tidak ditemukan")

	case "cancel":
		sk.mu.Lock()
		sk.pendingAction = nil
		sk.mu.Unlock()
		answer("Dibatalkan")

	case "refresh":
		sk.refreshAndSend(ctx, chatID)
		answer("")
	}

	return true
}

// ProcessPendingAction processes a pending rename/delete when user sends text.
// Returns true if handled.
func (sk *SessionKeyboard) ProcessPendingAction(ctx context.Context, chatID int64, text string) bool {
	pa := sk.GetPendingAction()
	if pa == nil {
		return false
	}

	newName := strings.TrimSpace(text)
	if newName == "" {
		return false
	}

	sk.mu.Lock()
	sessions := append([]string(nil), sk.sessions...)
	sk.mu.Unlock()

	if pa.index < 1 || pa.index > len(sessions) {
		return false
	}
	oldKey := sessions[pa.index-1]

	switch pa.action {
	case "rename":
		if sk.RenameSession != nil {
			if err := sk.RenameSession(oldKey, newName); err != nil {
				sk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID),
					fmt.Sprintf("Error rename: %v", err)))
				return true
			}
		}
	case "delete":
		if sk.DeleteSession != nil {
			if err := sk.DeleteSession(oldKey); err != nil {
				sk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID),
					fmt.Sprintf("Error hapus: %v", err)))
				return true
			}
		}
	}

	sk.refreshAndSend(ctx, chatID)
	return true
}

func (sk *SessionKeyboard) GetPendingAction() *pendingAction {
	sk.mu.Lock()
	defer sk.mu.Unlock()
	pa := sk.pendingAction
	sk.pendingAction = nil
	return pa
}

func truncateSessionName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "..."
}
