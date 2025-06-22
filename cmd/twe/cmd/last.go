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

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Print the timestamp of the end of the most recent Timewarrior interval",
	Run: func(cmd *cobra.Command, args []string) {
		cli := timew.NewCLI()
		intervals, err := cli.Export("@1")
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "could not get interval @1: %s", err)
		}
		lastInterval := intervals[0]

		// Get the "last time"
		var lastTime *timew.Datetime
		if lastInterval.End == nil {
			lastTime = &timew.Datetime{time.Now()}
		} else {
			lastTime = lastInterval.End
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", lastTime.LocalString())
	},
}

func init() {
	RootCmd.AddCommand(lastCmd)
}
