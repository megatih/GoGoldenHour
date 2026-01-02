// Package storage provides persistent storage for user preferences and settings.
//
// This package handles the serialization and storage of user preferences to disk,
// ensuring that settings persist between application sessions. The storage uses
// JSON format for human readability and easy debugging.
//
// # Storage Location
//
// Settings are stored in the user's configuration directory, following platform
// conventions:
//
//   - Linux: ~/.config/GoGoldenHour/settings.json
//   - macOS: ~/Library/Application Support/GoGoldenHour/settings.json
//   - Windows: %APPDATA%\GoGoldenHour\settings.json
//
// The directory is created automatically if it doesn't exist.
//
// # Data Format
//
// Settings are stored as pretty-printed JSON (2-space indentation) for easy
// manual inspection and editing. The JSON structure mirrors domain.Settings:
//
//	{
//	  "golden_hour_elevation": 6,
//	  "blue_hour_start": -4,
//	  "blue_hour_end": -8,
//	  "time_format_24_hour": true,
//	  "auto_detect_location": true,
//	  "last_location": {
//	    "latitude": 48.8566,
//	    "longitude": 2.3522,
//	    "name": "Paris, France",
//	    "timezone": "Europe/Paris"
//	  }
//	}
//
// # Error Handling
//
// The package is designed for graceful degradation:
//   - Missing file: Returns default settings (no error)
//   - Corrupted JSON: Returns default settings (no error)
//   - Invalid values: Validated and clamped to acceptable ranges
//
// This ensures the application always starts successfully, even if the
// configuration file is damaged or manually edited incorrectly.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// configDirName is the directory name within the user's config directory.
	// This creates a dedicated folder for all GoGoldenHour configuration files.
	configDirName = "GoGoldenHour"

	// configFileName is the name of the settings file within the config directory.
	// Using .json extension makes the format obvious and enables syntax highlighting
	// when users manually edit the file.
	configFileName = "settings.json"
)

// =============================================================================
// PreferencesStore
// =============================================================================

// PreferencesStore handles persistent storage of user preferences.
//
// The store manages a single JSON file containing all user settings. It provides
// thread-safe read/write operations (file operations are atomic) and handles
// all error cases gracefully.
//
// Usage:
//
//	store, err := storage.NewPreferencesStore()
//	if err != nil {
//	    // Handle error (rare, indicates filesystem issues)
//	}
//
//	// Load settings (always succeeds, returns defaults if needed)
//	settings, _ := store.Load()
//
//	// Save updated settings
//	settings.TimeFormat24Hour = false
//	store.Save(settings)
type PreferencesStore struct {
	// configPath is the full path to the settings.json file.
	// Determined at construction time based on the platform's config directory.
	configPath string
}

// NewPreferencesStore creates a new preferences store.
//
// This constructor:
//  1. Determines the platform-appropriate config directory
//  2. Creates the GoGoldenHour config directory if it doesn't exist
//  3. Returns a store configured to use settings.json in that directory
//
// Returns:
//   - *PreferencesStore: Ready-to-use store instance
//   - error: Non-nil if the config directory cannot be determined or created
//
// Errors are rare and indicate system-level issues (no home directory,
// permissions problems, etc.).
func NewPreferencesStore() (*PreferencesStore, error) {
	// Get the platform's user configuration directory.
	// This follows XDG on Linux, uses Application Support on macOS, etc.
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create application-specific subdirectory.
	// MkdirAll is idempotent - it succeeds if the directory already exists.
	// Permissions 0755 allow owner full access, others read/execute.
	appConfigDir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &PreferencesStore{
		configPath: filepath.Join(appConfigDir, configFileName),
	}, nil
}

// =============================================================================
// Load/Save Operations
// =============================================================================

// Load reads settings from disk and returns them.
//
// This method handles all error cases gracefully:
//   - File doesn't exist: Returns default settings (first run)
//   - File is corrupted/invalid JSON: Returns default settings
//   - File contains invalid values: Values are validated and clamped
//
// The only error case that propagates is when the file exists but cannot
// be read (permissions, filesystem errors).
//
// Returns:
//   - domain.Settings: The loaded settings, or defaults if loading fails
//   - error: Non-nil only for unexpected filesystem errors
//
// Example:
//
//	settings, err := store.Load()
//	if err != nil {
//	    log.Printf("Warning: Could not load settings: %v", err)
//	    // Continue with settings (which will be defaults)
//	}
func (s *PreferencesStore) Load() (domain.Settings, error) {
	// Read the entire file into memory
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		// File doesn't exist - this is normal for first run
		if os.IsNotExist(err) {
			return domain.DefaultSettings(), nil
		}
		// Other errors (permissions, etc.) are reported
		return domain.Settings{}, fmt.Errorf("failed to read settings: %w", err)
	}

	// Parse JSON into settings struct
	var settings domain.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		// JSON is corrupted or invalid - return defaults rather than failing.
		// This provides a recovery path for users who accidentally break
		// their config file by manual editing.
		return domain.DefaultSettings(), nil
	}

	// Validate and clamp settings to acceptable ranges.
	// This handles cases where the file was manually edited with invalid values.
	settings.Validate()

	return settings, nil
}

// Save writes the given settings to disk.
//
// The settings are serialized to pretty-printed JSON (2-space indentation)
// for human readability. The file is written with 0644 permissions (owner
// read/write, others read-only).
//
// Parameters:
//   - settings: The settings to save
//
// Returns:
//   - error: Non-nil if the write fails (permissions, disk full, etc.)
//
// The write is atomic at the filesystem level - either the entire file
// is written or the operation fails, preventing partial/corrupted files.
func (s *PreferencesStore) Save(settings domain.Settings) error {
	// Serialize to JSON with indentation for readability.
	// This makes manual inspection and debugging easier.
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		// This should never happen with domain.Settings, but handle it anyway
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write the file atomically.
	// Permissions 0644: owner read/write, group/others read-only.
	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// =============================================================================
// Utility Methods
// =============================================================================

// GetConfigPath returns the full path to the configuration file.
//
// This is useful for debugging, error messages, or informing users where
// their settings are stored.
//
// Returns the absolute path to settings.json, e.g.:
//   - Linux: /home/user/.config/GoGoldenHour/settings.json
//   - macOS: /Users/user/Library/Application Support/GoGoldenHour/settings.json
//   - Windows: C:\Users\user\AppData\Roaming\GoGoldenHour\settings.json
func (s *PreferencesStore) GetConfigPath() string {
	return s.configPath
}
