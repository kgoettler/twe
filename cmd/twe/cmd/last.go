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

type BackendCmdLast interface {
	Export(args ...string) ([]timew.Interval, error)
}

func RunCmdLast(backend BackendCmdLast) (*timew.Datetime, error) {
	intervals, err := backend.Export("@1")
	if err != nil {
		return nil, fmt.Errorf("exporting interval @1: %w", err)
	}
	if len(intervals) == 0 {
		return nil, fmt.Errorf("interval @1 not found (are there any intervals in the db?)")
	}
	lastInterval := intervals[0]

	// Get the "last time"
	var lastTime *timew.Datetime
	if lastInterval.End == nil {
		lastTime = &timew.Datetime{time.Now()}
	} else {
		lastTime = lastInterval.End
	}
	return lastTime, nil
}

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Print the timestamp of the end of the most recent Timewarrior interval",
	Run: func(cmd *cobra.Command, args []string) {
		cli := timew.NewCLI()
		lastTime, err := RunCmdLast(&cli)
		if err != nil {
			handleError(cmd, "getting last interval: %s", err.Error())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", lastTime.LocalString())
	},
}

func init() {
	RootCmd.AddCommand(lastCmd)
}
