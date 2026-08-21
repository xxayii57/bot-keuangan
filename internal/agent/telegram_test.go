package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelegramGetUpdatesSuccess(t *testing.T) {
	// Mock Telegram API server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"result":[{"update_id":100,"message":{"message_id":1,"chat":{"id":12345},"text":"hello"}},{"update_id":101,"message":{"message_id":2,"chat":{"id":12345},"text":"/start"}}]}`))
	}))
	defer server.Close()

	_ = server.Client()

	// Test the response parsing logic directly.
	body := `{"ok":true,"result":[{"update_id":100,"message":{"message_id":1,"chat":{"id":12345},"text":"hello"}},{"update_id":101,"message":{"message_id":2,"chat":{"id":12345},"text":"/start"}}]}`

	var apiResp tgAPIResponse
	if err := json.Unmarshal([]byte(body), &apiResp); err != nil {
		t.Fatalf("failed to parse API response: %v", err)
	}

	if !apiResp.OK {
		t.Error("expected ok=true")
	}

	var updates []tgUpdate
	if err := json.Unmarshal(apiResp.Result, &updates); err != nil {
		t.Fatalf("failed to parse updates: %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2", len(updates))
	}
	if updates[0].UpdateID != 100 {
		t.Errorf("update[0].UpdateID = %d, want 100", updates[0].UpdateID)
	}
	if updates[0].Message == nil || updates[0].Message.Text != "hello" {
		t.Errorf("update[0].Message.Text = %v, want hello", updates[0].Message.Text)
	}
	if updates[0].Message.Chat.ID != 12345 {
		t.Errorf("update[0].Message.Chat.ID = %d, want 12345", updates[0].Message.Chat.ID)
	}
	if updates[1].Message.Text != "/start" {
		t.Errorf("update[1].Message.Text = %v, want /start", updates[1].Message.Text)
	}
}

func TestTelegramGetUpdatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":false,"error_code":401,"description":"Unauthorized"}`))
	}))
	defer server.Close()

	_ = server.Client()

	body := `{"ok":false,"error_code":401,"description":"Unauthorized"}`
	var apiResp tgAPIResponse
	if err := json.Unmarshal([]byte(body), &apiResp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if apiResp.OK {
		t.Error("expected ok=false")
	}
	if apiResp.ErrorCode != 401 {
		t.Errorf("error_code = %d, want 401", apiResp.ErrorCode)
	}
	if apiResp.Description != "Unauthorized" {
		t.Errorf("description = %v, want Unauthorized", apiResp.Description)
	}
}

func TestTelegramEmptyTokenSkip(t *testing.T) {
	bot := NewTelegramBot("", nil)
	// Should not panic, just print message.
	bot.Start()
}

func TestTelegramProcessUpdateNilMessage(t *testing.T) {
	bot := &TelegramBot{
		token: "test",
		agent: nil,
	}
	// Should not panic.
	bot.processUpdate(tgUpdate{UpdateID: 1})
	bot.processUpdate(tgUpdate{UpdateID: 2, Message: &tgMessage{Text: ""}})
}

func TestTelegramStop(t *testing.T) {
	bot := &TelegramBot{running: true}
	bot.Stop()
	if bot.running {
		t.Error("expected running=false after Stop()")
	}
}

func TestTelegramJSONEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantOK  bool
		wantErr bool
	}{
		{"empty array", `{"ok":true,"result":[]}`, true, false},
		{"null result", `{"ok":true,"result":null}`, true, false},
		{"malformed json", `not json`, false, true},
		{"missing ok field", `{"result":[]}`, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiResp tgAPIResponse
			err := json.Unmarshal([]byte(tt.body), &apiResp)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshal error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && apiResp.OK != tt.wantOK {
				t.Errorf("ok = %v, want %v", apiResp.OK, tt.wantOK)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
