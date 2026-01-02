// Package app provides the main application controller for GoGoldenHour.
//
// This package implements the central orchestrator that coordinates all
// components of the application. Following the controller pattern, it:
//
//   - Owns and manages all services (solar calculator, geocoding, geolocation)
//   - Maintains application state (current location, date, settings)
//   - Handles user actions via callbacks from the UI
//   - Coordinates data flow between services and UI components
//   - Persists user preferences to disk
//
// # Architecture
//
// The App controller sits at the center of the application:
//
//	                    ┌─────────────────┐
//	                    │   MainWindow    │
//	                    │   (UI Layer)    │
//	                    └────────┬────────┘
//	                             │ callbacks
//	                             ▼
//	┌──────────────┐    ┌─────────────────┐    ┌──────────────┐
//	│ Preferences  │◄───│      App        │───►│   Services   │
//	│   Store      │    │  (Controller)   │    │ (Solar, Geo) │
//	└──────────────┘    └─────────────────┘    └──────────────┘
//
// # Thread Safety
//
// The App controller is designed to be used from the main Qt thread.
// Asynchronous operations (network requests) are performed in goroutines,
// but all UI updates and state modifications happen on the main thread
// using mainthread.Wait() from the miqt library.
//
// This pattern ensures:
//   - UI remains responsive during network operations
//   - No race conditions on application state
//   - Proper Qt thread safety (widgets can only be modified from main thread)
//
// # Initialization Order
//
// The App constructor performs initialization in a specific order:
//  1. Load preferences (or use defaults if first run)
//  2. Create configuration with loaded preferences
//  3. Create all services (solar, geocoding, geolocation)
//  4. Restore last location (or use default)
//  5. Create main window (which creates all widgets)
//
// This order is important because:
//   - Services need settings for proper configuration
//   - MainWindow needs the App reference for callbacks
//   - Initial recalculation needs both location and services
package app

import (
	"fmt"
	"time"

	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
	"github.com/megatih/GoGoldenHour/internal/service/geocoding"
	"github.com/megatih/GoGoldenHour/internal/service/geolocation"
	"github.com/megatih/GoGoldenHour/internal/service/solar"
	"github.com/megatih/GoGoldenHour/internal/service/timezone"
	"github.com/megatih/GoGoldenHour/internal/storage"
	"github.com/megatih/GoGoldenHour/internal/ui"
)

// =============================================================================
// App Controller
// =============================================================================

// App is the main application controller that orchestrates all components.
//
// The App implements the ui.AppController interface, which defines the callbacks
// that the MainWindow uses to communicate user actions back to the controller.
//
// State Management:
//   - location: The currently selected geographic location
//   - currentDate: The date for solar calculations (defaults to today)
//   - config.Settings: User preferences (elevation angles, display format, etc.)
//
// All state modifications go through public methods that also:
//  1. Update the relevant UI components
//  2. Trigger recalculation if needed
//  3. Persist changes to disk when appropriate
type App struct {
	// config holds the complete application configuration including settings.
	// This is the authoritative source for current settings values.
	config config.AppConfig

	// prefs handles persistence of user settings to disk.
	// Settings are saved automatically when they change.
	prefs *storage.PreferencesStore

	// solarCalc performs all solar position and time calculations.
	// It maintains the current elevation angle settings for golden/blue hour.
	solarCalc *solar.Calculator

	// geoService provides IP-based location detection.
	// Used for auto-detect on startup if enabled in settings.
	geoService *geolocation.IPAPIService

	// geocoding provides address search and reverse geocoding.
	// Used for the location search feature and map click handling.
	geocoding *geocoding.NominatimService

	// mainWindow is the main UI controller.
	// The App calls its methods to update the display.
	mainWindow *ui.MainWindow

	// location is the currently selected geographic location.
	// Solar calculations are performed for this location.
	location domain.Location

	// currentDate is the date for which solar times are calculated.
	// Defaults to today, can be changed via the date picker.
	currentDate time.Time
}

// =============================================================================
// Constructor
// =============================================================================

// New creates a new application instance with all components initialized.
//
// Initialization steps:
//  1. Create and load preferences store
//  2. Load settings from disk (or use defaults)
//  3. Create configuration with loaded settings
//  4. Create all services with current settings
//  5. Restore last location or use default
//  6. Create main window with callback bindings
//
// Returns:
//   - *App: The fully initialized application controller
//   - error: Non-nil if initialization fails (rare, indicates system issues)
//
// The only failure case is if the preferences store cannot be created,
// which indicates a problem with the user's config directory.
func New() (*App, error) {
	// =========================================================================
	// Step 1: Initialize Preferences Storage
	// =========================================================================
	// Create the preferences store first, as we need it to load settings.
	// This also creates the config directory if it doesn't exist.
	prefs, err := storage.NewPreferencesStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create preferences store: %w", err)
	}

	// =========================================================================
	// Step 2: Load User Settings
	// =========================================================================
	// Load settings from disk. If the file doesn't exist (first run) or is
	// corrupted, Load() returns default settings.
	settings, err := prefs.Load()
	if err != nil {
		// This is a fallback that should rarely be needed, as Load() handles
		// most error cases internally by returning defaults.
		settings = domain.DefaultSettings()
	}

	// =========================================================================
	// Step 3: Create Configuration
	// =========================================================================
	// Start with default config (window size, app name) and merge in loaded settings.
	cfg := config.DefaultConfig()
	cfg.Settings = settings

	// =========================================================================
	// Step 4: Create Services
	// =========================================================================
	// Create all services that the application needs. Each service is
	// independent and can be used immediately after creation.
	solarCalc := solar.New(settings)
	geoService := geolocation.NewIPAPIService()
	geocodingService := geocoding.NewNominatimService()

	// =========================================================================
	// Step 5: Restore or Default Location
	// =========================================================================
	// Try to restore the user's last location from settings.
	// Fall back to default location (London) if no saved location.
	location := domain.DefaultLocation()
	if settings.LastLocation != nil {
		location = *settings.LastLocation
	}

	// =========================================================================
	// Step 6: Assemble Application
	// =========================================================================
	app := &App{
		config:      cfg,
		prefs:       prefs,
		solarCalc:   solarCalc,
		geoService:  geoService,
		geocoding:   geocodingService,
		location:    location,
		currentDate: time.Now(),
	}

	// =========================================================================
	// Step 7: Create Main Window
	// =========================================================================
	// Create the main window last, after the App is fully constructed.
	// The window receives a reference to the App for callbacks.
	//
	// IMPORTANT: The SettingsPanel may trigger callbacks during construction
	// (when applySettings is called). The recalculate() method checks for
	// mainWindow == nil to handle this case safely.
	mainWindow := ui.NewMainWindow(cfg, app)
	app.mainWindow = mainWindow

	return app, nil
}

// =============================================================================
// Application Lifecycle
// =============================================================================

// Run starts the application and makes it visible.
//
// This method should be called after New() returns successfully. It:
//  1. Shows the main window
//  2. Either auto-detects location or uses saved/default location
//  3. Performs initial solar calculations
//
// After Run() returns, the application is ready and the Qt event loop
// should be started with qt.QApplication_Exec().
func (a *App) Run() {
	// Show the main window to the user
	a.mainWindow.Show()

	// Determine initial location based on user preference
	if a.config.Settings.AutoDetectLocation {
		// Start async location detection
		// This will update the UI when complete
		a.DetectLocation()
	} else {
		// Use saved or default location and calculate sun times immediately
		a.recalculate()
	}
}

// =============================================================================
// Location Management
// =============================================================================

// DetectLocation attempts to detect the user's location using IP geolocation.
//
// This method runs asynchronously to avoid blocking the UI. The detection
// process:
//  1. Queries the IP-API service in a background goroutine
//  2. Waits for the main thread before updating UI
//  3. Either updates to detected location or falls back to default
//
// Thread Safety: Uses mainthread.Wait() to ensure UI updates happen on
// the Qt main thread.
func (a *App) DetectLocation() {
	// Run geolocation in background to keep UI responsive
	go func() {
		// Make network request to IP-API
		location, err := a.geoService.DetectLocation()

		// Switch back to main thread for UI updates
		mainthread.Wait(func() {
			if err != nil {
				// Show error to user but don't fail completely
				a.mainWindow.ShowError(fmt.Sprintf("Failed to detect location: %v", err))
				// Fall back to default location (London)
				a.UpdateLocation(domain.DefaultLocation())
				return
			}
			// Success - update to detected location
			a.UpdateLocation(location)
		})
	}()
}

// UpdateLocation updates the current location and triggers recalculation.
//
// This is the central method for location changes, called by:
//   - DetectLocation (after IP geolocation)
//   - SearchLocation (after geocoding search results)
//   - OnMapClick (after map click with reverse geocoding)
//
// The method performs these actions:
//  1. Updates the internal location state
//  2. Updates the UI to show the new location
//  3. Recalculates sun times for the new location
//  4. Saves the location as "last location" for future sessions
func (a *App) UpdateLocation(loc domain.Location) {
	// Update internal state
	a.location = loc

	// Update UI components (location panel, map)
	a.mainWindow.UpdateLocation(loc)

	// Recalculate sun times for new location
	a.recalculate()

	// Persist as last used location for next app launch
	a.config.Settings.LastLocation = &loc
	a.saveSettings()
}

// =============================================================================
// Date Management
// =============================================================================

// UpdateDate changes the date for solar calculations and updates the display.
//
// This is called when the user:
//   - Clicks previous/next date buttons
//   - Selects a date from the calendar picker
//   - Clicks the "Today" button
//
// The method updates the date state, UI display, and recalculates sun times.
func (a *App) UpdateDate(date time.Time) {
	// Update internal state
	a.currentDate = date

	// Update UI date display
	a.mainWindow.UpdateDate(date)

	// Recalculate sun times for new date
	a.recalculate()
}

// =============================================================================
// Settings Management
// =============================================================================

// UpdateSettings applies new settings and triggers necessary updates.
//
// This is called when the user changes settings in the settings panel,
// such as elevation angles or time format preferences.
//
// The method:
//  1. Updates the configuration with new settings
//  2. Updates the solar calculator with new elevation angles
//  3. Saves settings to disk for persistence
//  4. Recalculates sun times with new parameters
func (a *App) UpdateSettings(settings domain.Settings) {
	// Update configuration
	a.config.Settings = settings

	// Update solar calculator with new elevation angles
	// This is necessary because the calculator caches the settings
	a.solarCalc.UpdateSettings(settings)

	// Persist to disk
	a.saveSettings()

	// Recalculate with new settings (may change golden/blue hour times)
	a.recalculate()
}

// =============================================================================
// Location Search
// =============================================================================

// SearchLocation performs a geocoding search and updates to the first result.
//
// This is called when the user types a location query and presses Enter or
// clicks the Search button. The search runs asynchronously to keep the UI
// responsive.
//
// Search flow:
//  1. Query the Nominatim geocoding service (background)
//  2. Wait for main thread
//  3. If successful, update to first result
//  4. If failed or no results, show error message
//
// Thread Safety: Uses mainthread.Wait() for UI updates.
func (a *App) SearchLocation(query string) {
	// Run geocoding in background
	go func() {
		// Search for up to 5 matching locations
		locations, err := a.geocoding.Search(query, 5)

		// Switch back to main thread for UI updates
		mainthread.Wait(func() {
			if err != nil {
				a.mainWindow.ShowError(fmt.Sprintf("Search failed: %v", err))
				return
			}
			if len(locations) == 0 {
				a.mainWindow.ShowError("No locations found")
				return
			}
			// Use the first (most relevant) result
			a.UpdateLocation(locations[0])
		})
	}()
}

// =============================================================================
// Map Interaction
// =============================================================================

// OnMapClick handles map click events by reverse geocoding the clicked location.
//
// When the user clicks on the map, this method:
//  1. Attempts to reverse geocode the coordinates to get a place name
//  2. Creates a location with the coordinates and name
//  3. Falls back to coordinate display if reverse geocoding fails
//  4. Updates to the new location
//
// The reverse geocoding is optional - the app works fine with just coordinates.
// This is why errors from ReverseGeocode are intentionally ignored.
//
// Thread Safety: Uses mainthread.Wait() for UI updates.
func (a *App) OnMapClick(lat, lon float64) {
	// Reverse geocode in background
	go func() {
		// Try to get a human-readable name for the coordinates.
		// Error is intentionally ignored - we fall back to coordinate display.
		name, _ := a.geocoding.ReverseGeocode(lat, lon)

		// Switch back to main thread for UI updates
		mainthread.Wait(func() {
			// Build location with timezone from coordinates
			loc := domain.Location{
				Latitude:  lat,
				Longitude: lon,
				Name:      name,
				Timezone:  timezone.FromCoordinates(lat, lon),
			}

			// Fall back to coordinate display if no name was found
			if name == "" {
				loc.Name = fmt.Sprintf("%.4f, %.4f", lat, lon)
			}

			a.UpdateLocation(loc)
		})
	}()
}

// =============================================================================
// State Getters (implements ui.AppController interface)
// =============================================================================

// GetSettings returns the current settings.
//
// This is part of the ui.AppController interface, allowing the UI to query
// current settings values (e.g., for initializing the settings panel).
func (a *App) GetSettings() domain.Settings {
	return a.config.Settings
}

// GetLocation returns the current location.
//
// This is part of the ui.AppController interface, allowing the UI to query
// the current location (e.g., for displaying in the location panel).
func (a *App) GetLocation() domain.Location {
	return a.location
}

// GetDate returns the current date for calculations.
//
// This is part of the ui.AppController interface, allowing the UI to query
// the current date (e.g., for initializing the date picker).
func (a *App) GetDate() time.Time {
	return a.currentDate
}

// =============================================================================
// Internal Methods
// =============================================================================

// recalculate performs solar calculations and updates the UI with results.
//
// This is called whenever the location, date, or settings change. It:
//  1. Calculates sun times using the solar calculator
//  2. Updates the UI to display the new times
//  3. Shows an error if calculation fails (rare)
//
// IMPORTANT: This method checks if mainWindow is nil because it may be called
// during initialization when the SettingsPanel triggers OnValueChanged callbacks.
// At that point, the mainWindow hasn't been assigned to the App yet.
func (a *App) recalculate() {
	// Guard against calls during initialization
	// (SettingsPanel triggers callbacks before mainWindow is set)
	if a.mainWindow == nil {
		return
	}

	// Calculate sun times for current location and date
	sunTimes, err := a.solarCalc.Calculate(a.location, a.currentDate)
	if err != nil {
		// Calculation errors are rare with valid input, but handle them
		a.mainWindow.ShowError(fmt.Sprintf("Calculation error: %v", err))
		return
	}

	// Update the time display panel with calculated values
	a.mainWindow.UpdateSunTimes(sunTimes)
}

// saveSettings persists the current settings to disk.
//
// This is called whenever settings change, including:
//   - User changes settings in the settings panel
//   - Location is updated (saves as LastLocation)
//
// Errors are displayed to the user but don't prevent the app from functioning.
// The app can continue working even if settings can't be saved; they just
// won't persist to the next session.
func (a *App) saveSettings() {
	if err := a.prefs.Save(a.config.Settings); err != nil && a.mainWindow != nil {
		// Only show error if mainWindow exists (avoid error during init)
		a.mainWindow.ShowError(fmt.Sprintf("Failed to save settings: %v", err))
	}
}
