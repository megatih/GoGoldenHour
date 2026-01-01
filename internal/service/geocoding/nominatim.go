package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/megatih/GoGoldenHour/internal/domain"
	"github.com/megatih/GoGoldenHour/internal/service/timezone"
)

const (
	// Nominatim OpenStreetMap geocoding endpoint
	nominatimEndpoint = "https://nominatim.openstreetmap.org/search"
	// Request timeout
	requestTimeout = 10 * time.Second
	// User agent required by Nominatim
	userAgent = "GoGoldenHour/1.0 (https://github.com/megatih/GoGoldenHour)"
)

// nominatimResult represents a single search result from Nominatim
type nominatimResult struct {
	PlaceID     int64   `json:"place_id"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	Importance  float64 `json:"importance"`
}

// NominatimService handles geocoding (address to coordinates)
type NominatimService struct {
	client *http.Client
}

// NewNominatimService creates a new geocoding service
func NewNominatimService() *NominatimService {
	return &NominatimService{
		client: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Search searches for locations matching the query string
func (s *NominatimService) Search(query string, limit int) ([]domain.Location, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 || limit > 10 {
		limit = 5
	}

	// Build request URL
	reqURL, err := url.Parse(nominatimEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("limit", fmt.Sprintf("%d", limit))
	reqURL.RawQuery = q.Encode()

	// Create request with required User-Agent
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Nominatim returned status %d", resp.StatusCode)
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to domain locations
	locations := make([]domain.Location, 0, len(results))
	for _, r := range results {
		var lat, lon float64
		fmt.Sscanf(r.Lat, "%f", &lat)
		fmt.Sscanf(r.Lon, "%f", &lon)

		locations = append(locations, domain.Location{
			Latitude:  lat,
			Longitude: lon,
			Elevation: 0, // Nominatim doesn't provide elevation
			Name:      r.DisplayName,
			Timezone:  timezone.FromCoordinates(lat, lon),
		})
	}

	return locations, nil
}

// ReverseGeocode converts coordinates to a location name
func (s *NominatimService) ReverseGeocode(lat, lon float64) (string, error) {
	reqURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=json", lat, lon)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reverse geocode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Nominatim returned status %d", resp.StatusCode)
	}

	var result struct {
		DisplayName string `json:"display_name"`
		Error       string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("Nominatim error: %s", result.Error)
	}

	return result.DisplayName, nil
}
