package edit

import (
	"github.com/charmbracelet/lipgloss"

	styles "github.com/kgoettler/twe/internal/styles"
)

type TextStyle struct {
	Base      lipgloss.Style
	Highlight lipgloss.Style
	Focus     lipgloss.Style
}

var (
	DefaultStyle     = lipgloss.NewStyle()
	PlaceholderStyle = lipgloss.NewStyle().Foreground(styles.ColorMutedText)
	// HighlightStyle    = lipgloss.NewStyle().Background(styles.ColorBackground)
	// FocusStyle        = lipgloss.NewStyle().Background(styles.ColorBackground).Foreground(styles.ColorAccent) //.Foreground(lipgloss.Color("11"))
	ErrStyle          = lipgloss.NewStyle().Foreground(styles.ColorError)
	TableBorderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(styles.ColorStructure)
	HeaderBorderStyle = TableBorderStyle.BorderBottom(true)
	CellBorderStyle   = TableBorderStyle.BorderRight(true)

	textBaseStyle      = lipgloss.NewStyle()
	textHighlightStyle = textBaseStyle.Background(styles.ColorBackground)
	textFocusStyle     = textHighlightStyle.Bold(true)

	timeBaseStyle      = lipgloss.NewStyle().Foreground(styles.ColorAccent)
	timeHighlightStyle = timeBaseStyle.Background(styles.ColorBackground)
	timeFocusStyle     = timeHighlightStyle.Bold(true)

)

var (
	DescStyle = TextStyle{
		Base:      textBaseStyle,
		Highlight: textHighlightStyle,
		Focus:     textFocusStyle,
	}
	TimeStyle = TextStyle{
		Base:      timeBaseStyle,
		Highlight: timeHighlightStyle,
		Focus:     timeFocusStyle,
	}
)