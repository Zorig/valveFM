package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	contentWidth := m.width - 4
	if contentWidth < 10 {
		contentWidth = m.width
	}

	compact := contentWidth < 60 || m.height < 24
	tiny := contentWidth < 44 || m.height < 18
	panelInnerWidth := innerWidthForPanel(contentWidth)

	header := m.renderHeader(contentWidth)
	dial := m.styles.Panel.Width(contentWidth).Render(m.renderDial(panelInnerWidth, compact, tiny))

	var meta string
	if tiny {
		meta = m.styles.InsetPanel.Width(contentWidth).Render(m.renderStationMetaCompact(contentWidth, true))
	} else if compact {
		meta = m.styles.InsetPanel.Width(contentWidth).Render(m.renderStationMetaCompact(contentWidth, false))
	} else {
		meta = m.styles.InsetPanel.Width(contentWidth).Render(m.renderStationMeta())
	}

	keyHints := m.styles.KeyHint.Width(contentWidth).Render(m.renderKeyHints(contentWidth))

	var errLine string
	if m.errMsg != "" {
		errLine = m.styles.Error.Width(contentWidth).Render(m.errMsg)
	}

	var prompt string
	if m.inputMode == inputLocation {
		prompt = m.styles.Panel.Width(contentWidth).Render(m.location.View())
	}
	if m.inputMode == inputSearch {
		prompt = m.styles.Panel.Width(contentWidth).Render(m.search.View())
	}

	appPadding := 2
	baseHeight := lipgloss.Height(header) + lipgloss.Height(dial) + lipgloss.Height(meta) + lipgloss.Height(keyHints)
	if errLine != "" {
		baseHeight += lipgloss.Height(errLine)
	}
	if prompt != "" {
		baseHeight += lipgloss.Height(prompt)
	}

	available := m.height - baseHeight - appPadding
	listPadding := 5
	if compact {
		listPadding = 4
	}
	if tiny {
		listPadding = 3
	}
	listItems := available - listPadding
	if listItems < 1 {
		listItems = 1
	}

	list := m.renderList(contentWidth, listItems)

	sections := []string{header, dial, meta, list, keyHints}
	if errLine != "" {
		sections = append(sections, errLine)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if prompt != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, prompt)
	}

	view := m.styles.App.Render(content)

	if m.showHelp {
		help := m.renderHelp()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, help)
	}
	if m.inputMode == inputCountrySelect {
		selector := m.renderCountrySelect(contentWidth, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, selector)
	}

	return view
}

func (m Model) renderHeader(width int) string {
	status := "STOPPED"
	statusStyle := m.styles.Muted
	if m.playing {
		status = "PLAYING"
		statusStyle = m.styles.Accent
	}

	left := "VALVE FM"
	if width >= 30 {
		left = fmt.Sprintf("VALVE FM [%s] FM STEREO", strings.ToUpper(m.country))
	} else if width >= 20 {
		left = fmt.Sprintf("VALVE FM [%s]", strings.ToUpper(m.country))
	}
	right := statusStyle.Render(status)
	line := joinHeader(left, right, width)
	return m.styles.Header.Width(width).Render(line)
}

func (m Model) renderDial(width int, compact bool, tiny bool) string {
	labels, bar, minor := buildDialScale(width, m.dialMin, m.dialMax, m.dialUseFreq)
	ptrLine := m.pointerLine(bar)
	freq, exact := m.selectedFrequency()
	freqPrefix := ""
	if !exact {
		freqPrefix = "~"
	}
	freqLine := fmt.Sprintf("%s%.1f MHz", freqPrefix, freq)

	lines := []string{}
	showLabels := width >= 28
	if !tiny {
		title := "FM BAND"
		if compact {
			title = "FM"
		}
		lines = append(lines, m.styles.DialLabel.Render(title))
	}
	if showLabels {
		lines = append(lines, m.styles.DialScale.Render(labels))
	}
	lines = append(lines, m.styles.DialScale.Render(bar))
	if showLabels && minor != "" {
		lines = append(lines, m.styles.DialScale.Render(minor))
	}
	lines = append(lines, m.styles.DialPointer.Render(ptrLine))
	lines = append(lines, m.styles.DialLabel.Render(freqLine))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) pointerLine(bar string) string {
	pos := max(m.pointerPosition(bar), 0)
	if pos >= len(bar) {
		pos = len(bar) - 1
	}
	return strings.Repeat(" ", pos) + "^"
}

func (m Model) pointerPosition(bar string) int {
	list := m.visibleStations()
	if len(list) <= 1 {
		return 0
	}
	var pos float64
	if m.dialUseFreq && m.dialMax > m.dialMin {
		pos = (m.dialPos - m.dialMin) / (m.dialMax - m.dialMin)
	} else {
		maxIndex := len(list) - 1
		pos = m.dialPos / float64(maxIndex)
	}
	return int(math.Round(pos * float64(len(bar)-1)))
}

func (m Model) selectedFrequency() (float64, bool) {
	list := m.visibleStations()
	if len(list) == 0 {
		return 98.0, false
	}
	if m.selected < 0 || m.selected >= len(list) {
		return 98.0, false
	}

	station := list[m.selected]
	if freq := station.Frequency.Float64(); freq > 0 {
		return freq, true
	}

	if m.dialUseFreq && m.dialMax > m.dialMin {
		return m.dialValueForIndex(m.selected), false
	}

	return m.linearFrequency(m.selected, len(list)), false
}

func (m Model) renderStationMeta() string {
	station, ok := m.currentStation()
	if !ok {
		return m.styles.Meta.Render("No stations available")
	}

	name := m.styles.StationName.Render(station.Name)
	tags := fmt.Sprintf("Tags: %s", fallback(station.Tags, "-"))
	bitrate := fmt.Sprintf("Bitrate: %d kbps", station.Bitrate)
	if station.Bitrate <= 0 {
		bitrate = "Bitrate: -"
	}
	status := "Status: STOPPED"
	if m.playing && station.UUID == m.playingUUID {
		status = "Status: LIVE"
	}
	country := fmt.Sprintf("Country: %s", fallback(station.Country, "-"))

	lines := []string{
		name,
		m.styles.Meta.Render(country),
		m.styles.Meta.Render(tags),
		m.styles.Meta.Render(bitrate),
		m.styles.Meta.Render(status),
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderStationMetaCompact(width int, tiny bool) string {
	station, ok := m.currentStation()
	if !ok {
		return m.styles.Meta.Render("No stations available")
	}

	status := "STOP"
	if m.playing && station.UUID == m.playingUUID {
		status = "LIVE"
	}

	name := truncateText(station.Name, max(width-8, 10))
	if tiny {
		line := fmt.Sprintf("Now: %s [%s]", name, status)
		return m.styles.Meta.Render(line)
	}

	line1 := m.styles.StationName.Render(name)
	meta := fmt.Sprintf("Tags: %s | %s", fallback(station.Tags, "-"), status)
	meta = truncateText(meta, max(width-6, 12))
	line2 := m.styles.Meta.Render(meta)
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
}

func (m Model) renderList(width int, maxItems int) string {
	list := m.visibleStations()
	lines := []string{m.styles.ListHeader.Render("Stations")}

	if m.loading {
		lines = append(lines, m.styles.Muted.Render("Loading stations..."))
		return m.styles.Panel.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	if len(list) == 0 {
		lines = append(lines, m.styles.Muted.Render("No stations found"))
		return m.styles.Panel.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	start, end := listWindow(len(list), m.selected, maxItems)
	lineWidth := innerWidthForPanel(width)
	if lineWidth < 10 {
		lineWidth = max(width-2, 4)
	}
	if lineWidth > width {
		lineWidth = width
	}
	showFreq := lineWidth >= 32

	for i := start; i < end; i++ {
		station := list[i]
		marker := "  "
		style := m.styles.ListItem
		if i == m.selected {
			marker = "> "
			style = m.styles.ListActive
		}

		fav := ""
		if m.favorites != nil && m.favorites.IsFavorite(station.UUID) {
			fav = " *"
		}

		name := station.Name
		if showFreq {
			freq, exact := m.listFrequency(i, len(list), station.Frequency.Float64())
			prefix := " "
			if !exact {
				prefix = "~"
			}
			reserved := 2 + 1 + 5 + 1 + len(fav)
			nameWidth := max(lineWidth-reserved, 4)
			name = truncateText(name, nameWidth)
			line := fmt.Sprintf("%s%s%5.1f %s%s", marker, prefix, freq, name, fav)
			lines = append(lines, style.Width(lineWidth).MaxWidth(lineWidth).Render(line))
			continue
		}

		reserved := 2 + len(fav)
		nameWidth := max(lineWidth-reserved, 4)
		name = truncateText(name, nameWidth)
		line := fmt.Sprintf("%s%s%s", marker, name, fav)
		lines = append(lines, style.Width(lineWidth).MaxWidth(lineWidth).Render(line))
	}

	return m.styles.Panel.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderKeyHints(width int) string {
	if width < 30 {
		return "Enter Play  Q Quit"
	}
	if width < 44 {
		return "Arrows Tune  Enter Play  Space Stop  Q Quit"
	}
	if width < 62 {
		return "Arrows Tune  Enter Play  Space Stop  L Country  / Search  ? Help  Q Quit"
	}
	return "Arrows Tune  Up/Down Browse  Enter Play  Space Stop  L Country  / Search  F Favorite  ? Help  Q Quit"
}

func (m Model) renderHelp() string {
	lines := []string{
		"Controls",
		"",
		"Left/Right   Tune dial",
		"Up/Down      Browse list",
		"Enter        Play station",
		"Space        Stop/Resume",
		"L            Choose country",
		"/            Search stations",
		"F            Favorite station",
		"?            Close help",
		"Q            Quit",
	}
	return m.styles.HelpBox.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderCountrySelect(width int, height int) string {
	panelWidth := width
	if panelWidth <= 0 {
		panelWidth = 10
	}
	innerWidth := innerWidthForPanel(panelWidth)
	if innerWidth < 10 {
		innerWidth = max(panelWidth-2, 4)
	}

	search := m.countrySearch.View()
	if m.countryLoading {
		lines := []string{
			m.styles.ListHeader.Render("Select Country"),
			m.styles.Muted.Render("Loading countries..."),
		}
		return m.styles.Panel.Width(panelWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	list := m.visibleCountries()
	lines := []string{
		m.styles.ListHeader.Render("Select Country"),
		m.styles.Meta.Render(search),
	}

	maxItems := max(height-12, 4)
	if maxItems > 12 {
		maxItems = 12
	}
	if len(list) == 0 {
		lines = append(lines, m.styles.Muted.Render("No matches"))
		return m.styles.Panel.Width(panelWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	start, end := listWindow(len(list), m.countryIndex, maxItems)
	for i := start; i < end; i++ {
		country := list[i]
		marker := "  "
		style := m.styles.ListItem
		if i == m.countryIndex {
			marker = "> "
			style = m.styles.ListActive
		}
		label := fmt.Sprintf("%s - %s", country.Code, country.Name)
		label = truncateText(label, innerWidth)
		lines = append(lines, style.Width(innerWidth).MaxWidth(innerWidth).Render(marker+label))
	}

	return m.styles.Panel.Width(panelWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func joinHeader(left, right string, width int) string {
	if width <= 0 {
		return ""
	}

	rightWidth := lipgloss.Width(right)
	if rightWidth >= width {
		return truncateText(right, width)
	}

	maxLeft := width - rightWidth - 1
	left = truncateText(left, maxLeft)
	space := max(width-lipgloss.Width(left)-rightWidth, 1)
	return left + strings.Repeat(" ", space) + right
}

func listWindow(length, selected, max int) (int, int) {
	if length <= max {
		return 0, length
	}
	start := selected - max/2
	if start < 0 {
		start = 0
	}
	end := start + max
	if end > length {
		end = length
		start = end - max
	}
	if start < 0 {
		start = 0
	}
	return start, end
}

func fallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return value
}

func (m Model) listFrequency(index, total int, stationFreq float64) (float64, bool) {
	if stationFreq > 0 {
		return stationFreq, true
	}
	if total <= 1 {
		return 98.0, false
	}
	if m.dialUseFreq && m.dialMax > m.dialMin {
		frac := float64(index) / float64(total-1)
		return m.dialMin + (m.dialMax-m.dialMin)*frac, false
	}
	return m.linearFrequency(index, total), false
}

func (m Model) linearFrequency(index, total int) float64 {
	if total <= 1 {
		return 98.0
	}
	min := 88.0
	max := 108.0
	frac := float64(index) / float64(total-1)
	return min + (max-min)*frac
}

func buildDialScale(width int, min, max float64, dynamic bool) (string, string, string) {
	if width < 10 {
		width = 10
	}
	if dynamic && max > min {
		return buildDialScaleDynamic(width, min, max)
	}
	return buildDialScaleFixed(width)
}

func buildDialScaleFixed(width int) (string, string, string) {
	labels := []string{"88", "90", "92", "94", "96", "98", "100", "102", "104", "106", "108"}
	return buildDialScaleLabels(width, labels)
}

func buildDialScaleDynamic(width int, min, max float64) (string, string, string) {
	labelCount := 5
	if width < 26 {
		labelCount = 3
	} else if width < 40 {
		labelCount = 4
	}

	labels := make([]string, labelCount)
	for i := 0; i < labelCount; i++ {
		frac := float64(i) / float64(labelCount-1)
		labels[i] = formatFreqLabel(min + (max-min)*frac)
	}
	return buildDialScaleLabels(width, labels)
}

func buildDialScaleLabels(width int, labels []string) (string, string, string) {
	bar := make([]rune, width)
	minor := make([]rune, width)
	labelLine := make([]rune, width)
	for i := 0; i < width; i++ {
		bar[i] = '-'
		minor[i] = ' '
		labelLine[i] = ' '
	}

	positions := make([]int, len(labels))
	for i := range labels {
		pos := int(math.Round(float64(i) / float64(len(labels)-1) * float64(width-1)))
		positions[i] = pos
		bar[pos] = '|'
	}

	if width >= 60 {
		for i := 0; i < len(positions)-1; i++ {
			mid := (positions[i] + positions[i+1]) / 2
			if mid >= 0 && mid < width {
				minor[mid] = ':'
			}
		}
	}

	for i := range labels {
		placeLabel(labelLine, positions[i], labels[i])
	}

	minorLine := ""
	if width >= 60 {
		minorLine = string(minor)
	}
	return string(labelLine), string(bar), minorLine
}

func placeLabel(line []rune, pos int, label string) {
	runes := []rune(label)
	start := pos - len(runes)/2
	if start < 0 {
		start = 0
	}
	if start+len(runes) > len(line) {
		start = len(line) - len(runes)
	}
	if start < 0 {
		return
	}
	for i := 0; i < len(runes); i++ {
		if line[start+i] != ' ' {
			return
		}
	}
	copy(line[start:], runes)
}

func formatFreqLabel(value float64) string {
	rounded := math.Round(value)
	if math.Abs(value-rounded) < 0.05 {
		return fmt.Sprintf("%.0f", rounded)
	}
	return fmt.Sprintf("%.1f", value)
}

func truncateText(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func innerWidthForPanel(width int) int {
	inner := width - 6
	if inner < 4 {
		inner = max(width-2, 2)
	}
	if inner > width {
		inner = width
	}
	return inner
}
