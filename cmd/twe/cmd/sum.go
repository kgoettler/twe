/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"time"

	timew "github.com/kgoettler/twe/pkg/timewarrior"

	"github.com/spf13/cobra"
)

type BackendCmdSum interface {
	Export(args ...string) ([]timew.Interval, error)
}

func RunCmdSum(backend BackendCmdSum, tags []string) (time.Duration, error) {
	intervals, err := backend.Export(tags...)
	if err != nil {
		return 0, fmt.Errorf("exporting interval for tags: %w", err)
	}
	if len(intervals) == 0 {
		return 0, fmt.Errorf("no interval found to match specified tags")
	}

	duration := time.Duration(0)
	for _, interval := range intervals {
		dur, err := interval.Duration()
		if err == nil {
			duration += dur
		}
	}
	return duration, nil
}

var sumCmd = &cobra.Command{
	Use:   "sum",
	Short: "Sum interval durations for specific tags",
	Run: func(cmd *cobra.Command, args []string) {
		cli := timew.NewCLI()
		total, err := RunCmdSum(&cli, args)
		if err != nil {
			handleError(cmd, "getting last interval: %s", err.Error())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", timew.FormatDurationTime(total))
	},
}

func init() {
	RootCmd.AddCommand(sumCmd)
}
