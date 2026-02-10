package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	App         lipgloss.Style
	Header      lipgloss.Style
	Panel       lipgloss.Style
	InsetPanel  lipgloss.Style
	DialLabel   lipgloss.Style
	DialScale   lipgloss.Style
	DialPointer lipgloss.Style
	StationName lipgloss.Style
	Meta        lipgloss.Style
	ListHeader  lipgloss.Style
	ListItem    lipgloss.Style
	ListActive  lipgloss.Style
	KeyHint     lipgloss.Style
	HelpBox     lipgloss.Style
	Error       lipgloss.Style
	Accent      lipgloss.Style
	Muted       lipgloss.Style
}

func DefaultStyles() Styles {
	cream := lipgloss.Color("#F5E6C8")
	amber := lipgloss.Color("#D9A441")
	brown := lipgloss.Color("#6E4A2F")
	deep := lipgloss.Color("#2B1A12")
	green := lipgloss.Color("#6A8F4E")
	muted := lipgloss.Color("#B89C7A")

	border := lipgloss.RoundedBorder()

	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(cream).
			Background(deep),
		Header: lipgloss.NewStyle().
			Foreground(cream).
			Background(brown).
			Padding(0, 1).
			Bold(true),
		Panel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(brown).
			Padding(1, 2),
		InsetPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(amber).
			Padding(1, 2),
		DialLabel: lipgloss.NewStyle().
			Foreground(amber).
			Bold(true),
		DialScale: lipgloss.NewStyle().
			Foreground(cream),
		DialPointer: lipgloss.NewStyle().
			Foreground(green).
			Bold(true),
		StationName: lipgloss.NewStyle().
			Foreground(cream).
			Bold(true),
		Meta: lipgloss.NewStyle().
			Foreground(muted),
		ListHeader: lipgloss.NewStyle().
			Foreground(amber).
			Bold(true),
		ListItem: lipgloss.NewStyle().
			Foreground(cream),
		ListActive: lipgloss.NewStyle().
			Foreground(deep).
			Background(amber).
			Bold(true),
		KeyHint: lipgloss.NewStyle().
			Foreground(muted),
		HelpBox: lipgloss.NewStyle().
			Border(border).
			BorderForeground(amber).
			Padding(1, 2).
			Background(brown).
			Foreground(cream),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F29F8E")).
			Bold(true),
		Accent: lipgloss.NewStyle().
			Foreground(amber),
		Muted: lipgloss.NewStyle().
			Foreground(muted),
	}
}
