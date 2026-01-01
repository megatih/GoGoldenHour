package ui

import (
	"fmt"
	"time"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
	"github.com/megatih/GoGoldenHour/internal/ui/widgets"
)

// AppController defines the interface for application control
type AppController interface {
	DetectLocation()
	UpdateLocation(loc domain.Location)
	UpdateDate(date time.Time)
	UpdateSettings(settings domain.Settings)
	SearchLocation(query string)
	OnMapClick(lat, lon float64)
	GetSettings() domain.Settings
	GetLocation() domain.Location
	GetDate() time.Time
}

// MainWindow is the main application window
type MainWindow struct {
	window        *qt.QMainWindow
	controller    AppController
	config        config.AppConfig
	mapView       *widgets.MapView
	locationPanel *widgets.LocationPanel
	timePanel     *widgets.TimePanel
	datePanel     *widgets.DatePanel
	settingsPanel *widgets.SettingsPanel
	statusLabel   *qt.QLabel
}

// NewMainWindow creates the main application window
func NewMainWindow(cfg config.AppConfig, controller AppController) *MainWindow {
	mw := &MainWindow{
		config:     cfg,
		controller: controller,
	}

	mw.setupUI()
	return mw
}

// setupUI creates and arranges all UI components
func (mw *MainWindow) setupUI() {
	// Create main window
	mw.window = qt.NewQMainWindow(nil)
	mw.window.SetWindowTitle("GoGoldenHour - Golden & Blue Hour Calculator")
	mw.window.Resize(mw.config.WindowWidth, mw.config.WindowHeight)
	mw.window.SetMinimumSize2(800, 600)

	// Create central widget
	centralWidget := qt.NewQWidget(nil)
	mainLayout := qt.NewQVBoxLayout(centralWidget)
	mainLayout.SetContentsMargins(10, 10, 10, 10)
	mainLayout.SetSpacing(10)

	// Create splitter for map and info panels
	splitter := qt.NewQSplitter(nil)
	splitter.SetOrientation(qt.Horizontal)

	// Left side: Map view
	mw.mapView = widgets.NewMapView(mw.onMapClick)
	splitter.AddWidget(mw.mapView.Widget())

	// Right side: Info panels
	rightPanel := qt.NewQWidget(nil)
	rightLayout := qt.NewQVBoxLayout(rightPanel)
	rightLayout.SetContentsMargins(0, 0, 0, 0)
	rightLayout.SetSpacing(8)

	// Location panel
	mw.locationPanel = widgets.NewLocationPanel(mw.onLocationSearch, mw.onDetectLocation)
	rightLayout.AddWidget(mw.locationPanel.Widget().QWidget)

	// Date panel
	mw.datePanel = widgets.NewDatePanel(mw.onDateChanged)
	rightLayout.AddWidget(mw.datePanel.Widget().QWidget)

	// Time panels (Golden Hour and Blue Hour)
	mw.timePanel = widgets.NewTimePanel(mw.config.Settings.TimeFormat24Hour)
	rightLayout.AddWidget(mw.timePanel.Widget().QWidget)

	// Add stretch to push settings to bottom
	rightLayout.AddStretch()

	// Settings panel
	mw.settingsPanel = widgets.NewSettingsPanel(mw.config.Settings, mw.onSettingsChanged)
	rightLayout.AddWidget(mw.settingsPanel.Widget().QWidget)

	splitter.AddWidget(rightPanel)

	// Set splitter sizes (60% map, 40% info)
	splitter.SetSizes([]int{480, 320})

	mainLayout.AddWidget(splitter.QWidget)

	// Status bar
	mw.statusLabel = qt.NewQLabel3("")
	statusBar := mw.window.StatusBar()
	statusBar.AddPermanentWidget(mw.statusLabel.QWidget)

	mw.window.SetCentralWidget(centralWidget)
}

// Show displays the main window
func (mw *MainWindow) Show() {
	mw.window.Show()
}

// UpdateLocation updates the location display
func (mw *MainWindow) UpdateLocation(loc domain.Location) {
	if mw.locationPanel != nil {
		mw.locationPanel.SetLocation(loc)
	}
	if mw.mapView != nil {
		mw.mapView.SetLocation(loc.Latitude, loc.Longitude)
	}
	mw.setStatus(fmt.Sprintf("Location: %s", loc.Name))
}

// UpdateDate updates the date display
func (mw *MainWindow) UpdateDate(date time.Time) {
	if mw.datePanel != nil {
		mw.datePanel.SetDate(date)
	}
}

// UpdateSunTimes updates the sun times display
func (mw *MainWindow) UpdateSunTimes(sunTimes domain.SunTimes) {
	if mw.timePanel != nil {
		mw.timePanel.SetSunTimes(sunTimes, mw.config.Settings.TimeFormat24Hour)
	}
}

// ShowError displays an error message
func (mw *MainWindow) ShowError(message string) {
	mw.setStatus(fmt.Sprintf("Error: %s", message))
}

// setStatus updates the status bar
func (mw *MainWindow) setStatus(message string) {
	if mw.statusLabel != nil {
		mw.statusLabel.SetText(message)
	}
}

// Event handlers

func (mw *MainWindow) onMapClick(lat, lon float64) {
	mw.controller.OnMapClick(lat, lon)
}

func (mw *MainWindow) onLocationSearch(query string) {
	mw.controller.SearchLocation(query)
}

func (mw *MainWindow) onDetectLocation() {
	mw.controller.DetectLocation()
}

func (mw *MainWindow) onDateChanged(date time.Time) {
	mw.controller.UpdateDate(date)
}

func (mw *MainWindow) onSettingsChanged(settings domain.Settings) {
	mw.config.Settings = settings
	mw.timePanel.SetTimeFormat(settings.TimeFormat24Hour)
	mw.controller.UpdateSettings(settings)
}
