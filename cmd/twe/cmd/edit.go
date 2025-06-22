/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	edit "github.com/kgoettler/twe/internal/edit"
	timew "github.com/kgoettler/twe/pkg/timewarrior"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MaximumNArgs(1),
	Short: "Edit today's timewarrior data",
	Run: func(cmd *cobra.Command, args []string) {
		// Setup logger
		var f *os.File
		var err error
		if len(os.Getenv("DEBUG")) > 0 {
			f, err = tea.LogToFile("debug.log", "debug")
			if err != nil {
				handleError(cmd, "configuring logger: %v", err)
			}
			defer f.Close()
		}

		// Setup CLI backend
		cli := timew.NewCLI()

		// Parse date argument (if provided)
		var date time.Time
		if len(args) > 0 {
			date, err = timew.ConvertDateStringToTime(time.Now(), strings.ToLower(args[0]))
			if err != nil {
				handleError(cmd, "input date '%s' is not a valid date", args[0])
			}
		} else {
			date = time.Now()
		}

		// Setup application model
		m, err := edit.NewModel(&cli, date)
		if err != nil {
			handleError(cmd, "initializing application: %v", err)
			os.Exit(1)
		}
		if f != nil {
			m.Logfile = f
		}
		fmt.Fprintf(f, "Test\n")

		// Run application
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			handleError(cmd, "running application: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
