package timecard 

import (

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for table formatting
	HeaderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Align(lipgloss.Center)
	EvenRowStyle  = lipgloss.NewStyle().Padding(0, 1)
	OddRowStyle   = EvenRowStyle.Foreground(lipgloss.Color("245"))
	TotalRowStyle = EvenRowStyle.Background(lipgloss.Color("0")).Foreground(lipgloss.Color("11"))
)