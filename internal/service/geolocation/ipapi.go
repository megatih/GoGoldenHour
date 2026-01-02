package geolocation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

const (
	// IP-API endpoint (free tier, no API key needed)
	ipAPIEndpoint = "http://ip-api.com/json/"
)

// ipAPIResponse represents the JSON response from ip-api.com
type ipAPIResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Query       string  `json:"query"`
}

// IPAPIService handles IP-based geolocation
type IPAPIService struct {
	client *http.Client
}

// NewIPAPIService creates a new IP geolocation service
func NewIPAPIService() *IPAPIService {
	return &IPAPIService{
		client: &http.Client{
			Timeout: config.DefaultHTTPTimeout,
		},
	}
}

// DetectLocation attempts to detect the user's location based on their IP address
func (s *IPAPIService) DetectLocation() (domain.Location, error) {
	resp, err := s.client.Get(ipAPIEndpoint)
	if err != nil {
		return domain.Location{}, fmt.Errorf("failed to fetch location: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Location{}, fmt.Errorf("IP-API returned status %d", resp.StatusCode)
	}

	var apiResp ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.Location{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Status != "success" {
		msg := apiResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return domain.Location{}, fmt.Errorf("IP-API error: %s", msg)
	}

	// Build location name
	name := apiResp.City
	if apiResp.Country != "" {
		if name != "" {
			name += ", "
		}
		name += apiResp.Country
	}
	if name == "" {
		name = "Unknown Location"
	}

	return domain.Location{
		Latitude:  apiResp.Lat,
		Longitude: apiResp.Lon,
		Elevation: 0, // IP-API doesn't provide elevation
		Name:      name,
		Timezone:  apiResp.Timezone,
	}, nil
}
