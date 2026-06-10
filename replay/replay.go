package replay

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options configures replay generation.
type Options struct {
	Output         string
	Format         string // cast, gif, mp4
	Width          int
	Height         int
	Prompt         string
	Notes          string // hide, comments, narrate
	Speed          float64
	TypingCPS      int
	PrintLPS       int
	PauseBetween   time.Duration
	MaxOutputLines int
	Theme          string
	Loop           bool
	Verify         bool
}

// DefaultOptions returns sensible defaults per the spec.
func DefaultOptions() Options {
	return Options{
		Format:       "cast",
		Width:        110,
		Height:       30,
		Prompt:       "$ ",
		Notes:        "hide",
		Speed:        1.0,
		TypingCPS:    40,
		PrintLPS:     50,
		PauseBetween: 800 * time.Millisecond,
		Theme:        "asciinema",
	}
}

// Replay converts a showboat markdown file into a terminal recording.
func Replay(inputFile string, opts Options) error {
	if opts.Speed < 0.1 || opts.Speed > 10 {
		return fmt.Errorf("speed must be between 0.1 and 10")
	}
	if opts.Notes != "hide" && opts.Notes != "comments" && opts.Notes != "narrate" {
		return fmt.Errorf("notes must be hide, comments, or narrate")
	}

	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scenes, err := ParseDocument(f)
	if err != nil {
		return err
	}

	docTitle := ""
	for _, scene := range scenes {
		if t, ok := scene.(TitleScene); ok {
			docTitle = t.Text
			break
		}
	}

	outputPath := opts.Output
	if outputPath == "" {
		ext := ".cast"
		switch opts.Format {
		case "gif":
			ext = ".gif"
		case "mp4":
			ext = ".mp4"
		}
		base := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
		outputPath = base + ext
	}

	var buf bytes.Buffer
	emitter := NewEmitter(&buf, opts)
	if err := emitter.RenderScenes(scenes, docTitle); err != nil {
		return err
	}
	castData := buf.Bytes()

	if opts.Verify {
		existing, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("verify: %w", err)
		}
		if !bytes.Equal(existing, castData) {
			return fmt.Errorf("verify: output drifted from %s", outputPath)
		}
		return nil
	}

	castPath := outputPath
	if opts.Format == "gif" || opts.Format == "mp4" {
		castPath = outputPath + ".tmp.cast"
	}

	if err := os.WriteFile(castPath, castData, 0644); err != nil {
		return err
	}

	switch opts.Format {
	case "cast":
		return nil
	case "gif", "mp4":
		defer os.Remove(castPath)
		return renderWithAgg(castPath, outputPath, opts)
	default:
		return fmt.Errorf("unknown format: %s", opts.Format)
	}
}
