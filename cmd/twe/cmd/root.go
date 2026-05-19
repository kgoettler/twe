/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"path/filepath"

	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
var cfgFile string

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
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: $HOME/.config/twe/config.yaml)")
}

func initConfig() {
	viper.SetConfigFile(resolveConfigFilePath())
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "error reading config file:", err)
		os.Exit(1)
	}
}

// resolveConfigFilePath returns the config file path using the precedence:
// --config flag > TWE_CONFIG env var > default ($HOME/.config/twe/config.yaml).
func resolveConfigFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}
	if p := os.Getenv("TWE_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "twe", "config.yaml")
}
