package domain

import (
	"fmt"
	"time"
)

// =============================================================================
// TimeRange
// =============================================================================

// TimeRange represents a time period with defined start and end moments.
//
// This type is used throughout the application to represent golden hour and
// blue hour periods, which are defined by sun elevation angles. Each period
// has a start time (when the sun reaches one elevation) and an end time
// (when it reaches another elevation).
//
// TimeRange provides methods for:
//   - Checking validity (both times set, end after start)
//   - Calculating duration
//   - Formatting duration for human display
//
// A TimeRange may be invalid (have zero values) at extreme latitudes where
// the sun doesn't reach certain elevation angles. For example, during polar
// summer, the sun may never drop below -4° for blue hour to occur.
type TimeRange struct {
	// Start is the beginning of the time period.
	// For golden morning, this is sunrise (sun at 0° elevation).
	// For blue morning, this is when the sun rises above the blue hour end angle.
	Start time.Time `json:"start"`

	// End is the conclusion of the time period.
	// For golden morning, this is when the sun exceeds the golden hour elevation.
	// For blue morning, this is when the sun rises above the blue hour start angle.
	End time.Time `json:"end"`
}

// Duration returns the length of the time range as a time.Duration.
//
// This calculates the difference between End and Start times. If the TimeRange
// is invalid (zero times or End before Start), the result may be zero or negative.
// Callers should check IsValid() before relying on this value.
//
// Typical golden hour durations are 20-40 minutes depending on latitude and
// season. Blue hour durations are similar.
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// IsValid checks if the time range represents a real, usable period.
//
// A TimeRange is valid when:
//   - Start time is not zero (has been set)
//   - End time is not zero (has been set)
//   - End time is after Start time
//
// Invalid TimeRanges occur at extreme latitudes during certain seasons when
// the sun doesn't reach the required elevation angles. For example:
//   - Polar regions during summer: sun never sets, no blue hour
//   - High latitudes near solstices: golden hour may not exist
//
// The UI displays "N/A" for invalid time ranges.
func (tr TimeRange) IsValid() bool {
	return !tr.Start.IsZero() && !tr.End.IsZero() && tr.End.After(tr.Start)
}

// FormatDuration returns the duration as a human-readable string.
//
// Format rules:
//   - Under 60 minutes: "X min" (e.g., "45 min")
//   - Exactly N hours: "Nh" (e.g., "1h", "2h")
//   - Hours and minutes: "Nh Mm" (e.g., "1h 30m")
//
// This is used in the UI to show photographers how long each golden/blue
// hour period lasts, helping them plan their shoots.
func (tr TimeRange) FormatDuration() string {
	d := tr.Duration()
	minutes := int(d.Minutes())

	// Short durations: show only minutes
	if minutes < 60 {
		return fmt.Sprintf("%d min", minutes)
	}

	hours := minutes / 60
	mins := minutes % 60

	// Exact hours: omit minutes
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	// Hours and minutes combined
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// =============================================================================
// SunTimes
// =============================================================================

// SunTimes contains all calculated sun-related times for a specific date and location.
//
// This is the primary output of the solar calculator and contains everything
// needed to display sun information to the user:
//   - Basic sun events: sunrise, sunset, solar noon
//   - Golden hour periods: morning (after sunrise) and evening (before sunset)
//   - Blue hour periods: morning (before sunrise) and evening (after sunset)
//
// Golden Hour occurs when the sun is low on the horizon (typically 0-6° elevation),
// producing warm, soft light ideal for photography. Blue Hour occurs when the sun
// is below the horizon (typically -4° to -8°), creating cool, diffused light.
//
// The elevation angles that define these periods are user-configurable via Settings.
// Different photographers may prefer different definitions based on their style.
//
// Example calculation flow:
//
//	location := domain.Location{Latitude: 48.8566, Longitude: 2.3522, Timezone: "Europe/Paris"}
//	date := time.Now()
//	sunTimes := solarCalc.Calculate(location, date)
//	// sunTimes now contains all sun events for Paris on the current date
type SunTimes struct {
	// Date is the calendar date for which these times were calculated.
	// Times are in the location's local timezone.
	Date time.Time `json:"date"`

	// Location is the geographic position used for calculations.
	// Stored here for reference and potential recalculation.
	Location Location `json:"location"`

	// Sunrise is when the sun's upper edge appears above the horizon.
	// This marks the end of blue morning and start of golden morning.
	Sunrise time.Time `json:"sunrise"`

	// Sunset is when the sun's upper edge disappears below the horizon.
	// This marks the end of golden evening and start of blue evening.
	Sunset time.Time `json:"sunset"`

	// SolarNoon is when the sun reaches its highest point in the sky.
	// This is the midpoint between sunrise and sunset.
	SolarNoon time.Time `json:"solar_noon"`

	// GoldenMorning is the golden hour period after sunrise.
	// Starts at sunrise (0°) and ends when sun reaches golden elevation (default 6°).
	// Produces warm, directional light from the east.
	GoldenMorning TimeRange `json:"golden_morning"`

	// GoldenEvening is the golden hour period before sunset.
	// Starts when sun drops to golden elevation and ends at sunset (0°).
	// Produces warm, directional light from the west.
	GoldenEvening TimeRange `json:"golden_evening"`

	// BlueMorning is the blue hour period before sunrise.
	// Starts when sun rises above blue end angle (default -8°) and ends at
	// blue start angle (default -4°). Sky has blue/purple tones.
	BlueMorning TimeRange `json:"blue_morning"`

	// BlueEvening is the blue hour period after sunset.
	// Starts at blue start angle (default -4°) and ends at blue end angle
	// (default -8°). Sky transitions from orange to deep blue.
	BlueEvening TimeRange `json:"blue_evening"`
}

// HasValidGoldenHour returns true if at least one golden hour period is available.
//
// At extreme latitudes during certain seasons (e.g., polar summer), the sun may
// never reach the required elevation angles for golden hour. This method allows
// the UI to check availability before displaying golden hour information.
//
// Returns true if either GoldenMorning or GoldenEvening is valid.
func (st SunTimes) HasValidGoldenHour() bool {
	return st.GoldenMorning.IsValid() || st.GoldenEvening.IsValid()
}

// HasValidBlueHour returns true if at least one blue hour period is available.
//
// Similar to HasValidGoldenHour, this checks for blue hour availability at
// extreme latitudes. Blue hour requires the sun to be between specific negative
// elevations, which may not occur during polar day/night conditions.
//
// Returns true if either BlueMorning or BlueEvening is valid.
func (st SunTimes) HasValidBlueHour() bool {
	return st.BlueMorning.IsValid() || st.BlueEvening.IsValid()
}

// =============================================================================
// Time Formatting
// =============================================================================

// FormatTime formats a time according to the user's format preference.
//
// This is a utility function used throughout the UI for consistent time display.
// It handles zero times gracefully by returning a placeholder string.
//
// Format options:
//   - 24-hour: "15:04" (e.g., "14:30", "06:45")
//   - 12-hour: "3:04 PM" (e.g., "2:30 PM", "6:45 AM")
//
// Parameters:
//   - t: The time to format. If zero (unset), returns "--:--"
//   - use24Hour: If true, uses 24-hour format; otherwise uses 12-hour with AM/PM
//
// Returns the formatted time string, or "--:--" for zero times.
func FormatTime(t time.Time, use24Hour bool) string {
	// Handle zero/unset times with placeholder
	if t.IsZero() {
		return "--:--"
	}

	// Format according to user preference
	if use24Hour {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}
