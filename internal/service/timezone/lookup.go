package timezone

import (
	"time"

	"github.com/ringsaturn/tzf"
)

var finder tzf.F

func init() {
	var err error
	finder, err = tzf.NewDefaultFinder()
	if err != nil {
		panic("failed to initialize timezone finder: " + err.Error())
	}
}

// FromCoordinates returns the IANA timezone identifier for the given coordinates.
// Falls back to "UTC" if the timezone cannot be determined.
func FromCoordinates(lat, lon float64) string {
	tz := finder.GetTimezoneName(lon, lat)
	if tz == "" {
		return "UTC"
	}
	return tz
}

// LoadLocation returns the *time.Location for the given coordinates.
// Falls back to time.UTC if the timezone cannot be loaded.
func LoadLocation(lat, lon float64) *time.Location {
	tzName := FromCoordinates(lat, lon)
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return time.UTC
	}
	return loc
}
