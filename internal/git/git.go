package git

import (
	"fmt"
	"os"
	"os/exec"
)

// CreateTempWorktree creates a temporary git worktree for a given reference.
// Returns the directory of the worktree, a cleanup function, and an error.
func CreateTempWorktree(repoDir, baseRef string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "godepsguard-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create worktree
	// git worktree add --detach <path> <commit-ish>
	cmd := exec.Command("git", "-C", repoDir, "worktree", "add", "--detach", "-f", tempDir, baseRef)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("failed to create git worktree for ref %s: %w", baseRef, err)
	}

	cleanup := func() {
		// Clean up worktree from git tracking
		exec.Command("git", "-C", repoDir, "worktree", "remove", "-f", tempDir).Run()
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup, nil
}

// GetCurrentCommit returns the HEAD commit hash.
func GetCurrentCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	// Return trimmed hash
	return string(out[:len(out)-1]), nil // remove trailing newline safely
}
