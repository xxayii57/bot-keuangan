package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// ModelKeyboard manages inline keyboards for model selection in Telegram.
type ModelKeyboard struct {
	bot *telego.Bot
	mu  sync.Mutex
}

// NewModelKeyboard creates a model keyboard manager.
func NewModelKeyboard(bot *telego.Bot) *ModelKeyboard {
	return &ModelKeyboard{bot: bot}
}

// ModelInfo represents a model from the provider API.
type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// fetchModels fetches models from an OpenAI-compatible endpoint.
func fetchModels(apiBase, apiKey string, timeout time.Duration) ([]ModelInfo, error) {
	if apiBase == "" || apiKey == "" {
		return nil, fmt.Errorf("API base and key are required")
	}
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", strings.TrimRight(apiBase, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateStr(string(body), 200))
	}
	var data struct {
		Data []ModelInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	sort.Slice(data.Data, func(i, j int) bool { return data.Data[i].ID < data.Data[j].ID })
	return data.Data, nil
}

// SendModelList sends the model list as inline keyboard from live API data.
func (mk *ModelKeyboard) SendModelList(ctx context.Context, chatID int64, apiBase, apiKey, currentModel string) error {
	models, err := fetchModels(apiBase, apiKey, 15*time.Second)
	if err != nil {
		_, sendErr := mk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID),
			fmt.Sprintf("Gagal ambil model dari API: %v", err)))
		return sendErr
	}

	if len(models) == 0 {
		_, err := mk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID), "Tidak ada model tersedia di endpoint ini."))
		return err
	}

	text := fmt.Sprintf("*Models* (`%d` tersedia)\nCurrent: `%s`\nPilih model:", len(models), currentModel)

	// Build rows: max 3 buttons per row
	var rows [][]telego.InlineKeyboardButton
	var row []telego.InlineKeyboardButton
	for _, m := range models {
		prefix := ""
		if m.ID == currentModel {
			prefix = "✓ "
		}
		btn := telego.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s%s", prefix, truncateStr(m.ID, 25)),
			CallbackData: fmt.Sprintf("model:sel:%s", m.ID),
		}
		row = append(row, btn)
		if len(row) == 3 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	keyboard := telegoutil.InlineKeyboard(rows...)
	msg := telegoutil.Message(telegoutil.ID(chatID), text).
		WithParseMode("Markdown").
		WithReplyMarkup(keyboard)
	_, err = mk.bot.SendMessage(ctx, msg)
	return err
}

// HandleCallback processes model selection callbacks. Returns true if handled.
func (mk *ModelKeyboard) HandleCallback(ctx context.Context, query telego.CallbackQuery, switchModel func(name string) error) bool {
	data := query.Data
	if !strings.HasPrefix(data, "model:") {
		return false
	}

	parts := strings.SplitN(data, ":", 3)
	if len(parts) < 3 {
		return true
	}
	action := parts[1]
	modelName := parts[2]
	chatID := query.Message.GetChat().ID

	switch action {
	case "sel":
		answer := fmt.Sprintf("Model %s dipilih.", modelName)
		mk.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID, Text: answer,
		})
		if switchModel != nil {
			if err := switchModel(modelName); err != nil {
				mk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID),
					fmt.Sprintf("Error: %v", err)))
				return true
			}
		}
		// Confirm with a simple message
		mk.bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(chatID),
			fmt.Sprintf("✅ Model changed to `%s`", modelName)).WithParseMode("Markdown"))
		return true
	}
	return false
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
