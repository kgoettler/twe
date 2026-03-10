package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBackground  = lipgloss.Color("238")
	ColorPrimaryText = lipgloss.Color("15")
	ColorMutedText   = lipgloss.Color("#8B93A6")
	ColorStructure   = lipgloss.Color("#3A3F4B")
	ColorAccent      = lipgloss.Color("12")
	ColorSuccess     = lipgloss.Color("10")
	ColorWarning     = lipgloss.Color("11")
	ColorError       = lipgloss.Color("9")
)

var (
	// Styles for table formatting
	BaseStyle     = lipgloss.NewStyle().Padding(0, 1).Foreground(ColorPrimaryText)
	BorderStyle   = lipgloss.NewStyle().Foreground(ColorStructure)
	HeaderStyle   = BaseStyle.Foreground(ColorPrimaryText)
	EvenRowStyle  = BaseStyle
	OddRowStyle   = BaseStyle
	TotalRowStyle = EvenRowStyle.Foreground(ColorAccent)
)
