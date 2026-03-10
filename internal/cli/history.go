package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/git"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Track dependency growth over time",
}

var historyRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a new snapshot in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(".godepsguard.yaml")
		dir := "."

		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return err
		}

		commit, _ := git.GetCurrentCommit(dir)

		record := types.HistoryRecord{
			Date:          time.Now(),
			Commit:        commit,
			TotalPackages: len(snap.Packages),
			BinarySize:    snap.BinarySize,
		}

		for _, m := range snap.Modules {
			if !m.Indirect {
				record.DirectDeps++
			}
		}

		histDir := ".godepsguard"
		os.MkdirAll(histDir, 0755)

		histFile := filepath.Join(histDir, "history.json")
		hist := loadHistory(histFile, cfg.Build.Target)

		hist.Records = append(hist.Records, record)

		data, _ := json.MarshalIndent(hist, "", "  ")
		return os.WriteFile(histFile, data, 0644)
	},
}

var historyReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Print the historical growth report",
	RunE: func(cmd *cobra.Command, args []string) error {
		histDir := ".godepsguard"
		histFile := filepath.Join(histDir, "history.json")
		hist := loadHistory(histFile, "")

		if len(hist.Records) == 0 {
			fmt.Println("No history recorded yet.")
			return nil
		}

		fmt.Println("Dependency growth over time")
		fmt.Println("---------------------------------------------------------")
		fmt.Printf("%-12s %-10s %-12s %-12s\\n", "Date", "Deps", "Packages", "Binary")
		fmt.Println("---------------------------------------------------------")

		for _, r := range hist.Records {
			fmt.Printf("%-12s %-10d %-12d %-12s\\n",
				r.Date.Format("2006-01-02"), r.DirectDeps, r.TotalPackages, formatBytes(r.BinarySize))
		}

		if len(hist.Records) > 1 {
			first := hist.Records[0]
			last := hist.Records[len(hist.Records)-1]
			fmt.Println("---------------------------------------------------------")
			fmt.Printf("Total Growth:\\n")
			fmt.Printf("Deps:     %+d\\n", last.DirectDeps-first.DirectDeps)
			fmt.Printf("Packages: %+d\\n", last.TotalPackages-first.TotalPackages)
			fmt.Printf("Binary:   %s\\n", formatBytesDelta(last.BinarySize-first.BinarySize))
		}

		return nil
	},
}

func loadHistory(path, target string) types.History {
	data, err := os.ReadFile(path)
	var hist types.History
	if err == nil {
		json.Unmarshal(data, &hist)
	}
	if hist.Target == "" {
		hist.Target = target
	}
	return hist
}

func formatBytes(b int64) string {
	mb := float64(b) / 1024 / 1024
	return fmt.Sprintf("%.2fMB", mb)
}

func formatBytesDelta(b int64) string {
	mb := float64(b) / 1024 / 1024
	if mb > 0 {
		return fmt.Sprintf("+%.2fMB", mb)
	}
	return fmt.Sprintf("%.2fMB", mb)
}

func init() {
	historyCmd.AddCommand(historyRecordCmd)
	historyCmd.AddCommand(historyReportCmd)
	rootCmd.AddCommand(historyCmd)
}
