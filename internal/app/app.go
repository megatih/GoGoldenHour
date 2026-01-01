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
	"github.com/megatih/GoGoldenHour/internal/storage"
	"github.com/megatih/GoGoldenHour/internal/ui"
)

// App is the main application controller
type App struct {
	config      config.AppConfig
	prefs       *storage.PreferencesStore
	solarCalc   *solar.Calculator
	geoService  *geolocation.IPAPIService
	geocoding   *geocoding.NominatimService
	mainWindow  *ui.MainWindow
	location    domain.Location
	currentDate time.Time
}

// New creates a new application instance
func New() (*App, error) {
	// Load preferences
	prefs, err := storage.NewPreferencesStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create preferences store: %w", err)
	}

	settings, err := prefs.Load()
	if err != nil {
		settings = domain.DefaultSettings()
	}

	// Create config
	cfg := config.DefaultConfig()
	cfg.Settings = settings

	// Create services
	solarCalc := solar.New(settings)
	geoService := geolocation.NewIPAPIService()
	geocodingService := geocoding.NewNominatimService()

	// Set initial location
	location := domain.DefaultLocation()
	if settings.LastLocation != nil {
		location = *settings.LastLocation
	}

	app := &App{
		config:      cfg,
		prefs:       prefs,
		solarCalc:   solarCalc,
		geoService:  geoService,
		geocoding:   geocodingService,
		location:    location,
		currentDate: time.Now(),
	}

	// Create main window
	mainWindow := ui.NewMainWindow(cfg, app)
	app.mainWindow = mainWindow

	return app, nil
}

// Run starts the application
func (a *App) Run() {
	// Show the main window
	a.mainWindow.Show()

	// Auto-detect location if enabled
	if a.config.Settings.AutoDetectLocation {
		a.DetectLocation()
	} else {
		// Calculate sun times for current location
		a.recalculate()
	}
}

// DetectLocation attempts to detect the user's location via IP
func (a *App) DetectLocation() {
	go func() {
		location, err := a.geoService.DetectLocation()
		mainthread.Wait(func() {
			if err != nil {
				a.mainWindow.ShowError(fmt.Sprintf("Failed to detect location: %v", err))
				// Use default location
				a.UpdateLocation(domain.DefaultLocation())
				return
			}
			a.UpdateLocation(location)
		})
	}()
}

// UpdateLocation updates the current location and recalculates sun times
func (a *App) UpdateLocation(loc domain.Location) {
	a.location = loc
	a.mainWindow.UpdateLocation(loc)
	a.recalculate()

	// Save as last location
	a.config.Settings.LastLocation = &loc
	a.saveSettings()
}

// UpdateDate updates the current date and recalculates sun times
func (a *App) UpdateDate(date time.Time) {
	a.currentDate = date
	a.mainWindow.UpdateDate(date)
	a.recalculate()
}

// UpdateSettings updates the application settings
func (a *App) UpdateSettings(settings domain.Settings) {
	a.config.Settings = settings
	a.solarCalc.UpdateSettings(settings)
	a.saveSettings()
	a.recalculate()
}

// SearchLocation searches for a location by name
func (a *App) SearchLocation(query string) {
	go func() {
		locations, err := a.geocoding.Search(query, 5)
		mainthread.Wait(func() {
			if err != nil {
				a.mainWindow.ShowError(fmt.Sprintf("Search failed: %v", err))
				return
			}
			if len(locations) == 0 {
				a.mainWindow.ShowError("No locations found")
				return
			}
			// Use the first result
			a.UpdateLocation(locations[0])
		})
	}()
}

// OnMapClick handles map click events
func (a *App) OnMapClick(lat, lon float64) {
	// Reverse geocode to get location name
	go func() {
		name, _ := a.geocoding.ReverseGeocode(lat, lon)
		mainthread.Wait(func() {
			loc := domain.Location{
				Latitude:  lat,
				Longitude: lon,
				Name:      name,
				Timezone:  a.location.Timezone, // Keep current timezone
			}
			if name == "" {
				loc.Name = fmt.Sprintf("%.4f, %.4f", lat, lon)
			}
			a.UpdateLocation(loc)
		})
	}()
}

// GetSettings returns the current settings
func (a *App) GetSettings() domain.Settings {
	return a.config.Settings
}

// GetLocation returns the current location
func (a *App) GetLocation() domain.Location {
	return a.location
}

// GetDate returns the current date
func (a *App) GetDate() time.Time {
	return a.currentDate
}

// recalculate recalculates sun times and updates the UI
func (a *App) recalculate() {
	if a.mainWindow == nil {
		return // UI not ready yet
	}
	sunTimes, err := a.solarCalc.Calculate(a.location, a.currentDate)
	if err != nil {
		a.mainWindow.ShowError(fmt.Sprintf("Calculation error: %v", err))
		return
	}
	a.mainWindow.UpdateSunTimes(sunTimes)
}

// saveSettings saves the current settings to disk
func (a *App) saveSettings() {
	if err := a.prefs.Save(a.config.Settings); err != nil && a.mainWindow != nil {
		a.mainWindow.ShowError(fmt.Sprintf("Failed to save settings: %v", err))
	}
}
