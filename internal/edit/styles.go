package edit

import "github.com/charmbracelet/lipgloss"

var (
	DefaultStyle      = lipgloss.NewStyle()
	PlaceholderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	HighlightStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	FocusStyle        = lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("212")).Bold(true) //.Foreground(lipgloss.Color("11"))
	ErrStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	TableBorderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))
	HeaderBorderStyle = TableBorderStyle.BorderBottom(true)
	CellBorderStyle   = TableBorderStyle.BorderRight(true)
)
