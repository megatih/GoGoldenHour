package solar

import (
	"fmt"
	"time"

	"github.com/hablullah/go-sampa"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// Calculator handles all solar position and time calculations
type Calculator struct {
	settings domain.Settings
}

// New creates a new solar calculator with the given settings
func New(settings domain.Settings) *Calculator {
	return &Calculator{settings: settings}
}

// UpdateSettings updates the calculator's settings
func (c *Calculator) UpdateSettings(settings domain.Settings) {
	c.settings = settings
}

// toSampaLocation converts a domain.Location to sampa.Location
func toSampaLocation(loc domain.Location) sampa.Location {
	return sampa.Location{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Elevation: loc.Elevation,
	}
}

// extractTimeRange extracts a time range from sampa sun positions
func extractTimeRange(events map[string]sampa.SunPosition, startKey, endKey string) domain.TimeRange {
	start, hasStart := events[startKey]
	end, hasEnd := events[endKey]
	if hasStart && hasEnd {
		return domain.TimeRange{
			Start: start.DateTime,
			End:   end.DateTime,
		}
	}
	return domain.TimeRange{}
}

// Calculate computes all sun times for a given location and date
func (c *Calculator) Calculate(loc domain.Location, date time.Time) (domain.SunTimes, error) {
	// Load timezone
	tz, err := time.LoadLocation(loc.Timezone)
	if err != nil {
		// Fall back to local timezone
		tz = time.Local
	}

	// Ensure date is in the correct timezone
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)

	sampaLoc := toSampaLocation(loc)

	// Define custom events for Golden and Blue Hour boundaries
	customEvents := c.createCustomEvents()

	// Calculate sun events
	events, err := sampa.GetSunEvents(date, sampaLoc, nil, customEvents...)
	if err != nil {
		return domain.SunTimes{}, fmt.Errorf("failed to calculate sun events: %w", err)
	}

	// Build the result
	sunTimes := domain.SunTimes{
		Date:          date,
		Location:      loc,
		Sunrise:       events.Sunrise.DateTime,
		Sunset:        events.Sunset.DateTime,
		SolarNoon:     events.Transit.DateTime,
		GoldenMorning: extractTimeRange(events.Others, "GoldenMorningStart", "GoldenMorningEnd"),
		GoldenEvening: extractTimeRange(events.Others, "GoldenEveningStart", "GoldenEveningEnd"),
		BlueMorning:   extractTimeRange(events.Others, "BlueMorningStart", "BlueMorningEnd"),
		BlueEvening:   extractTimeRange(events.Others, "BlueEveningStart", "BlueEveningEnd"),
	}

	return sunTimes, nil
}

// createCustomEvents creates the custom sun events for Golden and Blue Hour
func (c *Calculator) createCustomEvents() []sampa.CustomSunEvent {
	goldenElevation := c.settings.GoldenHourElevation
	blueStart := c.settings.BlueHourStart
	blueEnd := c.settings.BlueHourEnd

	return []sampa.CustomSunEvent{
		// Morning Golden Hour: from sunrise (0°) to golden elevation
		{
			Name:          "GoldenMorningStart",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return 0.0 // Sunrise
			},
		},
		{
			Name:          "GoldenMorningEnd",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return goldenElevation
			},
		},

		// Evening Golden Hour: from golden elevation to sunset (0°)
		{
			Name:          "GoldenEveningStart",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return goldenElevation
			},
		},
		{
			Name:          "GoldenEveningEnd",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return 0.0 // Sunset
			},
		},

		// Morning Blue Hour: from blue end to blue start (before sunrise)
		{
			Name:          "BlueMorningStart",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueEnd // e.g., -8°
			},
		},
		{
			Name:          "BlueMorningEnd",
			BeforeTransit: true,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueStart // e.g., -4°
			},
		},

		// Evening Blue Hour: from blue start to blue end (after sunset)
		{
			Name:          "BlueEveningStart",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueStart // e.g., -4°
			},
		},
		{
			Name:          "BlueEveningEnd",
			BeforeTransit: false,
			Elevation: func(_ sampa.SunPosition) float64 {
				return blueEnd // e.g., -8°
			},
		},
	}
}

// GetCurrentSunPosition returns the current position of the sun
func (c *Calculator) GetCurrentSunPosition(loc domain.Location) (float64, float64, error) {
	pos, err := sampa.GetSunPosition(time.Now(), toSampaLocation(loc), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get sun position: %w", err)
	}

	return pos.TopocentricElevationAngle, pos.TopocentricAzimuthAngle, nil
}
