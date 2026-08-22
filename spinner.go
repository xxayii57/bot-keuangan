package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// BrandSpinner is an animated activity indicator that highlights
// characters of the "intimclaw" brand text on a single terminal line.
// It includes a thinking timer that shows elapsed duration.
type BrandSpinner struct {
	text     string
	width    int
	interval time.Duration
	done     chan struct{}
	running  bool
	mu       sync.Mutex
	startAt  time.Time
}

// NewBrandSpinner creates a spinner with the given brand text and tick interval.
func NewBrandSpinner(text string, interval time.Duration) *BrandSpinner {
	return &BrandSpinner{
		text:     text,
		width:    len([]rune(text)),
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start begins the animation loop in a background goroutine.
func (s *BrandSpinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.startAt = time.Now()
	s.done = make(chan struct{})
	s.mu.Unlock()

	go s.animate()
}

// Stop halts the animation and clears the line.
func (s *BrandSpinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.done)

	// Clear the animation line
	fmt.Fprintf(os.Stderr, "\r\033[2K")
}

func (s *BrandSpinner) animate() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	pos := 0
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.running {
				s.mu.Unlock()
				return
			}
			elapsed := time.Since(s.startAt)
			s.renderFrame(pos, elapsed)
			s.mu.Unlock()
			pos = (pos + 1) % s.width
		}
	}
}

func (s *BrandSpinner) renderFrame(pos int, elapsed time.Duration) {
	var b strings.Builder

	b.WriteString("\r\033[2K") // clear line
	b.WriteString("  ")        // indent

	for i, ch := range s.text {
		if i == pos {
			b.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m", string(ch))) // bold cyan
		} else {
			b.WriteString(fmt.Sprintf("\033[2m%s\033[0m", string(ch))) // dim
		}
	}

	b.WriteString(fmt.Sprintf("  %s•  Thinking %s%s", colorDim, formatDuration(elapsed), colorReset))
	os.Stderr.WriteString(b.String())
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %02dm", h, m)
}
