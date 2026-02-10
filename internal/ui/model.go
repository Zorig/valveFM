package ui

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"radio-tui/internal/config"
	"radio-tui/internal/player"
	"radio-tui/internal/radio"
)

type inputMode int

const (
	inputNone inputMode = iota
	inputLocation
	inputSearch
	inputCountrySelect
)

type Model struct {
	api       *radio.Client
	player    *player.Player
	favorites *config.Favorites
	styles    Styles

	stations []radio.Station
	filtered []radio.Station
	selected int

	loading bool
	errMsg  string

	country string

	inputMode     inputMode
	location      textinput.Model
	search        textinput.Model
	countrySearch textinput.Model

	showHelp bool

	width  int
	height int

	playing     bool
	playingUUID string
	lastStation radio.Station

	dialPos     float64
	dialTarget  float64
	dialMin     float64
	dialMax     float64
	dialUseFreq bool

	countries         []radio.Country
	filteredCountries []radio.Country
	countryIndex      int
	countryLoading    bool
}

type stationsMsg struct {
	stations []radio.Station
	err      error
}

type countriesMsg struct {
	countries []radio.Country
	err       error
}

type playMsg struct {
	station radio.Station
	url     string
	err     error
}

type dialTickMsg struct{}

func NewModel(api *radio.Client, player *player.Player, favorites *config.Favorites, playerErr error, favErr error) Model {
	location := textinput.New()
	location.Prompt = "Country: "
	location.Placeholder = "US"
	location.CharLimit = 2
	location.Width = 6

	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "Search stations"
	search.Width = 26

	countrySearch := textinput.New()
	countrySearch.Prompt = "Search: "
	countrySearch.Placeholder = "Type country or code"
	countrySearch.Width = 26

	m := Model{
		api:           api,
		player:        player,
		favorites:     favorites,
		styles:        DefaultStyles(),
		country:       "US",
		location:      location,
		search:        search,
		countrySearch: countrySearch,
		loading:       true,
	}

	if playerErr != nil {
		m.errMsg = playerErr.Error()
	}
	if favErr != nil {
		if m.errMsg != "" {
			m.errMsg += " | " + favErr.Error()
		} else {
			m.errMsg = favErr.Error()
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.loadStationsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateInputWidths()
		return m, nil
	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" || key == "q" {
			if m.player != nil {
				_ = m.player.Stop()
			}
			return m, tea.Quit
		}

		if m.showHelp {
			if key == "?" || key == "esc" || key == "enter" {
				m.showHelp = false
			}
			return m, nil
		}

		switch m.inputMode {
		case inputLocation:
			return m.updateLocationInput(msg)
		case inputSearch:
			return m.updateSearchInput(msg)
		case inputCountrySelect:
			return m.updateCountrySelect(msg)
		}

		switch key {
		case "?":
			m.showHelp = true
		case "left":
			if m.moveSelection(-1) {
				return m, m.dialTickCmd()
			}
		case "right":
			if m.moveSelection(1) {
				return m, m.dialTickCmd()
			}
		case "up":
			if m.moveSelection(-1) {
				return m, m.dialTickCmd()
			}
		case "down":
			if m.moveSelection(1) {
				return m, m.dialTickCmd()
			}
		case "enter":
			if station, ok := m.currentStation(); ok {
				m.errMsg = ""
				return m, m.playStationCmd(station)
			}
		case " ":
			if m.playing {
				if m.player != nil {
					_ = m.player.Stop()
				}
				m.playing = false
				return m, nil
			}
			if m.lastStation.UUID != "" {
				return m, m.playStationCmd(m.lastStation)
			}
			return m, nil
		case "L", "l":
			m.inputMode = inputCountrySelect
			m.countrySearch.SetValue("")
			m.countrySearch.Focus()
			m.countrySearch.CursorEnd()
			if len(m.countries) == 0 && !m.countryLoading {
				m.countryLoading = true
				return m, m.loadCountriesCmd()
			}
			m.applyCountryFilter()
			m.ensureCountrySelection()
			return m, textinput.Blink
		case "/":
			m.inputMode = inputSearch
			m.search.Focus()
			m.search.CursorEnd()
			return m, textinput.Blink
		case "f", "F":
			if m.favorites != nil {
				if station, ok := m.currentStation(); ok {
					_, err := m.favorites.Toggle(station)
					if err != nil {
						m.errMsg = err.Error()
					}
				}
			}
		}
	case stationsMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.stations = nil
			m.filtered = nil
			m.selected = 0
			return m, nil
		}
		m.errMsg = ""
		m.stations = msg.stations
		m.selected = 0
		m.applyFilter()
		m.ensureSelection()
		m.updateDialRange()
		m.snapDial()
		return m, nil
	case countriesMsg:
		m.countryLoading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.countries = nil
			m.filteredCountries = nil
			m.countryIndex = 0
			return m, nil
		}
		m.errMsg = ""
		m.countries = msg.countries
		m.applyCountryFilter()
		m.ensureCountrySelection()
		return m, nil
	case playMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		if m.player == nil {
			m.errMsg = "audio player not available"
			return m, nil
		}
		if err := m.player.Play(msg.url); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.errMsg = ""
		m.playing = true
		m.playingUUID = msg.station.UUID
		m.lastStation = msg.station
		return m, nil
	case dialTickMsg:
		return m.updateDialAnimation()
	}

	return m, nil
}

func (m Model) loadStationsCmd() tea.Cmd {
	country := m.country
	api := m.api
	return func() tea.Msg {
		if api == nil {
			return stationsMsg{err: fmt.Errorf("radio api not available")}
		}
		stations, err := api.StationsByCountry(context.Background(), country)
		return stationsMsg{stations: stations, err: err}
	}
}

func (m Model) playStationCmd(station radio.Station) tea.Cmd {
	api := m.api
	return func() tea.Msg {
		if api == nil {
			return playMsg{err: fmt.Errorf("radio api not available")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		streamURL, err := api.ResolveStationURL(ctx, station.UUID)
		return playMsg{station: station, url: streamURL, err: err}
	}
}

func (m Model) loadCountriesCmd() tea.Cmd {
	api := m.api
	return func() tea.Msg {
		if api == nil {
			return countriesMsg{err: fmt.Errorf("radio api not available")}
		}
		countries, err := api.Countries(context.Background())
		return countriesMsg{countries: countries, err: err}
	}
}

func (m Model) dialTickCmd() tea.Cmd {
	if math.Abs(m.dialPos-m.dialTarget) < 0.01 {
		return nil
	}
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return dialTickMsg{}
	})
}

func (m Model) updateDialAnimation() (tea.Model, tea.Cmd) {
	diff := m.dialTarget - m.dialPos
	if math.Abs(diff) < 0.05 {
		m.dialPos = m.dialTarget
		return m, nil
	}
	step := math.Copysign(0.4, diff)
	if math.Abs(step) > math.Abs(diff) {
		step = diff
	}
	m.dialPos += step
	return m, m.dialTickCmd()
}

func (m Model) updateLocationInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.location, cmd = m.location.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			m.inputMode = inputNone
			m.location.Blur()
			m.country = strings.ToUpper(strings.TrimSpace(m.location.Value()))
			if m.country == "" {
				m.country = "US"
			}
			m.loading = true
			return m, m.loadStationsCmd()
		case "esc":
			m.inputMode = inputNone
			m.location.Blur()
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) updateSearchInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.applyFilter()
	m.ensureSelection()
	m.updateDialRange()
	m.snapDial()

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			m.inputMode = inputNone
			m.search.Blur()
			return m, nil
		case "esc":
			m.inputMode = inputNone
			m.search.SetValue("")
			m.search.Blur()
			m.applyFilter()
			m.ensureSelection()
			m.updateDialRange()
			m.snapDial()
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) updateCountrySelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.countrySearch, cmd = m.countrySearch.Update(msg)
	m.applyCountryFilter()
	m.ensureCountrySelection()

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up":
			m.moveCountrySelection(-1)
		case "down":
			m.moveCountrySelection(1)
		case "enter":
			if country, ok := m.currentCountry(); ok {
				m.inputMode = inputNone
				m.countrySearch.Blur()
				m.countrySearch.SetValue("")
				m.country = strings.ToUpper(strings.TrimSpace(country.Code))
				m.loading = true
				return m, m.loadStationsCmd()
			}
		case "esc":
			m.inputMode = inputNone
			m.countrySearch.Blur()
			m.countrySearch.SetValue("")
			m.applyCountryFilter()
			m.ensureCountrySelection()
			return m, nil
		}
	}

	return m, cmd
}

func (m *Model) applyFilter() {
	filter := strings.TrimSpace(strings.ToLower(m.search.Value()))
	if filter == "" {
		m.filtered = nil
		return
	}

	filtered := make([]radio.Station, 0, len(m.stations))
	for _, station := range m.stations {
		name := strings.ToLower(station.Name)
		tags := strings.ToLower(station.Tags)
		if strings.Contains(name, filter) || strings.Contains(tags, filter) {
			filtered = append(filtered, station)
		}
	}
	if len(filtered) == 0 {
		m.selected = 0
	}
	m.filtered = filtered
}

func (m *Model) applyCountryFilter() {
	filter := strings.TrimSpace(strings.ToLower(m.countrySearch.Value()))
	if filter == "" {
		m.filteredCountries = nil
		return
	}

	filtered := make([]radio.Country, 0, len(m.countries))
	for _, country := range m.countries {
		name := strings.ToLower(country.Name)
		code := strings.ToLower(country.Code)
		if strings.Contains(name, filter) || strings.Contains(code, filter) {
			filtered = append(filtered, country)
		}
	}
	if len(filtered) == 0 {
		m.countryIndex = 0
	}
	m.filteredCountries = filtered
}

func (m *Model) ensureSelection() {
	list := m.visibleStations()
	if len(list) == 0 {
		m.selected = 0
		return
	}
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(list) {
		m.selected = len(list) - 1
	}
}

func (m *Model) ensureCountrySelection() {
	list := m.visibleCountries()
	if len(list) == 0 {
		m.countryIndex = 0
		return
	}
	if m.countryIndex < 0 {
		m.countryIndex = 0
	}
	if m.countryIndex >= len(list) {
		m.countryIndex = len(list) - 1
	}
}

func (m *Model) moveSelection(delta int) bool {
	list := m.visibleStations()
	if len(list) == 0 {
		return false
	}
	prev := m.selected
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(list) {
		m.selected = len(list) - 1
	}
	m.dialTarget = m.dialValueForIndex(m.selected)
	return prev != m.selected
}

func (m *Model) moveCountrySelection(delta int) {
	list := m.visibleCountries()
	if len(list) == 0 {
		return
	}
	m.countryIndex += delta
	if m.countryIndex < 0 {
		m.countryIndex = 0
	}
	if m.countryIndex >= len(list) {
		m.countryIndex = len(list) - 1
	}
}

func (m *Model) currentStation() (radio.Station, bool) {
	list := m.visibleStations()
	if len(list) == 0 {
		return radio.Station{}, false
	}
	if m.selected < 0 || m.selected >= len(list) {
		return radio.Station{}, false
	}
	return list[m.selected], true
}

func (m *Model) currentCountry() (radio.Country, bool) {
	list := m.visibleCountries()
	if len(list) == 0 {
		return radio.Country{}, false
	}
	if m.countryIndex < 0 || m.countryIndex >= len(list) {
		return radio.Country{}, false
	}
	return list[m.countryIndex], true
}

func (m *Model) visibleStations() []radio.Station {
	if strings.TrimSpace(m.search.Value()) != "" {
		return m.filtered
	}
	return m.stations
}

func (m *Model) visibleCountries() []radio.Country {
	if strings.TrimSpace(m.countrySearch.Value()) != "" {
		return m.filteredCountries
	}
	return m.countries
}

func (m *Model) snapDial() {
	m.dialTarget = m.dialValueForIndex(m.selected)
	m.dialPos = m.dialTarget
}

func (m *Model) updateDialRange() {
	list := m.visibleStations()
	if len(list) == 0 {
		m.dialUseFreq = false
		m.dialMin = 0
		m.dialMax = 0
		return
	}

	min := math.MaxFloat64
	max := 0.0
	count := 0
	for _, station := range list {
		freq := station.Frequency.Float64()
		if freq > 0 {
			count++
			if freq < min {
				min = freq
			}
			if freq > max {
				max = freq
			}
		}
	}

	if count >= 2 && max > min {
		m.dialUseFreq = true
		m.dialMin = min
		m.dialMax = max
		return
	}

	m.dialUseFreq = false
	m.dialMin = 0
	m.dialMax = 0
}

func (m Model) dialValueForIndex(index int) float64 {
	list := m.visibleStations()
	if index < 0 || index >= len(list) {
		return 0
	}

	if m.dialUseFreq {
		if freq := list[index].Frequency.Float64(); freq > 0 {
			return freq
		}
		if len(list) > 1 && m.dialMax > m.dialMin {
			frac := float64(index) / float64(len(list)-1)
			return m.dialMin + (m.dialMax-m.dialMin)*frac
		}
	}

	return float64(index)
}

func (m *Model) updateInputWidths() {
	width := m.width - 20
	if width < 10 {
		width = 10
	}
	if width > 32 {
		width = 32
	}
	if width > 0 {
		m.search.Width = width
		m.countrySearch.Width = width
	}
}
