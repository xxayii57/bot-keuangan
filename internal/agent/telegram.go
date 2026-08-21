package agent

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TelegramBot handles the Telegram long-polling gateway.
type TelegramBot struct {
	token     string
	agent     *Agent
	offset    int
	running   bool
}

func NewTelegramBot(token string, agent *Agent) *TelegramBot {
	return &TelegramBot{
		token: token,
		agent: agent,
	}
}

func (b *TelegramBot) Start() {
	if b.token == "" {
		fmt.Println("[telegram] no bot token configured, skipping.")
		return
	}
	b.running = true
	fmt.Println("[telegram] bot started, polling...")

	for b.running {
		updates, err := b.getUpdates()
		if err != nil {
			fmt.Printf("[telegram] poll error: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, update := range updates {
			b.processUpdate(update)
		}
	}
}

func (b *TelegramBot) Stop() {
	b.running = false
}

type tgUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *tgMessage `json:"message"`
}

type tgMessage struct {
	MessageID int `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

func (b *TelegramBot) getUpdates() ([]tgUpdate, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.token, b.offset)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Minimal JSON parse for update array
	// In production, use encoding/json with proper structs
	var buf = make([]byte, 65536)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, `"ok":true`) || !strings.Contains(body, `"result"`) {
		return nil, fmt.Errorf("unexpected response: %s", body[:min(len(body), 200)])
	}

	// Extract update_id of last update to advance offset
	updates := parseUpdatesJSON(body)
	if len(updates) > 0 {
		b.offset = updates[len(updates)-1].UpdateID + 1
	}
	return updates, nil
}

func (b *TelegramBot) processUpdate(update tgUpdate) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Handle /start and /help
	if text == "/start" || text == "/help" {
		b.sendMessage(chatID, "IntimClaw Agent — Siap mengeksekusi perintah.")
		return
	}

	// Run agent
	response, err := b.agent.Run(text)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Error: %v", err))
		return
	}
	b.sendMessage(chatID, response)
}

func (b *TelegramBot) sendMessage(chatID int64, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("text", text)
	data.Set("parse_mode", "Markdown")

	http.PostForm(apiURL, data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseUpdatesJSON is a minimal JSON parser for Telegram updates.
// Extracts update_id values without importing encoding/json for simplicity.
func parseUpdatesJSON(body string) []tgUpdate {
	var updates []tgUpdate
	// Find all "update_id": N patterns
	for {
		idx := strings.Index(body, `"update_id":`)
		if idx == -1 {
			break
		}
		body = body[idx+12:]
		end := strings.IndexAny(body, ",}")
		if end == -1 {
			break
		}
		var uid int
		fmt.Sscanf(strings.TrimSpace(body[:end]), "%d", &uid)
		updates = append(updates, tgUpdate{UpdateID: uid})
	}
	return updates
}
