package cmd

import (
	"testing"
)

func TestParseReplayFlags(t *testing.T) {
	file, opts, err := ParseReplayFlags([]string{
		"demo.md",
		"-o", "out.cast",
		"--format", "gif",
		"--width", "80",
		"--notes", "comments",
		"--verify",
	})
	if err != nil {
		t.Fatal(err)
	}
	if file != "demo.md" {
		t.Fatalf("file = %q", file)
	}
	if opts.Output != "out.cast" || opts.Format != "gif" || opts.Width != 80 {
		t.Fatalf("unexpected opts: %#v", opts)
	}
	if opts.Notes != "comments" || !opts.Verify {
		t.Fatalf("unexpected opts: %#v", opts)
	}
}
