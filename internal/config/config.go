package config

import (
	"time"

	"github.com/megatih/GoGoldenHour/internal/domain"
)

// HTTP client configuration
const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests
	DefaultHTTPTimeout = 10 * time.Second
)

// AppConfig holds the complete application configuration
type AppConfig struct {
	// Window dimensions
	WindowWidth  int
	WindowHeight int

	// Application metadata
	AppName    string
	AppVersion string

	// User settings
	Settings domain.Settings
}

// DefaultConfig returns the default application configuration
func DefaultConfig() AppConfig {
	return AppConfig{
		WindowWidth:  800,
		WindowHeight: 600,
		AppName:      "GoGoldenHour",
		AppVersion:   "1.0.0",
		Settings:     domain.DefaultSettings(),
	}
}
