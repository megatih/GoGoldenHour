// Package solar provides astronomical calculations for sun positions and
// golden/blue hour times using the SAMPA (Solar Position Algorithm) library.
//
// This package is the core calculation engine of GoGoldenHour. It computes:
//
//   - Sunrise and sunset times
//   - Solar noon (sun's highest point)
//   - Golden hour periods (morning and evening)
//   - Blue hour periods (morning and evening)
//   - Real-time sun position (elevation and azimuth)
//
// The calculations use the go-sampa library, which implements the NOAA Solar
// Position Algorithm. This algorithm is accurate to within one minute for
// dates between 1950 and 2050.
//
// # Golden Hour and Blue Hour Definitions
//
// Golden Hour occurs when the sun is low on the horizon, producing warm,
// soft, directional light ideal for photography:
//
//   - Morning Golden Hour: Starts at sunrise (0°), ends at configurable elevation (default 6°)
//   - Evening Golden Hour: Starts at configurable elevation (default 6°), ends at sunset (0°)
//
// Blue Hour occurs when the sun is below the horizon, creating diffused
// blue light from the atmosphere:
//
//   - Morning Blue Hour: Before sunrise, sun between -8° and -4° (configurable)
//   - Evening Blue Hour: After sunset, sun between -4° and -8° (configurable)
//
// # Architecture
//
// The Calculator is created with user settings (elevation angles) and can be
// updated when settings change. It converts domain types to the go-sampa format,
// calculates sun events, and returns results as domain.SunTimes.
//
// # Thread Safety
//
// The Calculator is NOT thread-safe. If used from multiple goroutines, external
// synchronization is required. In the current app architecture, all calculations
// are performed on the main thread after location/date/settings changes.
//
// # Dependencies
//
//   - github.com/hablullah/go-sampa: Solar position calculations
//   - domain.Location: Input coordinates and timezone
//   - domain.Settings: Configurable elevation angles
//   - domain.SunTimes: Output structure with all calculated times
package solar

import (
	"fmt"
	"time"

	"github.com/hablullah/go-sampa"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// Calculator
// =============================================================================

// Calculator handles all solar position and time calculations.
//
// The calculator maintains user settings that control the elevation angles
// defining golden hour and blue hour boundaries. These settings can be updated
// at runtime when the user adjusts them in the settings panel.
//
// Usage:
//
//	calc := solar.New(settings)
//	sunTimes, err := calc.Calculate(location, date)
//	if err != nil {
//	    // Handle calculation error (rare, usually invalid input)
//	}
//	// Use sunTimes.Sunrise, sunTimes.GoldenMorning, etc.
type Calculator struct {
	// settings holds the current elevation angles for golden/blue hour definitions.
	// These are copied from domain.Settings when the calculator is created or updated.
	settings domain.Settings
}

// New creates a new solar calculator with the given settings.
//
// The settings determine the elevation angles that define golden hour and
// blue hour boundaries. These can be updated later via UpdateSettings.
//
// Parameters:
//   - settings: User preferences including elevation angles
//
// Returns a configured Calculator ready for use.
func New(settings domain.Settings) *Calculator {
	return &Calculator{settings: settings}
}

// UpdateSettings replaces the calculator's settings with new values.
//
// This is called when the user changes elevation angles in the settings panel.
// After updating settings, the app should call Calculate again to get new
// times based on the updated elevation angles.
//
// Parameters:
//   - settings: New user preferences to apply
func (c *Calculator) UpdateSettings(settings domain.Settings) {
	c.settings = settings
}

// =============================================================================
// Helper Functions
// =============================================================================

// toSampaLocation converts a domain.Location to the sampa.Location format.
//
// The go-sampa library uses its own Location type that doesn't include
// metadata like timezone or name. This function extracts just the
// geographic coordinates needed for calculations.
//
// Parameters:
//   - loc: Domain location with full metadata
//
// Returns a sampa.Location with only lat/lon/elevation.
func toSampaLocation(loc domain.Location) sampa.Location {
	return sampa.Location{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Elevation: loc.Elevation,
	}
}

// extractTimeRange extracts a time range from the sampa sun positions map.
//
// The go-sampa library returns custom events in a map keyed by event name.
// This function looks up the start and end events and combines them into
// a domain.TimeRange. If either event is missing (which can happen at
// extreme latitudes), an empty TimeRange is returned.
//
// Parameters:
//   - events: Map of event names to sun positions from sampa
//   - startKey: Name of the event marking the start of the range
//   - endKey: Name of the event marking the end of the range
//
// Returns a TimeRange, or an empty TimeRange if either event is missing.
func extractTimeRange(events map[string]sampa.SunPosition, startKey, endKey string) domain.TimeRange {
	start, hasStart := events[startKey]
	end, hasEnd := events[endKey]

	// Both events must exist for a valid range
	if hasStart && hasEnd {
		return domain.TimeRange{
			Start: start.DateTime,
			End:   end.DateTime,
		}
	}

	// Return empty range if events don't exist (extreme latitudes)
	return domain.TimeRange{}
}

// =============================================================================
// Main Calculation Method
// =============================================================================

// Calculate computes all sun times for a given location and date.
//
// This is the main entry point for solar calculations. It takes a location
// and date, and returns all sun events including sunrise, sunset, and
// golden/blue hour periods.
//
// The calculation process:
//  1. Load the timezone for accurate local time conversion
//  2. Normalize the date to midnight in the location's timezone
//  3. Define 8 custom sun events for golden/blue hour boundaries
//  4. Call go-sampa to calculate when the sun reaches each elevation
//  5. Extract and combine results into domain.SunTimes
//
// Parameters:
//   - loc: Geographic location with timezone information
//   - date: The date for which to calculate (time portion is ignored)
//
// Returns:
//   - domain.SunTimes: Complete sun event data for the date
//   - error: Non-nil if calculation fails (rare)
//
// Errors can occur if the timezone is invalid and can't be loaded, or if
// the go-sampa library encounters an internal error. In practice, these
// errors are rare with validated input.
func (c *Calculator) Calculate(loc domain.Location, date time.Time) (domain.SunTimes, error) {
	// Load the timezone for the location to ensure all times are in local time.
	// This is important because users expect to see times in their local timezone.
	tz, err := time.LoadLocation(loc.Timezone)
	if err != nil {
		// Fall back to system local timezone if the stored timezone is invalid.
		// This shouldn't happen with properly validated locations, but provides
		// a reasonable fallback.
		tz = time.Local
	}

	// Normalize the date to midnight in the target timezone.
	// go-sampa calculates events for the entire day starting from this time.
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)

	// Convert domain location to sampa format
	sampaLoc := toSampaLocation(loc)

	// Create custom events for all golden/blue hour boundaries.
	// These are defined based on the user's configured elevation angles.
	customEvents := c.createCustomEvents()

	// Calculate all sun events using the go-sampa library.
	// This returns standard events (sunrise, sunset, transit) plus our custom events.
	events, err := sampa.GetSunEvents(date, sampaLoc, nil, customEvents...)
	if err != nil {
		return domain.SunTimes{}, fmt.Errorf("failed to calculate sun events: %w", err)
	}

	// Build the result by extracting times from the events.
	// Standard events are in the SunEvents struct; custom events are in Others map.
	sunTimes := domain.SunTimes{
		Date:      date,
		Location:  loc,
		Sunrise:   events.Sunrise.DateTime,
		Sunset:    events.Sunset.DateTime,
		SolarNoon: events.Transit.DateTime,
		// Extract golden/blue hour ranges from custom events
		GoldenMorning: extractTimeRange(events.Others, "GoldenMorningStart", "GoldenMorningEnd"),
		GoldenEvening: extractTimeRange(events.Others, "GoldenEveningStart", "GoldenEveningEnd"),
		BlueMorning:   extractTimeRange(events.Others, "BlueMorningStart", "BlueMorningEnd"),
		BlueEvening:   extractTimeRange(events.Others, "BlueEveningStart", "BlueEveningEnd"),
	}

	return sunTimes, nil
}

// =============================================================================
// Custom Event Definitions
// =============================================================================

// createCustomEvents creates the 8 custom sun events for golden and blue hour.
//
// The go-sampa library supports custom events defined by elevation angles.
// Each event specifies:
//   - Name: Unique identifier for the event
//   - BeforeTransit: true for morning events, false for evening events
//   - Elevation: Function returning the target sun elevation angle
//
// We define 8 events total (4 pairs for golden/blue morning/evening):
//
// Golden Hour Events:
//   - GoldenMorningStart: Sunrise (0°) - when sun appears on horizon
//   - GoldenMorningEnd: Golden elevation (e.g., 6°) - sun too high for golden hour
//   - GoldenEveningStart: Golden elevation - sun low enough for golden hour
//   - GoldenEveningEnd: Sunset (0°) - sun disappears below horizon
//
// Blue Hour Events:
//   - BlueMorningStart: Blue end (e.g., -8°) - earliest blue hour
//   - BlueMorningEnd: Blue start (e.g., -4°) - end of blue, start of pre-dawn
//   - BlueEveningStart: Blue start - sun just below horizon, blue light begins
//   - BlueEveningEnd: Blue end - deep twilight, blue hour ends
//
// Note: The Elevation functions capture the settings values at creation time.
// If settings change, createCustomEvents must be called again to get updated events.
func (c *Calculator) createCustomEvents() []sampa.CustomSunEvent {
	// Capture current settings values for use in elevation functions
	goldenElevation := c.settings.GoldenHourElevation
	blueStart := c.settings.BlueHourStart
	blueEnd := c.settings.BlueHourEnd

	return []sampa.CustomSunEvent{
		// =========================================================================
		// Morning Golden Hour: sunrise (0°) → golden elevation (default 6°)
		// =========================================================================
		// This period starts when the sun rises above the horizon and ends when
		// it climbs too high for the warm, directional light of golden hour.
		{
			Name:          "GoldenMorningStart",
			BeforeTransit: true, // Morning = before solar noon
			Elevation: func(_ sampa.SunPosition) float64 {
				return 0.0 // Sunrise: sun at horizon level
			},
		},
		{
			Name:          "GoldenMorningEnd",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return goldenElevation // End when sun exceeds golden elevation
			},
		},

		// =========================================================================
		// Evening Golden Hour: golden elevation (default 6°) → sunset (0°)
		// =========================================================================
		// This period starts when the sun drops low enough for warm light and
		// ends when it sets below the horizon.
		{
			Name:          "GoldenEveningStart",
			BeforeTransit: false, // Evening = after solar noon
			Elevation: func(_ sampa.SunPosition) float64 {
				return goldenElevation // Start when sun drops to golden elevation
			},
		},
		{
			Name:          "GoldenEveningEnd",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return 0.0 // Sunset: sun at horizon level
			},
		},

		// =========================================================================
		// Morning Blue Hour: blue end (default -8°) → blue start (default -4°)
		// =========================================================================
		// This period occurs before sunrise when the sun is below the horizon
		// but high enough for blue light to illuminate the sky.
		// Note: Start is at the lower angle (deeper twilight) because time progresses
		// from darker to lighter in the morning.
		{
			Name:          "BlueMorningStart",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueEnd // e.g., -8° (deeper twilight = earlier time)
			},
		},
		{
			Name:          "BlueMorningEnd",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueStart // e.g., -4° (shallower twilight = later time)
			},
		},

		// =========================================================================
		// Evening Blue Hour: blue start (default -4°) → blue end (default -8°)
		// =========================================================================
		// This period occurs after sunset when the sun is below the horizon
		// creating the characteristic blue twilight.
		// Note: Start is at the higher angle (shallower twilight) because time
		// progresses from lighter to darker in the evening.
		{
			Name:          "BlueEveningStart",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueStart // e.g., -4° (shallower twilight = earlier time)
			},
		},
		{
			Name:          "BlueEveningEnd",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueEnd // e.g., -8° (deeper twilight = later time)
			},
		},
	}
}

// =============================================================================
// Real-Time Sun Position
// =============================================================================

// GetCurrentSunPosition returns the current position of the sun at a location.
//
// This provides real-time sun position data that could be used to display
// current sun elevation/azimuth in the UI or determine if it's currently
// golden/blue hour.
//
// Parameters:
//   - loc: Geographic location to calculate position for
//
// Returns:
//   - elevation: Sun's angle above/below horizon in degrees
//     (positive = above horizon, negative = below)
//   - azimuth: Sun's compass direction in degrees (0° = North, 90° = East)
//   - error: Non-nil if calculation fails
//
// Example:
//
//	elevation, azimuth, err := calc.GetCurrentSunPosition(location)
//	if elevation >= 0 && elevation <= 6 {
//	    fmt.Println("Currently golden hour!")
//	}
func (c *Calculator) GetCurrentSunPosition(loc domain.Location) (float64, float64, error) {
	// Use go-sampa to calculate the sun's current position
	pos, err := sampa.GetSunPosition(time.Now(), toSampaLocation(loc), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get sun position: %w", err)
	}

	// Return the topocentric angles (adjusted for observer's position on Earth's surface)
	// TopocentricElevationAngle: how high the sun is above the horizon
	// TopocentricAzimuthAngle: compass direction to the sun
	return pos.TopocentricElevationAngle, pos.TopocentricAzimuthAngle, nil
}
