package main

import (
	"testing"
	"time"

	"github.com/xxayii/intimclaw/internal/agent"
	"github.com/xxayii/intimclaw/internal/config"
)

func TestBrandSpinnerStartStop(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()
}

func TestBrandSpinnerDoubleStart(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Start() // no-op
	time.Sleep(100 * time.Millisecond)
	s.Stop()
}

func TestBrandSpinnerDoubleStop(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	s.Stop() // no-op
}

func TestBrandSpinnerStopBeforeStart(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Stop() // no-op
}

func TestBrandSpinnerTimerStops(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(350 * time.Millisecond)
	s.Stop()
	time.Sleep(100 * time.Millisecond)
}

func TestBrandSpinnerRequestError(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()
}

func TestBrandSpinnerCancellation(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(150 * time.Millisecond)
	s.Stop()
}

func TestBrandSpinnerTextWidth(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 100*time.Millisecond)
	if s.width != 9 {
		t.Errorf("expected width 9, got %d", s.width)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{500 * time.Millisecond, "0.5s"},
		{1500 * time.Millisecond, "1.5s"},
		{65 * time.Second, "1m 05s"},
		{3661 * time.Second, "1h 01m"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestChatSessionReset(t *testing.T) {
	sess := &chatSession{Name: "default", Started: time.Now()}
	sess.Messages = append(sess.Messages, agent.Message{Role: "user", Content: "hello"})
	if len(sess.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(sess.Messages))
	}
	sess.Messages = nil
	sess.Started = time.Now()
	if len(sess.Messages) != 0 {
		t.Errorf("expected 0 messages after reset, got %d", len(sess.Messages))
	}
}

func TestChatSessionContextEstimate(t *testing.T) {
	sess := &chatSession{Name: "test"}
	total := 0
	for _, m := range sess.Messages {
		total += len(m.Content)
	}
	if total != 0 {
		t.Errorf("expected 0 chars, got %d", total)
	}
	sess.Messages = append(sess.Messages, agent.Message{Role: "user", Content: "hello world"})
	total = 0
	for _, m := range sess.Messages {
		total += len(m.Content)
	}
	if total != 11 {
		t.Errorf("expected 11 chars, got %d", total)
	}
}

func TestFormatDurationEdgeCases(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "0.0s"},
		{100 * time.Millisecond, "0.1s"},
		{999 * time.Millisecond, "1.0s"},
		{59 * time.Second, "59.0s"},
		{60 * time.Second, "1m 00s"},
		{90 * time.Second, "1m 30s"},
		{3599 * time.Second, "59m 59s"},
		{3600 * time.Second, "1h 00m"},
		{5400 * time.Second, "1h 30m"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSlashHelpOutput(t *testing.T) {
	printSlashHelp()
}

func TestSlashModelOutput(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4o"
	printModelInfo(cfg)
}

func TestSlashModelsOutput(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4o"
	cfg.Providers = []config.ProviderConfig{
		{Name: "openai", Models: []string{"gpt-4o", "gpt-4o-mini"}},
		{Name: "anthropic", Models: []string{"claude-3-5-sonnet-latest"}},
	}
	printAvailableModels(cfg)
}

func TestSlashStatusOutput(t *testing.T) {
	printStatus()
}

func TestSlashContextOutput(t *testing.T) {
	sess := &chatSession{Name: "default"}
	sess.Messages = append(sess.Messages, agent.Message{Role: "user", Content: "hello"})
	sess.Messages = append(sess.Messages, agent.Message{Role: "assistant", Content: "hi there"})
	printContext(sess)
}

func TestSlashSessionOutput(t *testing.T) {
	sess := &chatSession{Name: "test", Started: time.Now()}
	sess.Messages = append(sess.Messages, agent.Message{Role: "user", Content: "hello"})
	printSessionInfo(sess)
}

func TestFormatDurationNegative(t *testing.T) {
	got := formatDuration(-1 * time.Second)
	if got != "-1.0s" {
		t.Logf("formatDuration(-1s) = %q", got)
	}
}
