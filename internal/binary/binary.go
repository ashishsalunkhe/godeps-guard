package binary

import (
	"fmt"
	"os"
	"os/exec"
)

// MeasureSize builds the target package and returns the binary size in bytes.
func MeasureSize(dir string, target string, outputPath string, ldflags []string) (int64, error) {
	// e.g. go build -o /tmp/app -ldflags="-s -w" ./cmd/api
	args := []string{"build", "-o", outputPath}

	if len(ldflags) > 0 {
		// we combine flags if requested, or just pass them
		// simple implementation: pass them as standard arguments
		// usually they come in like -ldflags "-s -w", so we just append.
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
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to build %s: %w", target, err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat built binary: %w", err)
	}

	return info.Size(), nil
}
