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

	tea "github.com/charmbracelet/bubbletea"
	edit "github.com/kgoettler/twe/internal/edit"
	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/spf13/cobra"
)

func RunCmdEdit(backend edit.TimewarriorBackend, args ...string) (*edit.Model, error) {
	// Parse date argument (if provided)
	var dateString string
	var date time.Time
	var err error
	if len(args) > 0 {
		dateString = args[0]
	} else if len(os.Getenv("TWE_EDIT_DATE")) > 0 {
		dateString = os.Getenv("TWE_EDIT_DATE")
	}
	if len(dateString) > 0 {
		date, err = timew.ConvertDateStringToTime(time.Now(), strings.ToLower(dateString))
		if err != nil {
			return nil, fmt.Errorf("input date '%s' is not a valid date", dateString)
		}
	} else {
		date = time.Now()
	}

	// Setup application model
	m, err := edit.NewModel(backend, date, nil)
	if err != nil {
		return nil, fmt.Errorf("initializing application: %v", err)
	}

	return &m, nil
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MaximumNArgs(1),
	Short: "Edit today's timewarrior data",
	Run: func(cmd *cobra.Command, args []string) {
		var f *os.File
		var err error
		if len(os.Getenv("TWE_LOGFILE")) > 0 {
			f, err = tea.LogToFile(os.Getenv("TWE_LOGFILE"), "debug")
			if err != nil {
				handleError(cmd, "configuring logger: %s", err.Error())
			}
			defer f.Close()
		}
		cli := timew.NewCLI()
		m, err := RunCmdEdit(&cli, args...)
		if err != nil {
			handleError(cmd, "initializing editor: %s", err.Error())
		}
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			handleError(cmd, "running editor: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
