package config

import (
	"github.com/megatih/GoGoldenHour/internal/domain"
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
