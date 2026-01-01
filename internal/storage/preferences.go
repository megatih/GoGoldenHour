package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/megatih/GoGoldenHour/internal/domain"
)

const (
	configDirName  = "GoGoldenHour"
	configFileName = "settings.json"
)

// PreferencesStore handles persistent storage of user preferences
type PreferencesStore struct {
	configPath string
}

// NewPreferencesStore creates a new preferences store
func NewPreferencesStore() (*PreferencesStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	appConfigDir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &PreferencesStore{
		configPath: filepath.Join(appConfigDir, configFileName),
	}, nil
}

// Load loads settings from disk, returning defaults if file doesn't exist
func (s *PreferencesStore) Load() (domain.Settings, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.DefaultSettings(), nil
		}
		return domain.Settings{}, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings domain.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		// If file is corrupted, return defaults
		return domain.DefaultSettings(), nil
	}

	// Validate loaded settings
	settings.Validate()

	return settings, nil
}

// Save saves settings to disk
func (s *PreferencesStore) Save(settings domain.Settings) error {
	settings.Validate()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func (s *PreferencesStore) GetConfigPath() string {
	return s.configPath
}
