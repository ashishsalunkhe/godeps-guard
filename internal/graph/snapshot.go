package graph

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/internal/binary"
	"github.com/ashishsalunkhe/godeps-guard/internal/golist"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// GenerateSnapshot creates a full dependency and module snapshot for a given directory.
func GenerateSnapshot(dir string, target string, outputPath string, ldflags []string) (*types.Snapshot, error) {
	// 1. Run go list
	packages, modules, err := golist.Run(dir)
	if err != nil {
		return nil, err
	}

	// 2. Get the current commit hash if we are in a git repo
	commit := ""
	if cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD"); cmd.Err == nil {
		out, err := cmd.Output()
		if err == nil {
			commit = strings.TrimSpace(string(out))
		}
	}

	size, _ := binary.MeasureSize(dir, target, outputPath, ldflags)

	snap := &types.Snapshot{
		Modules:    modules,
		Packages:   packages,
		BinarySize: size,
		Target:     target,
		Commit:     commit,
	}

	return snap, nil
}

// WriteJSON formats the snapshot to a JSON stream.
func WriteJSON(snap *types.Snapshot, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(snap)
}
