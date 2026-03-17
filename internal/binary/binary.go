package binary

import (
	"fmt"
	"os"
	"os/exec"
)

// MeasureSize builds the target package and returns the binary size in bytes.
func MeasureSize(dir string, target string, outputPath string, ldflags []string) (int64, error) {
	// If target is "./...", binary size measurement isn't meaningful for a single output file.
	// We skip measurement to avoid "multiple packages" errors from Go.
	if target == "./..." || target == "../..." {
		return 0, nil
	}

	args := []string{"build", "-o", outputPath}

	if len(ldflags) > 0 {
		args = append(args, "-ldflags")
		flagsStr := ""
		for i, f := range ldflags {
			if i > 0 {
				flagsStr += " "
			}
			flagsStr += f
		}
		args = append(args, flagsStr)
	}

	args = append(args, target)

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	// We don't pipe to os.Stderr here to avoid noise in the CLI
	// unless the user explicitly wants to debug build failures.

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to build %s: %w", target, err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat built binary: %w", err)
	}

	return info.Size(), nil
}
