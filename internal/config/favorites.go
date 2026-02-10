package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"radio-tui/internal/radio"
)

type Favorite struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Tags    string `json:"tags"`
}

type Favorites struct {
	mu    sync.Mutex
	path  string
	items map[string]Favorite
}

type favoritesFile struct {
	Stations []Favorite `json:"stations"`
}

func LoadFavorites() (*Favorites, error) {
	path, err := favoritesPath()
	if err != nil {
		return nil, err
	}

	favs := &Favorites{
		path:  path,
		items: map[string]Favorite{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return favs, nil
		}
		return nil, err
	}

	var stored favoritesFile
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	for _, fav := range stored.Stations {
		if fav.UUID != "" {
			favs.items[fav.UUID] = fav
		}
	}

	return favs, nil
}

func (f *Favorites) Toggle(station radio.Station) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if station.UUID == "" {
		return false, errors.New("station uuid is required")
	}

	if _, ok := f.items[station.UUID]; ok {
		delete(f.items, station.UUID)
		return false, f.saveLocked()
	}

	f.items[station.UUID] = Favorite{
		UUID:    station.UUID,
		Name:    station.Name,
		Country: station.Country,
		Tags:    station.Tags,
	}
	return true, f.saveLocked()
}

func (f *Favorites) IsFavorite(uuid string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.items[uuid]
	return ok
}

func (f *Favorites) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}

	list := make([]Favorite, 0, len(f.items))
	for _, fav := range f.items {
		list = append(list, fav)
	}

	data, err := json.MarshalIndent(favoritesFile{Stations: list}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0o644)
}

func favoritesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "valvefm", "favorites.json"), nil
}
