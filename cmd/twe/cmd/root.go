/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"os"
	"slices"

	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/spf13/cobra"
)

var TIMEW_COMMANDS = []string{
	"annotate",
	"cancel",
	// "config",
	"continue",
	"day",
	"delete",
	"diagnostics",
	"export",
	"extensions",
	"gaps",
	"get",
	"help",
	"join",
	"lengthen",
	"modify",
	"month",
	"move",
	"report",
	"retag",
	"shorten",
	"show",
	"split",
	"start",
	"stop",
	"summary",
	"tag",
	"tags",
	"track",
	"undo",
	"untag",
	"week",
}

var RootCmd = &cobra.Command{
	Use:   "twe",
	Short: "Timewarrior extensions for power users",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true, // hides cmd
	},
}

func Execute() {
	if len(os.Args) > 1 {
		if slices.Contains(TIMEW_COMMANDS, os.Args[1]) {
			cli := timew.NewCLI()
			out, err := cli.Run(os.Args[1:]...)
			if err != nil {
				handleCLIError(RootCmd, err)
			}
			fmt.Fprintf(RootCmd.OutOrStdout(), "%s", out)
			os.Exit(0)
		}
	}
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func handleError(cmd *cobra.Command, msg string, args ...any) {
	fmt.Fprintf(cmd.ErrOrStderr(), "error: ")
	fmt.Fprintf(cmd.ErrOrStderr(), msg, args...)
	fmt.Fprintf(cmd.ErrOrStderr(), "\n")
	os.Exit(1)
}

func handleCLIError(cmd *cobra.Command, err error) {
	if ee, ok := err.(*timew.CLIError); ok {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s", ee.Stderr)
		os.Exit(ee.ExitCode)
	} else {
		handleError(cmd, "%s", err.Error())
	}
}

func init() {
}
