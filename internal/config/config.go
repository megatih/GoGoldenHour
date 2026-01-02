// Package config provides centralized application configuration.
//
// This package contains shared constants and configuration structures used
// throughout the application. It serves as a single source of truth for:
//
//   - HTTP client settings (timeouts shared by geolocation and geocoding services)
//   - Application window dimensions
//   - Application metadata (name, version)
//   - User settings integration
//
// The configuration follows a layered approach:
//
//  1. Compile-time constants (DefaultHTTPTimeout) - cannot be changed at runtime
//  2. AppConfig - combines user settings with fixed application parameters
//  3. domain.Settings - user-configurable values loaded from disk
//
// Centralizing HTTP timeout here ensures consistent behavior across all external
// API calls (IP-API for geolocation, Nominatim for geocoding) and makes it easy
// to adjust timeout values during development or for different network conditions.
package config

import (
	"time"

	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// Shared HTTP Configuration
// =============================================================================

// HTTP client configuration constants shared across services.
// These values are used by both the geolocation (IP-API) and geocoding
// (Nominatim) services to ensure consistent network behavior.
const (
	// DefaultHTTPTimeout is the maximum time to wait for HTTP requests.
	//
	// This timeout applies to the complete request-response cycle, including:
	//   - DNS resolution
	//   - TCP connection establishment
	//   - TLS handshake (for HTTPS)
	//   - Sending the request
	//   - Receiving the response
	//
	// 10 seconds is a reasonable default that:
	//   - Allows for slow network conditions
	//   - Prevents the UI from hanging indefinitely
	//   - Matches typical expectations for responsive applications
	//
	// If this timeout is exceeded, the HTTP request will fail with a timeout
	// error, and the calling code should handle the failure gracefully
	// (e.g., show an error message, use cached/default data).
	//
	// Used by:
	//   - internal/service/geolocation/ipapi.go (IP-based location detection)
	//   - internal/service/geocoding/nominatim.go (address search, reverse geocoding)
	DefaultHTTPTimeout = 10 * time.Second
)

// =============================================================================
// Application Configuration
// =============================================================================

// AppConfig holds the complete application configuration.
//
// This struct combines fixed application parameters (window size, metadata)
// with user-configurable settings from domain.Settings. It is created at
// application startup and passed to the App controller.
//
// Configuration flow:
//
//  1. DefaultConfig() creates initial configuration with sensible defaults
//  2. PreferencesStore.Load() loads user settings from disk
//  3. Loaded settings are merged into AppConfig.Settings
//  4. AppConfig is passed to App.New() for initialization
//
// The AppConfig is owned by the App controller and may be modified during
// runtime when the user changes settings. Changes are persisted via
// PreferencesStore.Save().
type AppConfig struct {
	// WindowWidth is the initial/minimum width of the main window in pixels.
	// The actual window may be larger if the user resizes it, but it cannot
	// be made smaller than this value (enforced by Qt setMinimumSize).
	//
	// Default: 800 pixels (provides comfortable space for map + info panels)
	WindowWidth int

	// WindowHeight is the initial/minimum height of the main window in pixels.
	// Same behavior as WindowWidth regarding resizing constraints.
	//
	// Default: 600 pixels
	WindowHeight int

	// AppName is the application's display name shown in the window title bar.
	// This is also used by Qt for platform integration (taskbar, app switcher).
	//
	// Default: "GoGoldenHour"
	AppName string

	// AppVersion is the current version string for the application.
	// This may be displayed in the UI or used for update checking in the future.
	//
	// Default: "1.0.0"
	AppVersion string

	// Settings holds user-configurable preferences.
	// These are loaded from disk on startup and saved when the user changes them.
	// See domain.Settings for detailed documentation of each setting.
	Settings domain.Settings
}

// DefaultConfig returns the default application configuration.
//
// This provides sensible defaults for a fresh installation where no user
// preferences have been saved yet. The returned configuration includes:
//
//   - Window size: 800x600 pixels (comfortable for desktop use)
//   - App name/version: "GoGoldenHour" v1.0.0
//   - Settings: domain.DefaultSettings() (see that function for details)
//
// The defaults are designed to work well on most systems and provide a good
// first-run experience. Users can customize all settings after launching.
func DefaultConfig() AppConfig {
	return AppConfig{
		WindowWidth:  800,
		WindowHeight: 600,
		AppName:      "GoGoldenHour",
		AppVersion:   "1.0.0",
		Settings:     domain.DefaultSettings(),
	}
}
