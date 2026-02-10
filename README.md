# Valve FM

Vintage FM radio TUI for streaming stations from radio-browser.info.

## Requirements

- Go 1.22+
- `mpv` (preferred) or `ffplay` in `PATH`

## Run (TUI + Tray)

```bash
go run ./cmd/radio-tray
```

Notes:

- The app always runs with the tray enabled.
- macOS/Linux socket path: `~/.config/valvefm/ctl.sock`
- Windows address file: `~/.config/valvefm/ctl.addr`
- Windows auto-downloads `ffplay.exe` on first run if no player is found.

### Windows build note

If you want the TUI visible, build without the GUI subsystem:

```bash
GOOS=windows GOARCH=amd64 go build -o valvefm.exe ./cmd/radio-tray
```

## Keybindings

- Left / Right: tune dial
- Up / Down: browse stations
- Enter: play station
- Space: stop / resume
- L: choose country (searchable list)
- /: search stations
- F: toggle favorite
- ?: help
- Q / Ctrl+C: quit

## Notes

- Stations are fetched from the Radio Browser API and sorted by popularity.
- Country selection uses a searchable list from the API.
- Favorites are saved to `~/.config/valvefm/favorites.json`.

## Smoke Test Checklist

- Launch: `go run ./cmd/radio-tray` starts TUI and tray.
- Country selector: `L` opens list, filter works, Enter loads stations.
- Playback: Enter starts audio, Space stops/resumes.
- Next/Prev: tray controls move station and auto-play.
- Search: `/` filters stations and restores when cleared.
- Quit: tray Quit and `Q` cleanly stop playback.

## Licenses

Valve FM may download `ffplay.exe` on Windows. Include `THIRD_PARTY_NOTICES.md` in your distribution and follow FFmpeg's license terms.
