package domain

// Settings holds user-configurable preferences
type Settings struct {
	// GoldenHourElevation is the sun elevation angle that defines the end of golden hour
	// Default: 6 degrees above horizon
	GoldenHourElevation float64 `json:"golden_hour_elevation"`

	// BlueHourStart is the sun elevation angle where blue hour begins
	// Default: -4 degrees (below horizon)
	BlueHourStart float64 `json:"blue_hour_start"`

	// BlueHourEnd is the sun elevation angle where blue hour ends
	// Default: -8 degrees (below horizon)
	BlueHourEnd float64 `json:"blue_hour_end"`

	// TimeFormat24Hour determines the time display format
	// true = 24-hour format (14:30), false = 12-hour format (2:30 PM)
	TimeFormat24Hour bool `json:"time_format_24_hour"`

	// AutoDetectLocation enables automatic IP-based location detection on startup
	AutoDetectLocation bool `json:"auto_detect_location"`

	// LastLocation stores the last used location for persistence
	LastLocation *Location `json:"last_location,omitempty"`
}

// DefaultSettings returns the default application settings
func DefaultSettings() Settings {
	return Settings{
		GoldenHourElevation: 6.0,
		BlueHourStart:       -4.0,
		BlueHourEnd:         -8.0,
		TimeFormat24Hour:    true,
		AutoDetectLocation:  true,
		LastLocation:        nil,
	}
}

// Validate ensures settings are within acceptable ranges
func (s *Settings) Validate() {
	// Golden hour elevation should be between 0 and 15 degrees
	if s.GoldenHourElevation < 0 {
		s.GoldenHourElevation = 0
	} else if s.GoldenHourElevation > 15 {
		s.GoldenHourElevation = 15
	}

	// Blue hour start should be between 0 and -6 degrees
	if s.BlueHourStart > 0 {
		s.BlueHourStart = 0
	} else if s.BlueHourStart < -6 {
		s.BlueHourStart = -6
	}

	// Blue hour end should be between -6 and -18 degrees
	if s.BlueHourEnd > -6 {
		s.BlueHourEnd = -6
	} else if s.BlueHourEnd < -18 {
		s.BlueHourEnd = -18
	}

	// Ensure blue hour end is below blue hour start
	if s.BlueHourEnd > s.BlueHourStart {
		s.BlueHourEnd = s.BlueHourStart - 4
	}
}
