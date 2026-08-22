package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"unicode"
)

// InputEditor provides character-by-character input with cursor control.
// Uses stty raw mode on Unix for escape sequence detection.
// Falls back to bufio on platforms where stty is unavailable.
type InputEditor struct {
	reader    *bufio.Reader
	buf       []rune
	pos       int
	rawMode   bool
	savedTerm string
	fallback  bool // true if raw mode unavailable
}

func NewInputEditor(r io.Reader) *InputEditor {
	return &InputEditor{
		reader: bufio.NewReader(r),
		buf:    make([]rune, 0, 256),
	}
}

func (e *InputEditor) EnableRawMode() {
	if e.rawMode {
		return
	}
	// Try to save terminal settings
	out, err := exec.Command("stty", "-g").CombinedOutput()
	if err != nil {
		e.fallback = true
		return
	}
	e.savedTerm = string(out)
	if err := exec.Command("stty", "raw", "-echo", "min", "1", "time", "0").Run(); err != nil {
		e.fallback = true
		return
	}
	e.rawMode = true
}

func (e *InputEditor) DisableRawMode() {
	if !e.rawMode {
		return
	}
	if e.savedTerm != "" {
		exec.Command("stty", e.savedTerm).Run()
	} else {
		exec.Command("stty", "sane").Run()
	}
	e.rawMode = false
}

// ReadInput reads one line of input with special key handling.
// Returns (input, exited). exited=true means Ctrl+C/Ctrl+D/EOF.
func (e *InputEditor) ReadInput() (string, bool) {
	e.buf = e.buf[:0]
	e.pos = 0
	e.renderInput()

	for {
		b, err := e.reader.ReadByte()
		if err != nil {
			return "", true
		}

		switch b {
		case 3: // Ctrl+C
			return "", true
		case 4: // Ctrl+D
			return "", true
		case 13, 10: // Enter
			e.clearLine()
			return string(e.buf), false
		case 27: // Escape
			if !e.fallback {
				e.handleEscape()
			}
		case 127, 8: // Backspace
			if e.pos > 0 {
				e.buf = append(e.buf[:e.pos-1], e.buf[e.pos:]...)
				e.pos--
				e.renderInput()
			}
		case 9: // Tab - ignore
		default:
			if b >= 32 && unicode.IsPrint(rune(b)) {
				e.buf = append(e.buf, 0)
				copy(e.buf[e.pos+1:], e.buf[e.pos:])
				e.buf[e.pos] = rune(b)
				e.pos++
				e.renderInput()
			}
		}
	}
}

func (e *InputEditor) handleEscape() {
	b, err := e.reader.ReadByte()
	if err != nil {
		return
	}
	if b != '[' {
		return
	}
	// Read CSI parameter bytes until final byte
	var params []byte
	for {
		b, err = e.reader.ReadByte()
		if err != nil {
			return
		}
		if b >= 0x40 && b <= 0x7E {
			// This is the final byte
			switch b {
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
			case '~': // Handle sequences like ESC[3~ (Delete)
				if len(params) > 0 && params[0] == '3' {
					if e.pos < len(e.buf) {
						e.buf = append(e.buf[:e.pos], e.buf[e.pos+1:]...)
						e.renderInput()
					}
				}
			}
			return
		}
		params = append(params, b)
	}
}

func (e *InputEditor) renderInput() {
	fmt.Print("\r\033[2K")
	fmt.Printf("%s│%s > %s", colorDim, colorReset, string(e.buf))
	// Move cursor back to correct position
	if e.pos < len(e.buf) {
		fmt.Printf("\033[%dD", len(e.buf)-e.pos)
	}
}

func (e *InputEditor) clearLine() {
	fmt.Print("\r\033[2K")
}

func (e *InputEditor) GetWidth() int {
	out, err := exec.Command("stty", "size").CombinedOutput()
	if err != nil {
		return 80
	}
	var w int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &w)
	if w == 0 {
		return 80
	}
	return w
}
