// Package timezone provides geographic coordinate to timezone mapping.
//
// This package enables the application to automatically determine the correct
// timezone for any geographic location. This is essential for accurate solar
// calculations, as sunrise/sunset times must be expressed in the local timezone
// of the location being queried.
//
// # How It Works
//
// The package uses the tzf library, which contains an embedded database of
// timezone boundaries. Given a latitude and longitude, it performs a geometric
// lookup to determine which timezone polygon contains the point. This approach:
//
//   - Works entirely offline (no network requests needed)
//   - Is fast (geometric point-in-polygon lookup)
//   - Is accurate (uses the same IANA timezone database as most operating systems)
//
// # Usage in GoGoldenHour
//
// The timezone package is used in several places:
//
//  1. When geocoding search results (NominatimService.Search)
//  2. When the user clicks on the map (creating a new location)
//  3. When performing solar calculations (loading time.Location)
//
// # IANA Timezone Identifiers
//
// The package returns standard IANA timezone identifiers like:
//   - "America/New_York"
//   - "Europe/Paris"
//   - "Asia/Tokyo"
//   - "UTC" (fallback)
//
// These identifiers are compatible with Go's time.LoadLocation function
// and are used throughout the application for consistent time handling.
//
// # Fallback Behavior
//
// If a timezone cannot be determined (e.g., coordinates in the middle of the
// ocean), the package falls back to "UTC". This ensures the application always
// has a valid timezone, even if it may not be ideal for the specific location.
package timezone

import (
	"time"

	"github.com/ringsaturn/tzf"
)

// =============================================================================
// Package Initialization
// =============================================================================

// finder is the timezone lookup service from the tzf library.
// It's initialized once at package load time and reused for all lookups.
// The tzf library embeds the timezone boundary data, so no external files
// or network access is needed.
var finder tzf.F

// init initializes the timezone finder when the package is first imported.
//
// This function is called automatically by Go's runtime before any other
// code in this package runs. It creates the timezone finder with the default
// embedded timezone database.
//
// If initialization fails (extremely rare, would indicate a corrupted binary),
// the function panics. This is acceptable because:
//   - The error would occur at application startup, not during use
//   - There's no reasonable way to recover from a corrupted timezone database
//   - It's better to fail loudly than to produce incorrect results
func init() {
	var err error
	finder, err = tzf.NewDefaultFinder()
	if err != nil {
		// Panic on initialization failure - this should never happen with
		// a properly built binary. If it does, something is fundamentally wrong.
		panic("failed to initialize timezone finder: " + err.Error())
	}
}

// =============================================================================
// Public API
// =============================================================================

// FromCoordinates returns the IANA timezone identifier for the given coordinates.
//
// This function performs a geometric lookup to find which timezone boundary
// contains the specified point. The lookup is fast (typically < 1ms) and
// works entirely offline using embedded timezone data.
//
// Parameters:
//   - lat: Latitude in degrees (-90 to 90)
//   - lon: Longitude in degrees (-180 to 180)
//
// Returns:
//   - The IANA timezone identifier (e.g., "America/New_York", "Europe/Paris")
//   - "UTC" if the timezone cannot be determined (ocean, poles, etc.)
//
// Note: The tzf library uses (lon, lat) order internally, which is the opposite
// of the conventional (lat, lon) order. This function handles the conversion.
//
// Example:
//
//	tz := timezone.FromCoordinates(48.8566, 2.3522)
//	// tz = "Europe/Paris"
func FromCoordinates(lat, lon float64) string {
	// Note: tzf uses (lon, lat) order, which is geographic convention (x, y)
	// but opposite of the common (lat, lon) order used elsewhere in this app
	tz := finder.GetTimezoneName(lon, lat)

	// Fall back to UTC if no timezone found
	// This can happen for coordinates in:
	// - International waters (ocean)
	// - Antarctica (some areas have no civil timezone)
	// - Disputed or uninhabited territories
	if tz == "" {
		return "UTC"
	}
	return tz
}

// LoadLocation returns a *time.Location for the given coordinates.
//
// This is a convenience function that combines FromCoordinates with
// time.LoadLocation. It's useful when you need a time.Location for
// time zone conversions or formatting.
//
// Parameters:
//   - lat: Latitude in degrees (-90 to 90)
//   - lon: Longitude in degrees (-180 to 180)
//
// Returns:
//   - *time.Location for the coordinates' timezone
//   - time.UTC if the timezone cannot be determined or loaded
//
// The function handles two fallback scenarios:
//  1. FromCoordinates returns "UTC" (coordinates not in any timezone)
//  2. time.LoadLocation fails (timezone not in Go's tzdata, rare)
//
// Example:
//
//	loc := timezone.LoadLocation(48.8566, 2.3522)
//	parisTime := time.Now().In(loc)
func LoadLocation(lat, lon float64) *time.Location {
	// First, get the timezone name from coordinates
	tzName := FromCoordinates(lat, lon)

	// Then, load the time.Location from Go's timezone database
	// This uses the system timezone files or embedded tzdata
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		// This should rarely happen since FromCoordinates returns valid
		// IANA identifiers. Could occur if:
		// - tzf returns a timezone not in Go's database (very rare)
		// - System timezone files are missing and no embedded tzdata
		return time.UTC
	}

	return loc
}
