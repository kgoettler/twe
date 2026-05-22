/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"

	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/spf13/cobra"
)

type ImportOptions struct {
	InputFile string
}

var importOptions ImportOptions

type BackendCmdImport interface {
	Track(interval timew.Interval) error
}

func RunCmdImport(backend BackendCmdImport, reader io.Reader) error {
	// Get io.Reader for input
	// Parse input
	var input []timew.Interval
	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		return fmt.Errorf("decoding intervals: %w", err)
	}

	// Sort intervals by start time
	slices.SortFunc(input, func(a, b timew.Interval) int {
		if a.Start.Time.Before(b.Start.Time) {
			return -1
		} else if a.Start.Time.After(b.Start.Time) {
			return 1
		}
		return 0
	})

	// Import intervals one-by-one
	for _, interval := range input {
		err := backend.Track(interval)
		if err != nil {
			return fmt.Errorf("unable to import interval %v: %v", err, interval)
		}
	}

	return nil
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import intervals into Timewarrior",
	Run: func(cmd *cobra.Command, args []string) {
		// Get io.Reader for input
		var reader io.Reader
		if importOptions.InputFile != "" {
			file, err := os.Open(importOptions.InputFile)
			if err != nil {
				fmt.Printf("Error: Unable to open file %s: %v\n", importOptions.InputFile, err)
				return
			}
			defer file.Close()
			reader = file
		} else {
			reader = cmd.InOrStdin()
		}

		cli := timew.NewCLI()
		err := RunCmdImport(&cli, reader)
		if err != nil {
			handleError(cmd, "importing intervals: %s", err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(
		&importOptions.InputFile,
		"file",
		"f",
		"",
		"Input file to read from. If none specified, will read from STDIN.",
	)
}
