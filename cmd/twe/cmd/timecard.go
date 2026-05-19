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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var timecardOptions timecard.TimecardOptions

type BackendCmdTimecard interface {
	Report(args ...string) (io.Reader, error)
}

func RunCmdTimecard(backend BackendCmdTimecard, options timecard.TimecardOptions, args ...string) (string, error) {
	var tw *timew.Report
	var reader io.Reader
	var err error
	if options.InputFile != "" {
		file, err := os.Open(options.InputFile)
		if err != nil {
			return "", fmt.Errorf("opening input file %s: %w", options.InputFile, err)
		}
		defer file.Close()
		reader = file
	} else {
		// Get Intervals from export
		if len(args) == 0 {
			args = append(args, ":week")
		}
		reader, err = backend.Report(append([]string{"echo"}, args...)...)
		if err != nil {
			return "", fmt.Errorf("running 'echo' report: %w", err)
		}
	}
	options.OutputFormat = strings.ToLower(options.OutputFormat)

	// Create timewarrior report object
	tw, err = timew.NewReport(reader)
	if err != nil {
		return "", fmt.Errorf("parsing 'echo' output: %w", err)
	}

	// Run
	msg, err := timecard.Run(tw, options)
	if err != nil {
		return "", fmt.Errorf("running report: %w", err)
	}
	return msg, nil
}

var timecardCmd = &cobra.Command{
	Use:   "timecard",
	Short: "Weekly timecard report for Timewarrior",
	Long: `Prints a timecard containing the hours worked on each tag for each day of the week. 
	
	Useful for copying into a timecard software.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := timew.NewCLI()
		timecardOptions.Increment = viper.GetInt("timecard.increment")
		timecardOptions.IncludeTotalRow = viper.GetBool("timecard.total-row")
		timecardOptions.IncludeTotalCol = viper.GetBool("timecard.total-col")
		timecardOptions.Filters = viper.GetStringSlice("timecard.filter")
		timecardOptions.InputFile = viper.GetString("timecard.file")
		timecardOptions.OutputFormat = strings.ToLower(viper.GetString("timecard.format"))
		switch viper.GetString("timecard.unit") {
		case "decimal":
			timecardOptions.TimeUnit = timecard.TimeUnitDecimal
		case "time":
			timecardOptions.TimeUnit = timecard.TimeUnitHMS
		default:
			handleError(cmd, "unrecognized time unit: '%s'", viper.GetString("timecard.unit"))
		}

		msg, err := RunCmdTimecard(&cli, timecardOptions, args...)
		if err != nil {
			handleError(cmd, "running timecard: %s", err.Error())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", msg)
	},
}

func init() { //nolint: gochecknoinits // not applicable to cobra-cli files

	RootCmd.AddCommand(timecardCmd)
	timecardCmd.Flags().IntVar(
		&timecardOptions.Increment,
		"increment",
		6,
		"Increment up to which each duration will be rounded (in minutes)",
	)
	timecardCmd.Flags().BoolVar(
		&timecardOptions.IncludeTotalRow,
		"total-row",
		false,
		"Include row with daily totals",
	)
	timecardCmd.Flags().BoolVar(
		&timecardOptions.IncludeTotalCol,
		"total-col",
		false,
		"Include column with tag totals",
	)
	timecardCmd.Flags().String(
		"unit",
		"decimal",
		"Time unit for report (options: decimal, time)",
	)
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

	flags := timecardCmd.Flags()

	flags.VisitAll(func(f *pflag.Flag) {
		viperKey := fmt.Sprintf("timecard.%s", f.Name)
		ConfigKeys[viperKey] = struct{}{}
		if err := viper.BindPFlag(viperKey, f); err != nil {
			panic(fmt.Sprintf("viper.BindPFlag(%q): %v", viperKey, err))
		}
	})
}
