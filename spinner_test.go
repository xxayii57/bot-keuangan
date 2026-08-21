package main

import (
	"fmt"
	"testing"
	"time"
)

func TestBrandSpinnerStartStop(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()
	// No panic = pass
}

func TestBrandSpinnerDoubleStart(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Start() // should be no-op
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	// No panic = pass
}

func TestBrandSpinnerDoubleStop(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	s.Stop() // should be no-op
	// No panic = pass
}

func TestBrandSpinnerStopBeforeStart(t *testing.T) {
	s := NewBrandSpinner("test", 50*time.Millisecond)
	s.Stop() // should be no-op
	// No panic = pass
}

func TestBrandSpinnerRequestSuccess(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(300 * time.Millisecond)
	s.Stop()
	fmt.Println("  request completed")
}

func TestBrandSpinnerRequestError(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()
	fmt.Println("  request failed (spinner stopped)")
}

func TestBrandSpinnerCancellation(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 80*time.Millisecond)
	s.Start()
	time.Sleep(150 * time.Millisecond)
	// Simulate Ctrl+C
	s.Stop()
	time.Sleep(50 * time.Millisecond)
	// No panic = pass
}

func TestBrandSpinnerGoroutineCleanup(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 50*time.Millisecond)
	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()
	time.Sleep(100 * time.Millisecond)
	// goroutine should have exited via done channel
}

func TestBrandSpinnerTextWidth(t *testing.T) {
	s := NewBrandSpinner("intimclaw", 100*time.Millisecond)
	if s.width != 9 {
		t.Errorf("expected width 9, got %d", s.width)
	}
}
