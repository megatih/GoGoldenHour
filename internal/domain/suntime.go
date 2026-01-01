package domain

import (
	"fmt"
	"time"
)

// TimeRange represents a period with start and end times
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration returns the length of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// IsValid checks if the time range has valid times (end after start)
func (tr TimeRange) IsValid() bool {
	return !tr.Start.IsZero() && !tr.End.IsZero() && tr.End.After(tr.Start)
}

// FormatDuration returns the duration as a human-readable string
func (tr TimeRange) FormatDuration() string {
	d := tr.Duration()
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%d min", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// SunTimes contains all calculated sun-related times for a specific date and location
type SunTimes struct {
	Date          time.Time `json:"date"`
	Location      Location  `json:"location"`
	Sunrise       time.Time `json:"sunrise"`
	Sunset        time.Time `json:"sunset"`
	SolarNoon     time.Time `json:"solar_noon"`
	GoldenMorning TimeRange `json:"golden_morning"`
	GoldenEvening TimeRange `json:"golden_evening"`
	BlueMorning   TimeRange `json:"blue_morning"`
	BlueEvening   TimeRange `json:"blue_evening"`
}

// HasValidGoldenHour returns true if golden hour times are available
// (may not be available at extreme latitudes during certain seasons)
func (st SunTimes) HasValidGoldenHour() bool {
	return st.GoldenMorning.IsValid() || st.GoldenEvening.IsValid()
}

// HasValidBlueHour returns true if blue hour times are available
func (st SunTimes) HasValidBlueHour() bool {
	return st.BlueMorning.IsValid() || st.BlueEvening.IsValid()
}

// FormatTime formats a time according to the given format preference
func FormatTime(t time.Time, use24Hour bool) string {
	if t.IsZero() {
		return "--:--"
	}
	if use24Hour {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}
