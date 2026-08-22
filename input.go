package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// InputEditor provides a real text input with box UI and special key handling.
type InputEditor struct {
	reader    *bufio.Reader
	buf       []rune
	pos       int
宽度       int
	rawMode   bool
	savedTerm []byte
}

// NewInputEditor creates an editor reading from the given reader.
func NewInputEditor(r io.Reader) *InputEditor {
	return &InputEditor{
		reader: bufio.NewReader(r),
		buf:    make([]rune, 0, 256),
		pos:    0,
		宽度:    80,
	}
}

// EnableRawMode puts terminal into raw mode for character-by-character input.
func (e *InputEditor) EnableRawMode() {
	if e.rawMode {
		return
	}
	// Save current terminal settings
	if out, err := exec.Command("stty", "-g").CombinedOutput(); err == nil {
		e.savedTerm = out
	}
	// Set raw mode: no echo, no line buffering, read char by char
	exec.Command("stty", "raw", "-echo", "min", "1", "time", "0").Run()
	e.rawMode = true
}

// DisableRawMode restores terminal to original settings.
func (e *InputEditor) DisableRawMode() {
	if !e.rawMode {
		return
	}
	if len(e.savedTerm) > 0 {
		exec.Command("stty", string(e.savedTerm)).Run()
	} else {
		exec.Command("stty", "sane").Run()
	}
	e.rawMode = false
}

// ReadInput reads a complete line of input with special key handling.
// Returns the final input string and whether the user pressed Ctrl+C/Ctrl+D.
func (e *InputEditor) ReadInput() (string, bool) {
	e.buf = e.buf[:0]
	e.pos = 0
	e.renderInput()

	for {
		b, err := e.reader.ReadByte()
		if err != nil {
			return "", true // EOF
		}

		switch b {
		case 3: // Ctrl+C
			return "", true
		case 4: // Ctrl+D (EOF)
			return "", true
		case 13, 10: // Enter
			e.clearLine()
			return string(e.buf), false
		case 27: // Escape sequence
			seq, _ := e.reader.ReadByte()
			if seq == '[' {
				e.handleCSI()
			} else {
				// Alt+key or standalone escape
			}
		case 127, 8: // Backspace
			if e.pos > 0 {
				e.buf = append(e.buf[:e.pos-1], e.buf[e.pos:]...)
				e.pos--
				e.renderInput()
			}
		case 9: // Tab - ignore for now
		default:
			if b >= 32 && unicode.IsPrint(rune(b)) {
				// Insert character at cursor position
				e.buf = append(e.buf, 0)
				copy(e.buf[e.pos+1:], e.buf[e.pos:])
				e.buf[e.pos] = rune(b)
				e.pos++
				e.renderInput()
			}
		}
	}
}

func (e *InputEditor) handleCSI() {
	code, _ := e.reader.ReadByte()
	if code != 'A' && code != 'B' && code != 'C' && code != 'D' &&
		code != 'H' && code != 'F' {
		// Unknown CSI, discard rest
		for {
			b, err := e.reader.ReadByte()
			if err != nil || (b >= 0x40 && b <= 0x7E) {
				break
			}
		}
		return
	}

	switch code {
	case 'A': // Arrow Up - could be used for history
	case 'B': // Arrow Down - could be used for history
	case 'C': // Arrow Right
		if e.pos < len(e.buf) {
			e.pos++
			e.renderInput()
		}
	case 'D': // Arrow Left
		if e.pos > 0 {
			e.pos--
			e.renderInput()
		}
	case 'H': // Home
		e.pos = 0
		e.renderInput()
	case 'F': // End
		e.pos = len(e.buf)
		e.renderInput()
	}
}

func (e *InputEditor) renderInput() {
	// Clear current line and redraw
	fmt.Print("\r\033[2K")
	fmt.Printf("%s│%s > %s", colorDim, colorReset, string(e.buf))

	// Move cursor to correct position
	if e.pos < len(e.buf) {
		fmt.Printf("\033[%dD", len(e.buf)-e.pos)
	}
}

func (e *InputEditor) clearLine() {
	fmt.Print("\r\033[2K")
}

// GetWidth returns the current terminal width.
func (e *InputEditor) GetWidth() int {
	w, _, err := terminalSize()
	if err != nil {
		return 80
	}
	return w
}

func terminalSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 80, 24, err
	}
	var w, h int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d %d", &w, &h)
	return w, h, nil
}
