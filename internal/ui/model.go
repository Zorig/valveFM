package ui

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"radio-tui/internal/config"
	"radio-tui/internal/player"
	"radio-tui/internal/radio"
)

type inputMode int

type stationSource int

const (
	inputNone inputMode = iota
	inputLocation
	inputSearch
	inputCountrySelect

	stationPageSize = 200
)

const (
	sourceCountry stationSource = iota
	sourceFavorites
)

type Model struct {
	api       *radio.Client
	player    player.Backend
	favorites *config.Favorites
	styles    Styles
	ipc       *ipcServer

	stations []radio.Station
	selected int

	loading bool
	errMsg  string

	country string
	page    int
	hasMore bool

	stationSource stationSource
	activeSearch string

	inputMode     inputMode
	location      textinput.Model
	search        textinput.Model
	countrySearch textinput.Model

	showHelp  bool
	showTheme bool
	themeIdx  int
	theme     Theme

	width  int
	height int

	playing           bool
	playingUUID       string
	lastStation       radio.Station
	missingPlayer     bool
	downloadingPlayer bool

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
	source   stationSource
	page     int
	country  string
	search   string
	hasMore  bool
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

type playerDownloadMsg struct {
	path string
	err  error
}

type themeSavedMsg struct{ err error }

func NewModel(api *radio.Client, player player.Backend, favorites *config.Favorites, playerErr error, favErr error, themeName string) Model {
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

	theme := ThemeBySlug(themeName)
	themeIdx := 0
	for i, t := range Themes {
		if t.Slug == theme.Slug {
			themeIdx = i
			break
		}
	}

	m := Model{
		api:           api,
		player:        player,
		favorites:     favorites,
		styles:        BuildStyles(theme),
		theme:         theme,
		themeIdx:      themeIdx,
		country:       "US",
		stationSource: sourceCountry,
		location:      location,
		search:        search,
		countrySearch: countrySearch,
		loading:       true,
	}
	if favorites != nil && favorites.Count() > 0 {
		m.stationSource = sourceFavorites
	}

	if playerErr != nil {
		m.missingPlayer = true
		if runtime.GOOS == "windows" {
			m.downloadingPlayer = true
			m.errMsg = "Audio player not found. Downloading ffplay in the background..."
		} else {
			m.errMsg = "Audio player not found. Install mpv or ffplay and ensure it is in PATH."
		}
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
	return tea.Batch(m.loadStationsCmd(), m.startIPCCmd(), m.maybeDownloadPlayerCmd())
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
			if m.ipc != nil {
				m.ipc.Close()
			}
			return m, tea.Quit
		}

		if m.showHelp {
			if key == "?" || key == "esc" || key == "enter" {
				m.showHelp = false
			}
			return m, nil
		}

		if m.showTheme {
			switch key {
			case "t", "T", "esc":
				m.showTheme = false
			case "up", "k":
				if m.themeIdx > 0 {
					m.themeIdx--
					m.theme = Themes[m.themeIdx]
					m.styles = BuildStyles(m.theme)
				}
			case "down", "j":
				if m.themeIdx < len(Themes)-1 {
					m.themeIdx++
					m.theme = Themes[m.themeIdx]
					m.styles = BuildStyles(m.theme)
				}
			case "enter":
				m.showTheme = false
				return m, m.saveThemeCmd()
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
		case "]", "pgdown":
			if m.loading {
				return m, nil
			}
			if !m.hasMore {
				m.errMsg = "No more stations on the next page"
				return m, nil
			}
			m.page++
			m.selected = 0
			m.loading = true
			m.errMsg = ""
			return m, m.loadStationsCmd()
		case "[", "pgup":
			if m.loading {
				return m, nil
			}
			if m.page == 0 {
				m.errMsg = "Already on the first page"
				return m, nil
			}
			m.page--
			m.selected = 0
			m.loading = true
			m.errMsg = ""
			return m, m.loadStationsCmd()
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
		case "V", "v":
			if m.stationSource == sourceFavorites {
				// Toggle back to country list
				m.stationSource = sourceCountry
				m.activeSearch = ""
				m.search.SetValue("")
				m.page = 0
				m.hasMore = false
				m.selected = 0
				m.loading = true
				m.errMsg = ""
				return m, m.loadStationsCmd()
			}
			if m.favorites == nil || m.favorites.Count() == 0 {
				m.errMsg = "No favorites saved yet"
				return m, nil
			}
			m.stationSource = sourceFavorites
			m.activeSearch = ""
			m.search.SetValue("")
			m.page = 0
			m.hasMore = false
			m.selected = 0
			m.loading = true
			m.errMsg = ""
			return m, m.loadStationsCmd()
		case "/":
			// If in favorites view, switch to full station list before searching
			if m.stationSource == sourceFavorites {
				m.stationSource = sourceCountry
				m.activeSearch = ""
			}
			m.inputMode = inputSearch
			m.search.SetValue(m.activeSearch)
			m.search.Focus()
			m.search.CursorEnd()
			return m, textinput.Blink
		case "t", "T":
			m.showTheme = true
		case "f", "F":
			if m.favorites != nil {
				if station, ok := m.currentStation(); ok {
					_, err := m.favorites.Toggle(station)
					if err != nil {
						m.errMsg = err.Error()
						return m, nil
					}
					if m.stationSource == sourceFavorites {
						m.page = 0
						m.hasMore = false
						m.selected = 0
						m.loading = true
						return m, m.loadStationsCmd()
					}
				}
			}
		}
	case stationsMsg:
		if msg.source != m.stationSource || msg.page != m.page || msg.country != m.country || msg.search != m.activeSearch {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.stations = nil
			m.hasMore = false
			m.selected = 0
			return m, nil
		}
		m.errMsg = ""
		m.stations = msg.stations
		m.hasMore = msg.hasMore
		m.selected = 0
		m.ensureSelection()
		m.updateDialRange()
		m.snapDial()
		return m, nil
	case ipcReadyMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.ipc = msg.server
		return m, m.listenIPCCmd()
	case ipcMsg:
		return m.handleIPC(msg)
	case ipcClosedMsg:
		m.ipc = nil
		return m, nil
	case playerDownloadMsg:
		m.downloadingPlayer = false
		if msg.err != nil {
			m.errMsg = "Failed to download ffplay: " + msg.err.Error() + " (install mpv or ffplay and ensure it is in PATH)"
			return m, nil
		}
		p, err := player.New()
		if err != nil {
			m.errMsg = "Audio player not available: " + err.Error()
			return m, nil
		}
		m.player = p
		m.missingPlayer = false
		m.downloadingPlayer = false
		m.errMsg = ""
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
			if m.downloadingPlayer {
				m.errMsg = "Audio player not available yet. Downloading ffplay..."
			} else {
				m.errMsg = "Audio player not available. Install mpv or ffplay and ensure it is in PATH."
			}
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
	case themeSavedMsg:
		if msg.err != nil {
			m.errMsg = "Failed to save theme: " + msg.err.Error()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) saveThemeCmd() tea.Cmd {
	slug := m.theme.Slug
	return func() tea.Msg {
		err := config.SaveTheme(slug)
		return themeSavedMsg{err: err}
	}
}

func (m Model) loadStationsCmd() tea.Cmd {
	source := m.stationSource
	country := m.country
	search := strings.TrimSpace(m.activeSearch)
	page := m.page
	api := m.api
	favorites := m.favorites
	return func() tea.Msg {
		if source == sourceFavorites {
			all := []radio.Station{}
			if favorites != nil {
				all = favoritesToStations(favorites.List())
			}

			if search != "" {
				filtered := make([]radio.Station, 0, len(all))
				searchLower := strings.ToLower(search)
				for _, station := range all {
					name := strings.ToLower(station.Name)
					tags := strings.ToLower(station.Tags)
					countryName := strings.ToLower(station.Country)
					if strings.Contains(name, searchLower) || strings.Contains(tags, searchLower) || strings.Contains(countryName, searchLower) {
						filtered = append(filtered, station)
					}
				}
				all = filtered
			}

			offset := page * stationPageSize
			if offset < 0 {
				offset = 0
			}
			if offset >= len(all) {
				return stationsMsg{
					stations: nil,
					source:   source,
					page:     page,
					country:  country,
					search:   search,
					hasMore:  false,
				}
			}

			end := offset + stationPageSize
			hasMore := false
			if end < len(all) {
				hasMore = true
			} else {
				end = len(all)
			}

			return stationsMsg{
				stations: all[offset:end],
				source:   source,
				page:     page,
				country:  country,
				search:   search,
				hasMore:  hasMore,
			}
		}

		if api == nil {
			return stationsMsg{err: fmt.Errorf("radio api not available"), source: source}
		}
		offset := page * stationPageSize
		limit := stationPageSize + 1

		var (
			stations []radio.Station
			err      error
		)
		if search == "" {
			stations, err = api.StationsByCountryPage(context.Background(), country, limit, offset)
		} else {
			stations, err = api.SearchStationsByCountry(context.Background(), country, search, limit, offset)
		}
		if err != nil {
			return stationsMsg{
				err:     err,
				source:  source,
				page:    page,
				country: country,
				search:  search,
			}
		}

		hasMore := len(stations) > stationPageSize
		if hasMore {
			stations = stations[:stationPageSize]
		}
		return stationsMsg{
			stations: stations,
			source:   source,
			page:     page,
			country:  country,
			search:   search,
			hasMore:  hasMore,
		}
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

func (m Model) maybeDownloadPlayerCmd() tea.Cmd {
	if !m.missingPlayer || !m.downloadingPlayer {
		return nil
	}
	if runtime.GOOS != "windows" {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		path, err := player.DownloadFFplay(ctx)
		return playerDownloadMsg{path: path, err: err}
	}
}

func (m Model) startIPCCmd() tea.Cmd {
	return func() tea.Msg {
		server, err := newIPCServer()
		return ipcReadyMsg{server: server, err: err}
	}
}

func (m Model) listenIPCCmd() tea.Cmd {
	if m.ipc == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case msg := <-m.ipc.messages:
			return msg
		case <-m.ipc.done:
			return ipcClosedMsg{}
		}
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
			m.stationSource = sourceCountry
			m.country = strings.ToUpper(strings.TrimSpace(m.location.Value()))
			if m.country == "" {
				m.country = "US"
			}
			m.page = 0
			m.hasMore = false
			m.activeSearch = ""
			m.search.SetValue("")
			m.loading = true
			m.errMsg = ""
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

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			m.inputMode = inputNone
			m.activeSearch = strings.TrimSpace(m.search.Value())
			m.search.Blur()
			m.page = 0
			m.hasMore = false
			m.selected = 0
			m.loading = true
			m.errMsg = ""
			return m, m.loadStationsCmd()
		case "esc":
			m.inputMode = inputNone
			m.search.SetValue(m.activeSearch)
			m.search.Blur()
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
				m.stationSource = sourceCountry
				m.country = strings.ToUpper(strings.TrimSpace(country.Code))
				m.page = 0
				m.hasMore = false
				m.activeSearch = ""
				m.search.SetValue("")
				m.loading = true
				m.errMsg = ""
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

func (m Model) handleIPC(msg ipcMsg) (tea.Model, tea.Cmd) {
	cmd, err := parseIPCCommand(msg.cmd)
	if err != nil {
		sendIPCReply(msg.reply, ipcReply{ok: false, err: err.Error()})
		return m, m.listenIPCCmd()
	}

	var reply ipcReply
	var cmdTea tea.Cmd

	switch cmd {
	case "PLAY_PAUSE":
		cmdTea, reply = m.ipcPlayPause()
	case "NEXT":
		cmdTea, reply = m.ipcSelectAndPlay(1)
	case "PREV":
		cmdTea, reply = m.ipcSelectAndPlay(-1)
	case "QUIT":
		reply = ipcReply{ok: true}
		sendIPCReply(msg.reply, reply)
		if m.player != nil {
			_ = m.player.Stop()
		}
		if m.ipc != nil {
			m.ipc.Close()
		}
		return m, tea.Quit
	case "STATUS":
		reply = ipcReply{ok: true, data: m.ipcStatus()}
	case "PING":
		reply = ipcReply{ok: true, data: "OK"}
	default:
		reply = ipcReply{ok: false, err: "unknown command"}
	}

	sendIPCReply(msg.reply, reply)
	return m, tea.Batch(cmdTea, m.listenIPCCmd())
}

func (m *Model) ipcPlayPause() (tea.Cmd, ipcReply) {
	if m.playing {
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.playing = false
		return nil, ipcReply{ok: true}
	}

	station, ok := m.currentStation()
	if !ok {
		return nil, ipcReply{ok: false, err: "no station selected"}
	}
	return m.playStationCmd(station), ipcReply{ok: true, data: "QUEUED"}
}

func (m *Model) ipcSelectAndPlay(delta int) (tea.Cmd, ipcReply) {
	list := m.visibleStations()
	if len(list) == 0 {
		return nil, ipcReply{ok: false, err: "no stations available"}
	}
	m.moveSelection(delta)
	station, ok := m.currentStation()
	if !ok {
		return nil, ipcReply{ok: false, err: "no station selected"}
	}
	cmds := []tea.Cmd{m.dialTickCmd(), m.playStationCmd(station)}
	return tea.Batch(cmds...), ipcReply{ok: true, data: "QUEUED"}
}

func (m *Model) ipcStatus() string {
	station, _ := m.currentStation()
	name := station.Name
	if name == "" {
		name = "-"
	}

	playing := "false"
	if m.playing {
		playing = "true"
	}

	return fmt.Sprintf("{\"playing\":%s,\"station\":%q,\"country\":%q}", playing, name, m.country)
}

func sendIPCReply(ch chan ipcReply, reply ipcReply) {
	if ch == nil {
		return
	}
	select {
	case ch <- reply:
	case <-time.After(200 * time.Millisecond):
	}
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
	return m.stations
}

func (m *Model) visibleCountries() []radio.Country {
	if strings.TrimSpace(m.countrySearch.Value()) != "" {
		return m.filteredCountries
	}
	return m.countries
}

func (m *Model) isFavoritesSource() bool {
	return m.stationSource == sourceFavorites
}

func favoritesToStations(favs []config.Favorite) []radio.Station {
	stations := make([]radio.Station, 0, len(favs))
	for _, fav := range favs {
		if strings.TrimSpace(fav.UUID) == "" {
			continue
		}
		stations = append(stations, radio.Station{
			UUID:    fav.UUID,
			Name:    fav.Name,
			Country: fav.Country,
			Tags:    fav.Tags,
		})
	}
	return stations
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
