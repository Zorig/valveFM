package radio

// Country represents a country entry from Radio Browser API.
type Country struct {
	Name         string `json:"name"`
	Code         string `json:"iso_3166_1"`
	StationCount int    `json:"stationcount"`
}
