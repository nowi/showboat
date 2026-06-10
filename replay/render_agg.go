package replay

import (
	"fmt"
	"os/exec"
)

func renderWithAgg(castPath, outputPath string, opts Options) error {
	args := []string{castPath, outputPath, "--theme", opts.Theme}
	if opts.Loop {
		args = append(args, "--loop")
	}
	cmd := exec.Command("agg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("agg: %w\n%s", err, out)
	}
	return nil
}
