package commands

import (
	"fmt"
	"strings"
	"sync"
)

// sessionsStore holds the latest session list for inline keyboard rendering.
type sessionsStore struct {
	mu       sync.RWMutex
	sessions []string
}

var globalSessionStore = &sessionsStore{}

func (s *sessionsStore) Set(sessions []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = append([]string(nil), sessions...)
}

func (s *sessionsStore) Get() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.sessions...)
}

// BuildSessionsKeyboardText renders the session list as numbered text.
func BuildSessionsKeyboardText(sessions []string, currentKey string) string {
	var b strings.Builder
	b.WriteString("📋 *Sessions*\n\n")
	for i, s := range sessions {
		marker := "  "
		if s == currentKey {
			marker = "→ "
		}
		b.WriteString(fmt.Sprintf("%s`%d)` %s\n", marker, i+1, s))
	}
	b.WriteString("\nPilih nomor untuk aksi:")
	return b.String()
}

// ParseSessionAction parses callback data like "sess:select:2" or "sess:rename:2" or "sess:delete:3".
func ParseSessionAction(data string) (action string, index int, ok bool) {
	if !strings.HasPrefix(data, "sess:") {
		return "", 0, false
	}
	parts := strings.SplitN(data, ":", 3)
	if len(parts) < 3 {
		return "", 0, false
	}
	action = parts[1]
	n := 0
	fmt.Sscanf(parts[2], "%d", &n)
	return action, n, n > 0
}
