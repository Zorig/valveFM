package radio

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Station represents a station from Radio Browser API.
type Station struct {
	UUID        string    `json:"stationuuid"`
	Name        string    `json:"name"`
	Country     string    `json:"country"`
	CountryCode string    `json:"countrycode"`
	Tags        string    `json:"tags"`
	Bitrate     int       `json:"bitrate"`
	Frequency   Frequency `json:"frequency"`
	URLResolved string    `json:"url_resolved"`
	URL         string    `json:"url"`
	Favicon     string    `json:"favicon"`
	ClickCount  int       `json:"clickcount"`
	IsBroken    bool      `json:"is_broken"`
}

// Frequency captures station frequency when provided by the API.
type Frequency float64

func (f Frequency) Float64() float64 {
	return float64(f)
}

func (f *Frequency) UnmarshalJSON(data []byte) error {
	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = Frequency(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return nil
	}

	text = strings.TrimSpace(text)
	if text == "" {
		*f = 0
		return nil
	}

	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		*f = 0
		return nil
	}
	*f = Frequency(number)
	return nil
}
