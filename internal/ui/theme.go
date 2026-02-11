package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines a set of semantic colors used to build the UI styles.
type Theme struct {
	Name      string
	Slug      string
	Fg        string // primary text, dial scale, station names
	Accent    string // labels, active items, inset borders
	Secondary string // header bg, panel borders, help bg
	Bg        string // app background, active item text
	Success   string // dial pointer
	Muted     string // hints, metadata
	Error     string // error messages
}

// Themes is the ordered list of all built-in themes.
var Themes = []Theme{
	themeVintage(),
	themeTokyoNight(),
	themeNord(),
	themeCatppuccinMocha(),
	themeCatppuccinLatte(),
	themeGruvboxDark(),
	themeDracula(),
	themeSolarizedDark(),
	themeOneDark(),
	themeRosePine(),
	themeKanagawa(),
	themeEverforest(),
}

// ThemeBySlug returns the theme with the given slug, falling back to Vintage.
func ThemeBySlug(slug string) Theme {
	for _, t := range Themes {
		if t.Slug == slug {
			return t
		}
	}
	return Themes[0]
}

// BuildStyles constructs the full Styles set from a theme.
func BuildStyles(t Theme) Styles {
	fg := lipgloss.Color(t.Fg)
	accent := lipgloss.Color(t.Accent)
	secondary := lipgloss.Color(t.Secondary)
	bg := lipgloss.Color(t.Bg)
	success := lipgloss.Color(t.Success)
	muted := lipgloss.Color(t.Muted)
	errColor := lipgloss.Color(t.Error)

	border := lipgloss.RoundedBorder()

	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(fg).
			Background(bg),
		Header: lipgloss.NewStyle().
			Foreground(fg).
			Background(secondary).
			Padding(0, 1).
			Bold(true),
		Panel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(secondary).
			Padding(1, 2),
		InsetPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(accent).
			Padding(1, 2),
		DialLabel: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		DialScale: lipgloss.NewStyle().
			Foreground(fg),
		DialPointer: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		StationName: lipgloss.NewStyle().
			Foreground(fg).
			Bold(true),
		Meta: lipgloss.NewStyle().
			Foreground(muted),
		ListHeader: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		ListItem: lipgloss.NewStyle().
			Foreground(fg),
		ListActive: lipgloss.NewStyle().
			Foreground(bg).
			Background(accent).
			Bold(true),
		KeyHint: lipgloss.NewStyle().
			Foreground(muted),
		HelpBox: lipgloss.NewStyle().
			Border(border).
			BorderForeground(accent).
			Padding(1, 2).
			Background(secondary).
			Foreground(fg),
		Error: lipgloss.NewStyle().
			Foreground(errColor).
			Bold(true),
		Accent: lipgloss.NewStyle().
			Foreground(accent),
		Muted: lipgloss.NewStyle().
			Foreground(muted),
	}
}

func themeVintage() Theme {
	return Theme{
		Name:      "Vintage",
		Slug:      "vintage",
		Fg:        "#F5E6C8",
		Accent:    "#D9A441",
		Secondary: "#6E4A2F",
		Bg:        "#2B1A12",
		Success:   "#6A8F4E",
		Muted:     "#B89C7A",
		Error:     "#F29F8E",
	}
}

func themeTokyoNight() Theme {
	return Theme{
		Name:      "Tokyo Night",
		Slug:      "tokyo-night",
		Fg:        "#C0CAF5",
		Accent:    "#7AA2F7",
		Secondary: "#24283B",
		Bg:        "#1A1B26",
		Success:   "#9ECE6A",
		Muted:     "#565F89",
		Error:     "#F7768E",
	}
}

func themeNord() Theme {
	return Theme{
		Name:      "Nord",
		Slug:      "nord",
		Fg:        "#ECEFF4",
		Accent:    "#88C0D0",
		Secondary: "#3B4252",
		Bg:        "#2E3440",
		Success:   "#A3BE8C",
		Muted:     "#4C566A",
		Error:     "#BF616A",
	}
}

func themeCatppuccinMocha() Theme {
	return Theme{
		Name:      "Catppuccin Mocha",
		Slug:      "catppuccin-mocha",
		Fg:        "#CDD6F4",
		Accent:    "#CBA6F7",
		Secondary: "#313244",
		Bg:        "#1E1E2E",
		Success:   "#A6E3A1",
		Muted:     "#6C7086",
		Error:     "#F38BA8",
	}
}

func themeCatppuccinLatte() Theme {
	return Theme{
		Name:      "Catppuccin Latte",
		Slug:      "catppuccin-latte",
		Fg:        "#4C4F69",
		Accent:    "#8839EF",
		Secondary: "#CCD0DA",
		Bg:        "#EFF1F5",
		Success:   "#40A02B",
		Muted:     "#9CA0B0",
		Error:     "#D20F39",
	}
}

func themeGruvboxDark() Theme {
	return Theme{
		Name:      "Gruvbox Dark",
		Slug:      "gruvbox-dark",
		Fg:        "#EBDBB2",
		Accent:    "#FABD2F",
		Secondary: "#3C3836",
		Bg:        "#282828",
		Success:   "#B8BB26",
		Muted:     "#928374",
		Error:     "#FB4934",
	}
}

func themeDracula() Theme {
	return Theme{
		Name:      "Dracula",
		Slug:      "dracula",
		Fg:        "#F8F8F2",
		Accent:    "#BD93F9",
		Secondary: "#44475A",
		Bg:        "#282A36",
		Success:   "#50FA7B",
		Muted:     "#6272A4",
		Error:     "#FF5555",
	}
}

func themeSolarizedDark() Theme {
	return Theme{
		Name:      "Solarized Dark",
		Slug:      "solarized-dark",
		Fg:        "#839496",
		Accent:    "#B58900",
		Secondary: "#073642",
		Bg:        "#002B36",
		Success:   "#859900",
		Muted:     "#586E75",
		Error:     "#DC322F",
	}
}

func themeOneDark() Theme {
	return Theme{
		Name:      "One Dark",
		Slug:      "one-dark",
		Fg:        "#ABB2BF",
		Accent:    "#61AFEF",
		Secondary: "#3E4452",
		Bg:        "#282C34",
		Success:   "#98C379",
		Muted:     "#5C6370",
		Error:     "#E06C75",
	}
}

func themeRosePine() Theme {
	return Theme{
		Name:      "Rose Pine",
		Slug:      "rose-pine",
		Fg:        "#E0DEF4",
		Accent:    "#C4A7E7",
		Secondary: "#26233A",
		Bg:        "#191724",
		Success:   "#9CCFD8",
		Muted:     "#6E6A86",
		Error:     "#EB6F92",
	}
}

func themeKanagawa() Theme {
	return Theme{
		Name:      "Kanagawa",
		Slug:      "kanagawa",
		Fg:        "#DCD7BA",
		Accent:    "#7E9CD8",
		Secondary: "#2A2A37",
		Bg:        "#1F1F28",
		Success:   "#98BB6C",
		Muted:     "#727169",
		Error:     "#E82424",
	}
}

func themeEverforest() Theme {
	return Theme{
		Name:      "Everforest",
		Slug:      "everforest",
		Fg:        "#D3C6AA",
		Accent:    "#A7C080",
		Secondary: "#374145",
		Bg:        "#2D353B",
		Success:   "#83C092",
		Muted:     "#859289",
		Error:     "#E67E80",
	}
}
