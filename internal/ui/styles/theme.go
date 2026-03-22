package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Margin(1, 1)

	// Block Card
	BaseBlockBorderStyle = lipgloss.
				NewStyle().
				Foreground(BlockBorderColor)

	BaseBlockStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BlockBorderColor).
			Padding(0, 1)

	BlockTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240")).
			Padding(0, 1)

	// Input Field
	InputPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	// Status Bar
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)
)
