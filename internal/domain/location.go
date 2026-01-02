// Package domain defines the core business entities for the GoGoldenHour application.
//
// This package contains value objects that represent geographic locations, sun times,
// time ranges, and user settings. These entities are used throughout the application
// by services (solar calculator, geocoding, geolocation) and UI components.
//
// The domain layer is independent of external frameworks, databases, or UI concerns,
// following clean architecture principles. All entities are immutable value objects
// with validation methods to ensure data integrity.
//
// Key types:
//   - Location: Geographic coordinates with metadata (name, timezone, elevation)
//   - SunTimes: Complete sun event data including golden/blue hour periods
//   - TimeRange: A period with start/end times and duration formatting
//   - Settings: User-configurable preferences for calculations and display
package domain

// Location represents a geographic point on Earth with associated metadata.
//
// Locations are used as input to the solar calculator and are obtained from:
//   - IP-based geolocation (IPAPIService)
//   - Address search (NominatimService)
//   - Map clicks (reverse geocoding)
//   - User preferences (last saved location)
//
// The Timezone field uses IANA timezone identifiers (e.g., "America/New_York",
// "Europe/London") which are required for accurate solar calculations. The tzf
// library automatically determines the timezone from coordinates when a location
// is created via geocoding services.
//
// Elevation affects solar calculations slightly but is typically set to 0 since
// most geocoding services don't provide elevation data. For most photography use
// cases, the difference is negligible (a few seconds at most).
type Location struct {
	// Latitude is the north-south position in degrees (-90 to 90).
	// Positive values are north of the equator, negative values are south.
	Latitude float64 `json:"latitude"`

	// Longitude is the east-west position in degrees (-180 to 180).
	// Positive values are east of the Prime Meridian, negative values are west.
	Longitude float64 `json:"longitude"`

	// Elevation is the height above sea level in meters.
	// Used for precise solar calculations but typically set to 0 when unknown.
	Elevation float64 `json:"elevation"`

	// Name is a human-readable display name (e.g., "Paris, France").
	// This is shown in the UI and obtained from geocoding services.
	Name string `json:"name"`

	// Timezone is an IANA timezone identifier (e.g., "Europe/Paris").
	// Required for converting UTC sun times to local time for display.
	// Automatically determined from coordinates using the tzf library.
	Timezone string `json:"timezone"`
}

// IsValid checks if the location has valid geographic coordinates.
//
// This validates that:
//   - Latitude is within the valid range of -90 to 90 degrees
//   - Longitude is within the valid range of -180 to 180 degrees
//
// Note: This does not validate the Name, Timezone, or Elevation fields.
// A location can have valid coordinates but be missing other metadata.
//
// Returns true if coordinates are within valid ranges, false otherwise.
func (l Location) IsValid() bool {
	return l.Latitude >= -90 && l.Latitude <= 90 &&
		l.Longitude >= -180 && l.Longitude <= 180
}

// DefaultLocation returns London, UK as the fallback location.
//
// This is used when:
//   - The application starts without a saved location
//   - IP-based geolocation fails
//   - The user's saved location cannot be loaded
//
// London was chosen as it's in a timezone with well-defined golden/blue hours
// year-round (unlike extreme latitudes) and is a commonly recognized location.
func DefaultLocation() Location {
	return Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		Elevation: 11,
		Name:      "London, United Kingdom",
		Timezone:  "Europe/London",
	}
}
