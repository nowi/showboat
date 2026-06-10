package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/simonw/showboat/replay"
)

// ReplayOptions holds parsed CLI flags for the replay command.
type ReplayOptions struct {
	Output         string
	Format         string
	Width          int
	Height         int
	Prompt         string
	Notes          string
	Speed          float64
	TypingCPS      int
	PrintLPS       int
	PauseBetween   time.Duration
	MaxOutputLines int
	Theme          string
	Loop           bool
	Verify         bool
}

// Replay converts a showboat document into a terminal recording.
func Replay(file string, opts ReplayOptions) error {
	rOpts := replay.DefaultOptions()
	rOpts.Output = opts.Output
	if opts.Format != "" {
		rOpts.Format = opts.Format
	}
	if opts.Width > 0 {
		rOpts.Width = opts.Width
	}
	if opts.Height > 0 {
		rOpts.Height = opts.Height
	}
	if opts.Prompt != "" {
		rOpts.Prompt = opts.Prompt
	}
	if opts.Notes != "" {
		rOpts.Notes = opts.Notes
	}
	if opts.Speed > 0 {
		rOpts.Speed = opts.Speed
	}
	if opts.TypingCPS > 0 {
		rOpts.TypingCPS = opts.TypingCPS
	}
	if opts.PrintLPS > 0 {
		rOpts.PrintLPS = opts.PrintLPS
	}
	if opts.PauseBetween > 0 {
		rOpts.PauseBetween = opts.PauseBetween
	}
	if opts.MaxOutputLines > 0 {
		rOpts.MaxOutputLines = opts.MaxOutputLines
	}
	if opts.Theme != "" {
		rOpts.Theme = opts.Theme
	}
	rOpts.Loop = opts.Loop
	rOpts.Verify = opts.Verify

	return replay.Replay(file, rOpts)
}

// ParseReplayFlags parses replay-specific CLI flags from args.
func ParseReplayFlags(args []string) (file string, opts ReplayOptions, err error) {
	opts = ReplayOptions{
		Format: "cast",
		Notes:  "hide",
		Speed:  1.0,
		Theme:  "asciinema",
	}

	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-o", "--output":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("%s requires a value", arg)
			}
			opts.Output = args[i+1]
			i++
		case "--format":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--format requires a value")
			}
			opts.Format = args[i+1]
			i++
		case "--width":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--width requires a value")
			}
			opts.Width, err = strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --width: %w", err)
			}
			i++
		case "--height":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--height requires a value")
			}
			opts.Height, err = strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --height: %w", err)
			}
			i++
		case "--prompt":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--prompt requires a value")
			}
			opts.Prompt = args[i+1]
			i++
		case "--notes":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--notes requires a value")
			}
			opts.Notes = args[i+1]
			i++
		case "--speed":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--speed requires a value")
			}
			opts.Speed, err = strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return "", opts, fmt.Errorf("invalid --speed: %w", err)
			}
			i++
		case "--typing-cps":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--typing-cps requires a value")
			}
			opts.TypingCPS, err = strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --typing-cps: %w", err)
			}
			i++
		case "--print-lps":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--print-lps requires a value")
			}
			opts.PrintLPS, err = strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --print-lps: %w", err)
			}
			i++
		case "--pause-between":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--pause-between requires a value")
			}
			ms, err := strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --pause-between: %w", err)
			}
			opts.PauseBetween = time.Duration(ms) * time.Millisecond
			i++
		case "--max-output-lines":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--max-output-lines requires a value")
			}
			opts.MaxOutputLines, err = strconv.Atoi(args[i+1])
			if err != nil {
				return "", opts, fmt.Errorf("invalid --max-output-lines: %w", err)
			}
			i++
		case "--theme":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("--theme requires a value")
			}
			opts.Theme = args[i+1]
			i++
		case "--loop":
			opts.Loop = true
		case "--verify":
			opts.Verify = true
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) < 1 {
		return "", opts, fmt.Errorf("missing input file")
	}
	return positional[0], opts, nil
}
