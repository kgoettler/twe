package edit

import (
	"github.com/charmbracelet/lipgloss"

	styles "github.com/kgoettler/twe/internal/styles"
)

var (
	BaseStyle        = lipgloss.NewStyle().Foreground(styles.ColorPrimaryText)
	ErrStyle         = lipgloss.NewStyle().Foreground(styles.ColorError)
	FocusStyle       = HighlightStyle.Bold(true)
	HighlightStyle   = BaseStyle.Foreground(styles.ColorAccent)
	PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.ColorMutedText)

	// Table styles
	TableBorderStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(styles.ColorStructure)
	CellStyle        = TableBorderStyle.BorderRight(true).PaddingLeft(1)
)
