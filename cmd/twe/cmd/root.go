/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "twe",
	Short: "Timewarrior extensions for power users",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true, // hides cmd
	},
}

func Execute() {
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

func init() {
}
