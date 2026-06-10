package replay

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	dimPrefix = "\x1b[2m"
	dimSuffix = "\x1b[0m"
)

// Emitter writes deterministic asciicast v2 events.
type Emitter struct {
	w             io.Writer
	t             float64
	speed         float64
	typingCps     int
	printLps      int
	pauseAfter    time.Duration
	promptStr     string
	width          int
	height         int
	maxOutputLines int
	notesMode      string
}

// NewEmitter creates an asciicast emitter with the given options.
func NewEmitter(w io.Writer, opts Options) *Emitter {
	return &Emitter{
		w:              w,
		speed:          opts.Speed,
		typingCps:      opts.TypingCPS,
		printLps:       opts.PrintLPS,
		pauseAfter:     opts.PauseBetween,
		promptStr:      opts.Prompt,
		width:          opts.Width,
		height:         opts.Height,
		maxOutputLines: opts.MaxOutputLines,
		notesMode:      opts.Notes,
	}
}

// WriteHeader writes the asciicast v2 header line.
func (e *Emitter) WriteHeader(cols, rows int, title string) error {
	header := map[string]any{
		"version":   2,
		"width":     cols,
		"height":    rows,
		"timestamp": 0,
		"env": map[string]string{
			"SHELL": "/bin/bash",
			"TERM":  "xterm-256color",
		},
		"title": title,
	}
	data, err := json.Marshal(header)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(e.w, "%s\n", data)
	return err
}

// TypeString emits per-character typing events at the given chars/sec rate.
func (e *Emitter) TypeString(s string, cps int) error {
	if cps <= 0 {
		cps = 1
	}
	delta := 1.0 / float64(cps) / e.speed

	for _, ch := range s {
		if ch == '\n' {
			if err := e.emitOutput("\r\n> "); err != nil {
				return err
			}
			continue
		}
		if err := e.emitOutput(string(ch)); err != nil {
			return err
		}
		e.t += delta
	}
	return nil
}

// PrintLine emits a single output line followed by a newline.
func (e *Emitter) PrintLine(s string) error {
	if err := e.emitOutput(s + "\r\n"); err != nil {
		return err
	}
	e.t += (1.0 / float64(e.printLps)) / e.speed
	return nil
}

// Sleep advances time without writing output.
func (e *Emitter) Sleep(d time.Duration) {
	e.t += d.Seconds() / e.speed
}

// Flush is a no-op for the streaming writer.
func (e *Emitter) Flush() error {
	return nil
}

func (e *Emitter) emitOutput(chunk string) error {
	event := []any{
		formatTime(e.t),
		"o",
		chunk,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(e.w, "%s\n", data)
	return err
}

func formatTime(t float64) float64 {
	// Round to 3 decimal places for stable serialization.
	return float64(int(t*1000+0.5)) / 1000
}

func (e *Emitter) renderTitle(title string) error {
	if title == "" {
		return nil
	}
	top, bottom := titleBanner(title, e.width)
	if err := e.PrintLine(top); err != nil {
		return err
	}
	if err := e.PrintLine(bottom); err != nil {
		return err
	}
	e.Sleep(600 * time.Millisecond)
	return nil
}

func titleBanner(title string, width int) (string, string) {
	if width < 10 {
		width = 10
	}
	prefix := "┌─ " + title + " ─"
	dashes := width - len(prefix) - 1
	if dashes < 1 {
		dashes = 1
	}
	top := prefix + strings.Repeat("─", dashes) + "┐"
	bottom := "└" + strings.Repeat("─", width-2) + "┘"
	return top, bottom
}

func (e *Emitter) renderExec(command, output string) error {
	if err := e.emitOutput(e.promptStr); err != nil {
		return err
	}
	if err := e.TypeString(command, e.typingCps); err != nil {
		return err
	}
	if err := e.emitOutput("\r\n"); err != nil {
		return err
	}
	e.Sleep(120 * time.Millisecond)

	lines := splitOutputLines(output)
	if e.maxOutputLines > 0 && len(lines) > e.maxOutputLines {
		lines = lines[:e.maxOutputLines]
		lines = append(lines, "...(truncated)")
	}
	for _, line := range lines {
		if err := e.PrintLine(line); err != nil {
			return err
		}
	}
	e.Sleep(e.pauseAfter)
	return nil
}

func (e *Emitter) renderNote(lines []string) error {
	if e.notesMode == "hide" || len(lines) == 0 {
		return nil
	}
	cps := e.typingCps * 2
	for _, line := range lines {
		text := dimPrefix + "# " + line + dimSuffix
		if err := e.TypeString(text, cps); err != nil {
			return err
		}
		if err := e.emitOutput("\r\n"); err != nil {
			return err
		}
	}
	if e.notesMode == "narrate" {
		e.Sleep(400 * time.Millisecond)
	}
	return nil
}

func (e *Emitter) renderImage(alt, path string) error {
	label := alt
	if label == "" {
		label = path
	}
	if err := e.PrintLine("[image: " + label + "]"); err != nil {
		return err
	}
	e.Sleep(400 * time.Millisecond)
	return nil
}

func splitOutputLines(output string) []string {
	if output == "" {
		return nil
	}
	// Normalize to \n for splitting; preserve content without trailing empty line.
	text := strings.ReplaceAll(output, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// RenderScenes writes all scenes to the emitter.
func (e *Emitter) RenderScenes(scenes []Scene, docTitle string) error {
	if err := e.WriteHeader(e.width, e.height, docTitle); err != nil {
		return err
	}

	for _, scene := range scenes {
		switch s := scene.(type) {
		case TitleScene:
			if err := e.renderTitle(s.Text); err != nil {
				return err
			}
		case NoteScene:
			if err := e.renderNote(s.Lines); err != nil {
				return err
			}
		case ExecScene:
			if err := e.renderExec(s.Command, s.Output); err != nil {
				return err
			}
		case ImageScene:
			if err := e.renderImage(s.Alt, s.Path); err != nil {
				return err
			}
		}
	}
	return nil
}
