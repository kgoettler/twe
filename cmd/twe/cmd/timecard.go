/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kgoettler/twe/internal/timecard"
	timew "github.com/kgoettler/twe/pkg/timewarrior"

	"github.com/spf13/cobra"
)

var timecardOptions timecard.TimecardOptions

var timecardCmd = &cobra.Command{
	Use:   "timecard",
	Short: "Weekly timecard report for Timewarrior",
	Long: `Prints a timecard containing the hours worked on each tag for each day of the week. 
	
	Useful for copying into a timecard software.`,
	Run: func(cmd *cobra.Command, args []string) {
		var tw *timew.Report
		var reader io.Reader
		var err error
		if timecardOptions.InputFile != "" {
			file, err := os.Open(timecardOptions.InputFile)
			if err != nil {
				handleError(cmd, "opening input file %s: %s\n", timecardOptions.InputFile, err)
				os.Exit(1)
			}
			defer file.Close()
			reader = file
		} else {
			// Get Intervals from export
			if len(args) == 0 {
				args = append(args, ":week")
			}
			cli := timew.NewCLI()
			reader, err = cli.Report(append([]string{"echo"}, args...)...)
			if err != nil {
				handleError(cmd, "running 'echo' report: %s\n", err)
				os.Exit(1)
			}
		}
		timecardOptions.OutputFormat = strings.ToLower(timecardOptions.OutputFormat)

		// Create timewarrior report object
		tw, err = timew.NewReport(reader)
		if err != nil {
			handleError(cmd, "parsing 'echo' output: %s\n", err)
			os.Exit(1)
		}

		// Run
		msg, err := timecard.Run(tw, timecardOptions)
		if err != nil {
			handleError(cmd, "%s", err)
			os.Exit(1)
		}
		fmt.Fprint(cmd.OutOrStdout(), msg)
		fmt.Fprint(cmd.OutOrStdout(), "\n")
	},
}

func init() {
	RootCmd.AddCommand(timecardCmd)
	timecardCmd.Flags().StringVar(
		&timecardOptions.OutputFormat,
		"format",
		"table",
		"Output format for report (options: table, csv)",
	)
	timecardCmd.Flags().StringVar(
		&timecardOptions.InputFile,
		"file",
		"",
		"Input file to read from. If none specified, will read from STDIN.",
	)
	timecardCmd.Flags().StringArrayVar(
		&timecardOptions.Filters,
		"filter",
		[]string{},
		"List of filters to apply to tags. Regular expressions are supported",
	)
}
