package domain

// =============================================================================
// Settings
// =============================================================================

// Settings holds user-configurable preferences for calculations and display.
//
// These settings control two aspects of the application:
//
// 1. Solar Calculation Parameters:
//   - GoldenHourElevation: defines when golden hour ends (sun elevation angle)
//   - BlueHourStart/BlueHourEnd: define the blue hour period boundaries
//
// 2. Display Preferences:
//   - TimeFormat24Hour: controls time display format
//   - AutoDetectLocation: enables IP-based location detection on startup
//   - LastLocation: persists the user's last selected location
//
// Settings are persisted to disk via PreferencesStore and loaded on application
// startup. The Validate method ensures all values are within acceptable ranges
// to prevent calculation errors.
//
// Understanding Sun Elevation Angles:
//
//	+90° = directly overhead (zenith)
//	+6°  = default golden hour upper boundary
//	 0°  = horizon (sunrise/sunset)
//	-4°  = default blue hour start (civil twilight)
//	-6°  = civil twilight end
//	-8°  = default blue hour end
//	-12° = nautical twilight end
//	-18° = astronomical twilight end (full darkness)
//
// Photographers have different preferences for these boundaries depending on
// their style and the lighting conditions they prefer. The defaults represent
// commonly accepted definitions in the photography community.
type Settings struct {
	// GoldenHourElevation is the sun elevation angle that marks the upper boundary
	// of golden hour. When the sun is between 0° (horizon) and this angle, the
	// lighting conditions are considered "golden hour."
	//
	// Range: 0 to 15 degrees (validated by Validate method)
	// Default: 6 degrees
	//
	// Lower values result in shorter golden hours with warmer, more dramatic light.
	// Higher values extend golden hour but include less warm light.
	GoldenHourElevation float64 `json:"golden_hour_elevation"`

	// BlueHourStart is the sun elevation angle where blue hour begins.
	// This is the angle closest to sunset/sunrise, when the sky starts showing
	// blue tones but still has some warm colors near the horizon.
	//
	// Range: -6 to 0 degrees (validated by Validate method)
	// Default: -4 degrees
	//
	// This corresponds roughly to civil twilight. Lower (more negative) values
	// start blue hour earlier in the morning or later in the evening.
	BlueHourStart float64 `json:"blue_hour_start"`

	// BlueHourEnd is the sun elevation angle where blue hour ends.
	// This is the angle furthest from sunset/sunrise, when the blue light
	// fades to darkness (evening) or transitions to pre-dawn (morning).
	//
	// Range: -18 to -6 degrees (validated by Validate method)
	// Default: -8 degrees
	//
	// This is slightly past civil twilight. Lower (more negative) values extend
	// blue hour into deeper twilight with darker, more saturated blues.
	BlueHourEnd float64 `json:"blue_hour_end"`

	// TimeFormat24Hour determines whether times are displayed in 24-hour format.
	// true  = 24-hour format (e.g., "14:30", "06:45")
	// false = 12-hour format with AM/PM (e.g., "2:30 PM", "6:45 AM")
	//
	// Default: true (24-hour format)
	TimeFormat24Hour bool `json:"time_format_24_hour"`

	// AutoDetectLocation enables automatic IP-based location detection on startup.
	// When enabled, the app queries ip-api.com to determine the user's approximate
	// location based on their IP address. This is convenient but may not be accurate
	// for users behind VPNs or in regions with poor IP geolocation data.
	//
	// When disabled, the app uses the LastLocation if available, or falls back
	// to the default location (London, UK).
	//
	// Default: true (auto-detect enabled)
	AutoDetectLocation bool `json:"auto_detect_location"`

	// LastLocation stores the user's last selected location for persistence.
	// This is used to restore the user's location when they restart the app
	// (if AutoDetectLocation is disabled) and is updated whenever the user
	// changes their location.
	//
	// This field is a pointer so it can be nil (omitted from JSON) when no
	// location has been saved yet.
	LastLocation *Location `json:"last_location,omitempty"`
}

// DefaultSettings returns the default application settings.
//
// These defaults represent commonly accepted definitions in the photography
// community for golden hour and blue hour. Users can adjust these values
// via the settings panel to match their preferences.
//
// Default values:
//   - Golden hour elevation: 6° (sun 0-6° above horizon)
//   - Blue hour: -4° to -8° (sun 4-8° below horizon)
//   - Time format: 24-hour
//   - Auto-detect location: enabled
//   - Last location: none (will use London, UK as fallback)
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

// Validate ensures all settings are within acceptable ranges.
//
// This method is called after loading settings from disk to ensure that
// corrupted or hand-edited configuration files don't cause calculation errors.
// Invalid values are clamped to their nearest valid boundary.
//
// Validation rules:
//   - GoldenHourElevation: clamped to [0, 15] degrees
//   - BlueHourStart: clamped to [-6, 0] degrees
//   - BlueHourEnd: clamped to [-18, -6] degrees
//   - BlueHourEnd must be below BlueHourStart (if not, adjusted to Start - 4)
//
// The method modifies the Settings in place (receiver is a pointer).
func (s *Settings) Validate() {
	// Golden hour elevation must be between 0° (horizon) and 15° (very high sun)
	// Values outside this range would produce unrealistic golden hour periods
	if s.GoldenHourElevation < 0 {
		s.GoldenHourElevation = 0
	} else if s.GoldenHourElevation > 15 {
		s.GoldenHourElevation = 15
	}

	// Blue hour start must be between 0° and -6° (civil twilight boundary)
	// Values outside this range would overlap with golden hour or deep twilight
	if s.BlueHourStart > 0 {
		s.BlueHourStart = 0
	} else if s.BlueHourStart < -6 {
		s.BlueHourStart = -6
	}

	// Blue hour end must be between -6° and -18° (astronomical twilight)
	// Below -18° is full darkness, above -6° is civil twilight
	if s.BlueHourEnd > -6 {
		s.BlueHourEnd = -6
	} else if s.BlueHourEnd < -18 {
		s.BlueHourEnd = -18
	}

	// Blue hour end must be below (more negative than) blue hour start
	// Otherwise the blue hour period would have negative duration
	if s.BlueHourEnd > s.BlueHourStart {
		s.BlueHourEnd = s.BlueHourStart - 4
	}
}
