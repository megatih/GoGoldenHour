// Package geolocation provides IP-based location detection using the IP-API service.
//
// This package enables automatic location detection based on the user's public IP
// address. It's used when the "Auto-detect location on startup" setting is enabled,
// providing a convenient way to set an initial location without user input.
//
// # IP-API Service
//
// The package uses ip-api.com, a free geolocation API that requires no authentication.
// The free tier allows up to 45 requests per minute from a single IP address, which
// is more than sufficient for this application's use case (typically 1 request per
// app launch).
//
// # Accuracy Considerations
//
// IP-based geolocation has inherent limitations:
//
//   - Accuracy varies by region (typically city-level, not street-level)
//   - VPN or proxy users will get the VPN server's location
//   - Corporate networks may report the company's headquarters
//   - Mobile networks may report the carrier's regional hub
//
// For these reasons, the app also provides manual location search and map click
// functionality for users who need more precise location setting.
//
// # Security Note
//
// The IP-API endpoint uses HTTP (not HTTPS) by default. This is intentional for
// the free tier and is acceptable because:
//   - No sensitive user data is sent (the API only sees the source IP)
//   - The response contains only approximate geographic data
//   - HTTPS is available with a paid subscription if needed
package geolocation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// ipAPIEndpoint is the URL for the IP-API geolocation service.
	// This endpoint returns JSON data about the caller's location based on their
	// public IP address. No API key or authentication is required.
	//
	// The free tier uses HTTP; HTTPS requires a paid subscription.
	// Rate limit: 45 requests per minute from a single IP address.
	//
	// Documentation: https://ip-api.com/docs/api:json
	ipAPIEndpoint = "http://ip-api.com/json/"
)

// =============================================================================
// API Response Types
// =============================================================================

// ipAPIResponse represents the JSON response from ip-api.com.
//
// This struct maps to the IP-API JSON response format. Only fields needed by
// this application are included; the API returns additional fields that are
// ignored (ISP, ASN, etc.).
//
// Example successful response:
//
//	{
//	  "status": "success",
//	  "country": "United States",
//	  "countryCode": "US",
//	  "region": "CA",
//	  "regionName": "California",
//	  "city": "San Francisco",
//	  "lat": 37.7749,
//	  "lon": -122.4194,
//	  "timezone": "America/Los_Angeles",
//	  "query": "8.8.8.8"
//	}
//
// Example error response:
//
//	{
//	  "status": "fail",
//	  "message": "reserved range"
//	}
type ipAPIResponse struct {
	// Status is either "success" or "fail".
	// Always check this field before using other response data.
	Status string `json:"status"`

	// Message contains an error description when Status is "fail".
	// Common errors: "reserved range" (private IP), "invalid query"
	Message string `json:"message,omitempty"`

	// Country is the full country name (e.g., "United States", "Germany")
	Country string `json:"country"`

	// CountryCode is the ISO 3166-1 alpha-2 country code (e.g., "US", "DE")
	CountryCode string `json:"countryCode"`

	// Region is the region/state/province code (e.g., "CA", "BY")
	Region string `json:"region"`

	// RegionName is the full region name (e.g., "California", "Bavaria")
	RegionName string `json:"regionName"`

	// City is the city name (e.g., "San Francisco", "Munich")
	City string `json:"city"`

	// Lat is the latitude coordinate of the approximate location
	Lat float64 `json:"lat"`

	// Lon is the longitude coordinate of the approximate location
	Lon float64 `json:"lon"`

	// Timezone is the IANA timezone identifier (e.g., "America/Los_Angeles")
	Timezone string `json:"timezone"`

	// Query is the IP address that was looked up (useful for debugging)
	Query string `json:"query"`
}

// =============================================================================
// Service
// =============================================================================

// IPAPIService handles IP-based geolocation using the ip-api.com service.
//
// This service is created once at application startup and reused for any
// location detection requests. It maintains an HTTP client with a configured
// timeout to prevent hanging on network issues.
//
// Usage:
//
//	service := geolocation.NewIPAPIService()
//	location, err := service.DetectLocation()
//	if err != nil {
//	    // Handle error (network failure, API error, etc.)
//	    // Fall back to default location or last saved location
//	}
//	// Use location for solar calculations
type IPAPIService struct {
	// client is the HTTP client used for API requests.
	// Configured with a timeout from config.DefaultHTTPTimeout (10 seconds).
	client *http.Client
}

// NewIPAPIService creates a new IP geolocation service.
//
// The service is configured with a timeout from config.DefaultHTTPTimeout
// to prevent the application from hanging if the API is unreachable.
//
// Returns a ready-to-use IPAPIService instance.
func NewIPAPIService() *IPAPIService {
	return &IPAPIService{
		client: &http.Client{
			// Use the shared timeout constant for consistent network behavior
			Timeout: config.DefaultHTTPTimeout,
		},
	}
}

// DetectLocation attempts to detect the user's geographic location based on their IP address.
//
// This method makes an HTTP request to the IP-API service, which returns
// location data based on the public IP address of the request. The location
// is approximate (typically city-level) and may not be accurate for users
// behind VPNs or on mobile networks.
//
// Returns:
//   - domain.Location: The detected location with coordinates, name, and timezone
//   - error: Non-nil if detection fails (network error, API error, etc.)
//
// Error cases:
//   - Network timeout or connectivity issues
//   - API returns non-200 status code
//   - API returns "fail" status (e.g., reserved IP range)
//   - JSON parsing failure
//
// On error, callers should fall back to the default location (London, UK)
// or the user's last saved location.
func (s *IPAPIService) DetectLocation() (domain.Location, error) {
	// Make GET request to the IP-API endpoint.
	// The API uses the source IP address of the request to determine location.
	resp, err := s.client.Get(ipAPIEndpoint)
	if err != nil {
		// Network error (timeout, DNS failure, connection refused, etc.)
		return domain.Location{}, fmt.Errorf("failed to fetch location: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code (API should return 200 for all queries)
	if resp.StatusCode != http.StatusOK {
		return domain.Location{}, fmt.Errorf("IP-API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var apiResp ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.Location{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API-level status (may be "fail" even with HTTP 200)
	if apiResp.Status != "success" {
		msg := apiResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return domain.Location{}, fmt.Errorf("IP-API error: %s", msg)
	}

	// Build human-readable location name from city and country.
	// Handle cases where some fields might be empty.
	name := apiResp.City
	if apiResp.Country != "" {
		if name != "" {
			name += ", "
		}
		name += apiResp.Country
	}
	if name == "" {
		// Fallback for edge cases where both city and country are empty
		name = "Unknown Location"
	}

	// Build and return the domain.Location
	return domain.Location{
		Latitude:  apiResp.Lat,
		Longitude: apiResp.Lon,
		Elevation: 0, // IP-API doesn't provide elevation data
		Name:      name,
		Timezone:  apiResp.Timezone,
	}, nil
}
