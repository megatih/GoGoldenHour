// Package geocoding provides address-to-coordinates conversion using OpenStreetMap's Nominatim.
//
// This package enables two key features in GoGoldenHour:
//
//  1. Forward Geocoding (Search): Convert a text query (city name, address) to
//     geographic coordinates. Used when the user types a location in the search box.
//
//  2. Reverse Geocoding: Convert coordinates to a human-readable place name.
//     Used when the user clicks on the map to get the location name.
//
// # Nominatim Service
//
// The package uses Nominatim, the geocoding service provided by OpenStreetMap.
// Nominatim is free to use with the following requirements:
//
//   - Maximum 1 request per second (we're well within this with user interactions)
//   - Required User-Agent header identifying the application
//   - No bulk/automated queries (interactive use only)
//
// Documentation: https://nominatim.org/release-docs/latest/api/Overview/
//
// # Timezone Integration
//
// When converting search results to domain.Location, the package automatically
// determines the timezone for each location using the timezone package. This
// ensures that solar calculations use the correct local time.
//
// # Accuracy
//
// Nominatim provides precise geocoding based on OpenStreetMap data:
//   - Street addresses resolve to exact locations
//   - City/region names resolve to city centers
//   - Named places (parks, landmarks) resolve to their locations
//
// The quality depends on OpenStreetMap coverage, which is generally excellent
// in populated areas but may be sparse in remote regions.
package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
	"github.com/megatih/GoGoldenHour/internal/service/timezone"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// nominatimSearchEndpoint is the URL for forward geocoding (text to coordinates).
	// Accepts query parameters: q (search text), format (json/xml), limit (max results)
	nominatimSearchEndpoint = "https://nominatim.openstreetmap.org/search"

	// nominatimReverseEndpoint is the URL for reverse geocoding (coordinates to text).
	// Accepts query parameters: lat, lon, format (json/xml)
	nominatimReverseEndpoint = "https://nominatim.openstreetmap.org/reverse"

	// userAgent is the required User-Agent header for Nominatim requests.
	// Nominatim's usage policy requires a valid User-Agent that identifies
	// the application and provides contact information.
	// See: https://operations.osmfoundation.org/policies/nominatim/
	userAgent = "GoGoldenHour/1.0 (https://github.com/megatih/GoGoldenHour)"
)

// =============================================================================
// API Response Types
// =============================================================================

// nominatimResult represents a single search result from the Nominatim API.
//
// Nominatim returns results sorted by relevance (importance score). Each result
// contains coordinates as strings (not numbers) and a display name that combines
// multiple address components.
//
// Example result:
//
//	{
//	  "place_id": 282616932,
//	  "lat": "48.8588897",
//	  "lon": "2.3200410",
//	  "display_name": "Paris, Île-de-France, France métropolitaine, France",
//	  "type": "city",
//	  "importance": 0.9411
//	}
type nominatimResult struct {
	// PlaceID is Nominatim's internal identifier for this place.
	// Not used by this application but included for completeness.
	PlaceID int64 `json:"place_id"`

	// Lat is the latitude coordinate as a string (Nominatim quirk).
	// Must be parsed to float64 for use in domain.Location.
	Lat string `json:"lat"`

	// Lon is the longitude coordinate as a string (Nominatim quirk).
	// Must be parsed to float64 for use in domain.Location.
	Lon string `json:"lon"`

	// DisplayName is a human-readable location string combining multiple
	// address components (city, region, country, etc.).
	// Example: "Eiffel Tower, Champ de Mars, 7th Arrondissement, Paris, France"
	DisplayName string `json:"display_name"`

	// Type indicates the OSM object type (city, street, building, etc.).
	// Not currently used but useful for filtering or categorizing results.
	Type string `json:"type"`

	// Importance is a score indicating result relevance (0.0 to 1.0).
	// Higher values = more relevant/important places.
	// Results are sorted by this value in descending order.
	Importance float64 `json:"importance"`
}

// =============================================================================
// Service
// =============================================================================

// NominatimService handles geocoding operations using OpenStreetMap's Nominatim API.
//
// This service provides both forward and reverse geocoding:
//   - Search: Convert text queries to geographic coordinates
//   - ReverseGeocode: Convert coordinates to place names
//
// The service maintains an HTTP client with configured timeout and automatically
// includes the required User-Agent header for all requests.
//
// Usage:
//
//	service := geocoding.NewNominatimService()
//
//	// Forward geocoding (search)
//	locations, err := service.Search("Eiffel Tower", 5)
//
//	// Reverse geocoding (map click)
//	name, err := service.ReverseGeocode(48.8588, 2.3200)
type NominatimService struct {
	// client is the HTTP client used for API requests.
	// Configured with a timeout from config.DefaultHTTPTimeout (10 seconds).
	client *http.Client
}

// NewNominatimService creates a new geocoding service.
//
// The service is configured with a timeout from config.DefaultHTTPTimeout
// to prevent the application from hanging if the API is unreachable.
//
// Returns a ready-to-use NominatimService instance.
func NewNominatimService() *NominatimService {
	return &NominatimService{
		client: &http.Client{
			Timeout: config.DefaultHTTPTimeout,
		},
	}
}

// =============================================================================
// Internal Helper
// =============================================================================

// doRequest performs an HTTP GET request with the required User-Agent header.
//
// This helper method centralizes the HTTP request logic for both Search and
// ReverseGeocode methods. It handles:
//   - Setting the required User-Agent header (Nominatim policy compliance)
//   - Executing the request with the configured timeout
//   - Checking for HTTP-level errors
//
// Parameters:
//   - reqURL: The complete URL to request (with query parameters)
//
// Returns:
//   - *http.Response: The response (caller must close Body)
//   - error: Non-nil if request fails or returns non-200 status
//
// Note: The caller is responsible for closing resp.Body when done.
func (s *NominatimService) doRequest(reqURL string) (*http.Response, error) {
	// Create request object so we can add custom headers
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Nominatim requires a valid User-Agent header that identifies the application.
	// Requests without User-Agent may be blocked or rate-limited more aggressively.
	req.Header.Set("User-Agent", userAgent)

	// Execute the request with the configured timeout
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Check HTTP status (Nominatim returns 200 for successful requests)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close() // Clean up before returning error
		return nil, fmt.Errorf("Nominatim returned status %d", resp.StatusCode)
	}

	return resp, nil
}

// =============================================================================
// Forward Geocoding (Search)
// =============================================================================

// Search finds locations matching the given query string.
//
// This method performs forward geocoding, converting a text query (city name,
// address, landmark, etc.) into geographic coordinates. Results are sorted by
// relevance, with the most likely match first.
//
// The method automatically:
//   - Validates input parameters
//   - URL-encodes the query string
//   - Determines timezones for each result using the timezone package
//
// Parameters:
//   - query: The search text (city name, address, etc.). Cannot be empty.
//   - limit: Maximum number of results to return (1-10, default 5)
//
// Returns:
//   - []domain.Location: Matching locations with coordinates, names, and timezones
//   - error: Non-nil if search fails
//
// Example:
//
//	locations, err := service.Search("Paris, France", 5)
//	if err != nil {
//	    // Handle error
//	}
//	// Use locations[0] as the primary result
func (s *NominatimService) Search(query string, limit int) ([]domain.Location, error) {
	// Validate query - empty queries are not allowed
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Validate and default the limit parameter
	// Nominatim supports up to 50 results, but we cap at 10 for UI simplicity
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	// Build the request URL with query parameters
	reqURL, err := url.Parse(nominatimSearchEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Set query parameters
	// - q: the search query (URL-encoded by url.Values)
	// - format: response format (json)
	// - limit: maximum number of results
	q := reqURL.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("limit", strconv.Itoa(limit))
	reqURL.RawQuery = q.Encode()

	// Execute the request
	resp, err := s.doRequest(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response (array of results)
	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Nominatim results to domain.Location objects
	locations := make([]domain.Location, 0, len(results))
	for _, r := range results {
		// Parse coordinates from strings to floats
		// Nominatim returns coordinates as strings (API quirk)
		lat, _ := strconv.ParseFloat(r.Lat, 64)
		lon, _ := strconv.ParseFloat(r.Lon, 64)

		locations = append(locations, domain.Location{
			Latitude:  lat,
			Longitude: lon,
			Elevation: 0, // Nominatim doesn't provide elevation data
			Name:      r.DisplayName,
			// Automatically determine timezone from coordinates
			// This is crucial for accurate solar calculations
			Timezone: timezone.FromCoordinates(lat, lon),
		})
	}

	return locations, nil
}

// =============================================================================
// Reverse Geocoding
// =============================================================================

// ReverseGeocode converts geographic coordinates to a human-readable place name.
//
// This method is used when the user clicks on the map to determine the name
// of the clicked location. It queries Nominatim's reverse geocoding endpoint
// to get the address or place name at the specified coordinates.
//
// Parameters:
//   - lat: Latitude of the point to reverse geocode
//   - lon: Longitude of the point to reverse geocode
//
// Returns:
//   - string: The display name for the location (address or place name)
//   - error: Non-nil if reverse geocoding fails
//
// Error cases:
//   - Network errors or timeouts
//   - Coordinates in the ocean or uninhabited areas (no data available)
//   - API errors
//
// Example:
//
//	name, err := service.ReverseGeocode(48.8588, 2.3200)
//	// name = "Eiffel Tower, Champ de Mars, 7th Arrondissement, Paris, France"
func (s *NominatimService) ReverseGeocode(lat, lon float64) (string, error) {
	// Build the request URL with coordinate parameters
	reqURL, err := url.Parse(nominatimReverseEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Set query parameters
	// - lat, lon: coordinates (full precision, no rounding)
	// - format: response format (json)
	q := reqURL.Query()
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	q.Set("format", "json")
	reqURL.RawQuery = q.Encode()

	// Execute the request
	resp, err := s.doRequest(reqURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to reverse geocode: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response
	// Reverse geocoding returns a single object (not an array like search)
	var result struct {
		DisplayName string `json:"display_name"`
		Error       string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API-level errors (e.g., "Unable to geocode")
	if result.Error != "" {
		return "", fmt.Errorf("Nominatim error: %s", result.Error)
	}

	return result.DisplayName, nil
}
