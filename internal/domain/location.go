package domain

// Location represents a geographic location with coordinates and metadata
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"` // meters above sea level
	Name      string  `json:"name"`      // Display name (city, country)
	Timezone  string  `json:"timezone"`  // IANA timezone identifier
}

// IsValid checks if the location has valid coordinates
func (l Location) IsValid() bool {
	return l.Latitude >= -90 && l.Latitude <= 90 &&
		l.Longitude >= -180 && l.Longitude <= 180
}

// DefaultLocation returns a default location (London, UK)
func DefaultLocation() Location {
	return Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		Elevation: 11,
		Name:      "London, United Kingdom",
		Timezone:  "Europe/London",
	}
}
