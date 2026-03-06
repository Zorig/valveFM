# Valve FM

Vintage FM radio TUI for streaming stations from radio-browser.info.
<img width="3024" height="1898" alt="image" src="https://github.com/user-attachments/assets/c9c6ab00-7525-4402-afcd-32ba0e6cd83c" />

## Install

### Homebrew (macOS / Linux)

```bash
brew tap zorig/tap
brew install valvefm
```

### Chocolatey (Windows) soon!

```powershell
choco install valvefm
```

### From source

## Requirements

- **Go 1.24+**
- **Audio Backend:** Built-in pure Go MP3 player (no external deps).
- **Optional:** `mpv` or `ffplay` for AAC/OGG support and better streaming stability.
  - Windows: automatically downloads `ffplay.exe` if needed.

## Run (TUI + Tray)

```bash
go run ./cmd/radio-tray
```

Notes:

- The app always runs with the tray enabled.
- macOS/Linux socket path: `~/.config/valvefm/ctl.sock`
- Windows address file: `~/.config/valvefm/ctl.addr`
- Windows auto-downloads `ffplay.exe` on first run if no player is found.

### ⚠️ Windows SmartScreen Warning
When running `valvefm-windows-amd64.exe` for the first time, Windows might show a "Windows protected your PC" warning because the app is unsigned.
1. Click **More info**.
2. Click **Run anyway**.
(This is normal for open-source software without an expensive code signing certificate.)

### Windows build note

If you want the TUI visible, build without the GUI subsystem:

```bash
GOOS=windows GOARCH=amd64 go build -o valvefm.exe ./cmd/radio-tray
```

## Keybindings

- Left / Right: tune dial
- Up / Down: browse stations
- [ / ]: previous / next stations page
- Enter: play station
- Space: stop / resume
- L: choose country (searchable list)
- V: show favorites
- /: search stations (server-side in country mode, local in favorites mode)
- F: toggle favorite
- T: change theme
- ?: help
- Q / Ctrl+C: quit

## Notes

- Stations are fetched from the Radio Browser API and sorted by popularity.
- Station list and search results are paginated (200 stations per page).
- If favorites exist, app opens with favorites list by default.
- Country selection uses a searchable list from the API.
- Favorites are saved to `~/.config/valvefm/favorites.json`.
- Theme preference is saved to `~/.config/valvefm/config.json`.
- 12 built-in themes: Vintage, Tokyo Night, Nord, Catppuccin Mocha/Latte, Gruvbox Dark, Dracula, Solarized Dark, One Dark, Rose Pine, Kanagawa, Everforest.

## Smoke Test Checklist

- Launch: `go run ./cmd/radio-tray` starts TUI and tray.
- Country selector: `L` opens list, filter works, Enter loads stations.
- Favorites view: `V` opens saved favorites.
- Playback: Enter starts audio, Space stops/resumes.
- Next/Prev: tray controls move station and auto-play.
- Search: `/` runs server-side search in country mode and local search in favorites mode.
- Pagination: `[` and `]` move between station pages.
- Quit: tray Quit and `Q` cleanly stop playback.

## Licenses

Valve FM may download `ffplay.exe` on Windows. Include `THIRD_PARTY_NOTICES.md` in your distribution and follow FFmpeg's license terms.
