# Valve FM

Vintage FM radio TUI for streaming stations from radio-browser.info.

## Requirements

- Go 1.22+
- `mpv` (preferred) or `ffplay` in `PATH`

## Run

```bash
go run ./cmd/radio
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
