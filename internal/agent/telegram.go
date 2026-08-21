package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// TelegramBot handles the Telegram long-polling gateway.
type TelegramBot struct {
	token   string
	agent   *Agent
	offset  int
	running bool
	client  *http.Client
}

func NewTelegramBot(token string, agent *Agent) *TelegramBot {
	return &TelegramBot{
		token:  token,
		agent:  agent,
		client: &http.Client{Timeout: 60 * time.Second},
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

// Telegram API response and update types.

type tgAPIResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Description string     `json:"description,omitempty"`
	ErrorCode   int        `json:"error_code,omitempty"`
}

type tgUpdate struct {
	UpdateID int        `json:"update_id"`
	Message  *tgMessage `json:"message,omitempty"`
}

type tgMessage struct {
	MessageID int    `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text,omitempty"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

func (b *TelegramBot) getUpdates() ([]tgUpdate, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.token, b.offset)

	resp, err := b.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp tgAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w (body: %s)", err, truncate(string(body), 200))
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("API error %d: %s", apiResp.ErrorCode, apiResp.Description)
	}

	var updates []tgUpdate
	if len(apiResp.Result) > 0 {
		if err := json.Unmarshal(apiResp.Result, &updates); err != nil {
			return nil, fmt.Errorf("failed to parse updates array: %w", err)
		}
	}

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

	resp, err := b.client.PostForm(apiURL, data)
	if err != nil {
		fmt.Printf("[telegram] failed to send message to %d: %v\n", chatID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[telegram] sendMessage returned %d: %s\n", resp.StatusCode, truncate(string(body), 200))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
