// Package ui provides the graphical user interface for GoGoldenHour.
//
// This package contains the MainWindow (the primary UI controller) and
// coordinates all widget interactions. It acts as a bridge between the
// user interface and the application controller (app.App).
//
// # Architecture
//
// The UI follows a hierarchical structure:
//
//	MainWindow
//	├── MapView (Qt WebEngine + Leaflet.js)
//	├── LocationPanel (search, detect, display)
//	├── DatePanel (navigation, calendar)
//	├── TimePanel (golden/blue hour display)
//	├── SettingsPanel (elevation angles, preferences)
//	└── StatusBar (messages, errors)
//
// # Communication Pattern
//
// Widgets communicate with the application via callbacks:
//
//  1. User interacts with widget (e.g., clicks search button)
//  2. Widget calls its callback function (e.g., onSearch)
//  3. MainWindow's event handler receives the call
//  4. MainWindow delegates to the AppController interface
//  5. App processes the action and may call MainWindow.Update*() methods
//  6. MainWindow updates the relevant widgets
//
// This pattern keeps widgets decoupled from the application logic.
//
// # miqt Qt6 Bindings
//
// The package uses miqt (github.com/mappu/miqt) for Qt6 bindings.
// Key miqt patterns used throughout:
//
//   - Constructor suffixes: NewQLabel3("text") for text param, NewQLineEdit2() for no params
//   - Layout methods: AddWidget takes single argument (no stretch parameter)
//   - QWidget access: Most widgets have .QWidget field for layout compatibility
//   - QDate pointers: Date methods return *QDate, must dereference when setting
//
// # Thread Safety
//
// All UI operations must happen on the main Qt thread. The App controller
// ensures this by using mainthread.Wait() for any operations that update
// the UI from background goroutines.
package ui

import (
	"fmt"
	"time"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/config"
	"github.com/megatih/GoGoldenHour/internal/domain"
	"github.com/megatih/GoGoldenHour/internal/ui/widgets"
)

// =============================================================================
// AppController Interface
// =============================================================================

// AppController defines the interface for application control.
//
// This interface allows the MainWindow to communicate user actions to the
// application controller without knowing the concrete implementation.
// It's implemented by app.App.
//
// The interface includes:
//   - Action methods: DetectLocation, SearchLocation, OnMapClick
//   - Update methods: UpdateLocation, UpdateDate, UpdateSettings
//   - Query methods: GetSettings, GetLocation, GetDate
//
// This interface enables:
//   - Loose coupling between UI and application logic
//   - Easier testing (can mock the controller)
//   - Clear contract for UI-to-app communication
type AppController interface {
	// DetectLocation initiates IP-based location detection.
	// Called when user clicks "Detect My Location" button.
	DetectLocation()

	// UpdateLocation changes the current location.
	// Called after search results or map clicks.
	UpdateLocation(loc domain.Location)

	// UpdateDate changes the date for calculations.
	// Called when user navigates dates or uses calendar.
	UpdateDate(date time.Time)

	// UpdateSettings applies new user preferences.
	// Called when settings panel values change.
	UpdateSettings(settings domain.Settings)

	// SearchLocation performs geocoding search.
	// Called when user submits a location query.
	SearchLocation(query string)

	// OnMapClick handles map click events.
	// Called when user clicks on the map.
	OnMapClick(lat, lon float64)

	// GetSettings returns current settings.
	// Used for initializing UI components.
	GetSettings() domain.Settings

	// GetLocation returns current location.
	// Used for initializing UI components.
	GetLocation() domain.Location

	// GetDate returns current calculation date.
	// Used for initializing UI components.
	GetDate() time.Time
}

// =============================================================================
// MainWindow
// =============================================================================

// MainWindow is the main application window that contains all UI components.
//
// The MainWindow is responsible for:
//   - Creating and arranging all widgets
//   - Routing user events to the AppController
//   - Updating widgets when data changes
//   - Managing the status bar
//
// Layout:
//
//	┌────────────────────────────────────────────────────────────────────┐
//	│                    GoGoldenHour - Golden & Blue Hour Calculator    │
//	├────────────────────────────────┬───────────────────────────────────┤
//	│                                │  ┌─────────────────────────────┐  │
//	│                                │  │  Location Panel             │  │
//	│                                │  │  [Search] [Detect]          │  │
//	│                                │  └─────────────────────────────┘  │
//	│        Interactive Map         │  ┌─────────────────────────────┐  │
//	│       (60% of width)           │  │  Date Panel                 │  │
//	│                                │  │  < [Date] > [Today]         │  │
//	│                                │  └─────────────────────────────┘  │
//	│                                │  ┌─────────────────────────────┐  │
//	│                                │  │  Time Panel                 │  │
//	│                                │  │  Golden Hour | Blue Hour    │  │
//	│                                │  └─────────────────────────────┘  │
//	│                                │                                   │
//	│                                │  ┌─────────────────────────────┐  │
//	│                                │  │  Settings (collapsible)     │  │
//	│                                │  └─────────────────────────────┘  │
//	├────────────────────────────────┴───────────────────────────────────┤
//	│  Status: Location name or error message                            │
//	└────────────────────────────────────────────────────────────────────┘
type MainWindow struct {
	// window is the top-level Qt main window.
	// Provides title bar, status bar, and central widget area.
	window *qt.QMainWindow

	// controller is the application controller that handles business logic.
	// All user actions are delegated to this controller.
	controller AppController

	// config holds the application configuration including settings.
	// Used for window size and initial widget values.
	config config.AppConfig

	// mapView displays the interactive Leaflet map.
	// Uses Qt WebEngine to embed the web-based map.
	mapView *widgets.MapView

	// locationPanel provides search and location display.
	// Contains search input, detect button, and coordinate display.
	locationPanel *widgets.LocationPanel

	// timePanel displays calculated sun times.
	// Shows golden hour and blue hour in side-by-side columns.
	timePanel *widgets.TimePanel

	// datePanel provides date navigation.
	// Contains prev/next buttons, date picker, and today button.
	datePanel *widgets.DatePanel

	// settingsPanel allows adjusting elevation angles and preferences.
	// Starts collapsed to save space; can be expanded by user.
	settingsPanel *widgets.SettingsPanel

	// statusLabel displays status messages and errors.
	// Located in the status bar at the bottom of the window.
	statusLabel *qt.QLabel
}

// =============================================================================
// Constructor
// =============================================================================

// NewMainWindow creates the main application window with all widgets.
//
// This constructor:
//  1. Creates the MainWindow struct
//  2. Calls setupUI() to create and arrange all widgets
//  3. Returns the fully initialized window (but not yet shown)
//
// Parameters:
//   - cfg: Application configuration with window size and initial settings
//   - controller: The AppController for handling user actions
//
// Returns the created MainWindow. Call Show() to make it visible.
//
// Note: The SettingsPanel may trigger OnValueChanged callbacks during
// construction when applySettings() is called. The App controller handles
// this by checking if mainWindow is nil before using it.
func NewMainWindow(cfg config.AppConfig, controller AppController) *MainWindow {
	mw := &MainWindow{
		config:     cfg,
		controller: controller,
	}

	// Create and arrange all UI components
	mw.setupUI()
	return mw
}

// =============================================================================
// UI Setup
// =============================================================================

// setupUI creates and arranges all UI components.
//
// This method builds the complete UI hierarchy:
//  1. Creates main window with title and size
//  2. Creates central widget with main layout
//  3. Creates horizontal splitter (map | info panels)
//  4. Creates all widgets with their callbacks
//  5. Sets up status bar
//
// Layout uses Qt's layout system:
//   - QSplitter: Divides window between map and info panels
//   - QVBoxLayout: Stacks info panels vertically
//   - Individual widgets handle their internal layout
//
// miqt API notes:
//   - NewQMainWindow(nil): No parent (top-level window)
//   - SetMinimumSize2(w, h): miqt suffix "2" for int overload
//   - NewQLabel3("text"): miqt suffix "3" for text parameter
//   - widget.QWidget: Access base QWidget for layout compatibility
func (mw *MainWindow) setupUI() {
	// =========================================================================
	// Main Window Setup
	// =========================================================================
	// Create top-level window with title and size constraints
	mw.window = qt.NewQMainWindow(nil)
	mw.window.SetWindowTitle("GoGoldenHour - Golden & Blue Hour Calculator")
	mw.window.Resize(mw.config.WindowWidth, mw.config.WindowHeight)
	// SetMinimumSize2 uses integer overload (suffix "2" in miqt)
	mw.window.SetMinimumSize2(800, 600)

	// =========================================================================
	// Central Widget and Main Layout
	// =========================================================================
	// Create central widget that holds all content
	centralWidget := qt.NewQWidget(nil)
	mainLayout := qt.NewQVBoxLayout(centralWidget)
	mainLayout.SetContentsMargins(10, 10, 10, 10)
	mainLayout.SetSpacing(10)

	// =========================================================================
	// Splitter (Map | Info Panels)
	// =========================================================================
	// Horizontal splitter allows user to resize between map and info panels
	splitter := qt.NewQSplitter(nil)
	splitter.SetOrientation(qt.Horizontal)

	// =========================================================================
	// Left Side: Interactive Map
	// =========================================================================
	// Create map view with click handler callback
	mw.mapView = widgets.NewMapView(mw.onMapClick)
	splitter.AddWidget(mw.mapView.Widget())

	// =========================================================================
	// Right Side: Info Panels
	// =========================================================================
	rightPanel := qt.NewQWidget(nil)
	rightLayout := qt.NewQVBoxLayout(rightPanel)
	rightLayout.SetContentsMargins(0, 0, 0, 0)
	rightLayout.SetSpacing(8)

	// Location panel: Search and location display
	// Callbacks: onLocationSearch (search button/enter), onDetectLocation (detect button)
	mw.locationPanel = widgets.NewLocationPanel(mw.onLocationSearch, mw.onDetectLocation)
	rightLayout.AddWidget(mw.locationPanel.Widget().QWidget)

	// Date panel: Date navigation with calendar
	// Callback: onDateChanged (any date change)
	mw.datePanel = widgets.NewDatePanel(mw.onDateChanged)
	rightLayout.AddWidget(mw.datePanel.Widget().QWidget)

	// Time panel: Golden and blue hour display in side-by-side columns
	// No callback - this is a display-only widget
	mw.timePanel = widgets.NewTimePanel(mw.config.Settings.TimeFormat24Hour)
	rightLayout.AddWidget(mw.timePanel.Widget().QWidget)

	// Add stretch to push settings panel to the bottom
	// This keeps the settings collapsed at the bottom of the panel
	rightLayout.AddStretch()

	// Settings panel: Elevation angles and preferences
	// Callback: onSettingsChanged (any setting change)
	// Note: This may trigger callback during construction (applySettings)
	mw.settingsPanel = widgets.NewSettingsPanel(mw.config.Settings, mw.onSettingsChanged)
	rightLayout.AddWidget(mw.settingsPanel.Widget().QWidget)

	splitter.AddWidget(rightPanel)

	// =========================================================================
	// Splitter Proportions
	// =========================================================================
	// Set initial sizes: 60% for map (480px), 40% for info (320px)
	// User can drag the splitter to adjust these proportions
	splitter.SetSizes([]int{480, 320})

	// Add splitter to main layout (use .QWidget for layout compatibility)
	mainLayout.AddWidget(splitter.QWidget)

	// =========================================================================
	// Status Bar
	// =========================================================================
	// Create status label for messages and errors
	// NewQLabel3("") creates label with empty initial text (suffix "3" = text param)
	mw.statusLabel = qt.NewQLabel3("")
	statusBar := mw.window.StatusBar()
	// AddPermanentWidget keeps the label visible (not replaced by temporary messages)
	statusBar.AddPermanentWidget(mw.statusLabel.QWidget)

	// Set central widget to complete window setup
	mw.window.SetCentralWidget(centralWidget)
}

// =============================================================================
// Window Control
// =============================================================================

// Show makes the main window visible on screen.
//
// This should be called after the window is fully constructed and
// the application is ready to display. Typically called from App.Run().
func (mw *MainWindow) Show() {
	mw.window.Show()
}

// =============================================================================
// Update Methods (called by App controller)
// =============================================================================

// UpdateLocation updates the location display in all relevant widgets.
//
// This is called by the App controller after a location change from:
//   - IP-based detection
//   - Search results
//   - Map clicks
//
// The method updates:
//   - LocationPanel: Shows coordinates and location name
//   - MapView: Centers and marks the new location
//   - StatusBar: Shows location name
//
// Nil checks protect against calls during initialization.
func (mw *MainWindow) UpdateLocation(loc domain.Location) {
	// Update location panel (coordinates and name display)
	if mw.locationPanel != nil {
		mw.locationPanel.SetLocation(loc)
	}

	// Update map view (center and marker)
	if mw.mapView != nil {
		mw.mapView.SetLocation(loc.Latitude, loc.Longitude)
	}

	// Update status bar with location name
	mw.setStatus(fmt.Sprintf("Location: %s", loc.Name))
}

// UpdateDate updates the date display in the date panel.
//
// This is called by the App controller after a date change from:
//   - Previous/next buttons
//   - Calendar selection
//   - Today button
//
// Nil check protects against calls during initialization.
func (mw *MainWindow) UpdateDate(date time.Time) {
	if mw.datePanel != nil {
		mw.datePanel.SetDate(date)
	}
}

// UpdateSunTimes updates the time panel with calculated sun times.
//
// This is called by the App controller after recalculation due to:
//   - Location change
//   - Date change
//   - Settings change (elevation angles)
//
// The time format (12/24 hour) is passed from current settings.
// Nil check protects against calls during initialization.
func (mw *MainWindow) UpdateSunTimes(sunTimes domain.SunTimes) {
	if mw.timePanel != nil {
		mw.timePanel.SetSunTimes(sunTimes, mw.config.Settings.TimeFormat24Hour)
	}
}

// ShowError displays an error message in the status bar.
//
// This is called by the App controller when operations fail:
//   - Location detection failure
//   - Search failure
//   - Calculation errors
//   - Settings save failures
//
// Errors are prefixed with "Error: " for visibility.
func (mw *MainWindow) ShowError(message string) {
	mw.setStatus(fmt.Sprintf("Error: %s", message))
}

// =============================================================================
// Internal Helpers
// =============================================================================

// setStatus updates the status bar text.
//
// This is an internal helper used by UpdateLocation and ShowError.
// Nil check protects against calls during initialization.
func (mw *MainWindow) setStatus(message string) {
	if mw.statusLabel != nil {
		mw.statusLabel.SetText(message)
	}
}

// =============================================================================
// Event Handlers (callbacks from widgets)
// =============================================================================

// onMapClick handles map click events from the MapView widget.
//
// This is passed to MapView as a callback during construction.
// When the user clicks on the map, MapView parses the coordinates
// from the JavaScript console message and calls this handler.
//
// The handler simply delegates to the AppController for processing.
func (mw *MainWindow) onMapClick(lat, lon float64) {
	mw.controller.OnMapClick(lat, lon)
}

// onLocationSearch handles search submissions from the LocationPanel widget.
//
// This is passed to LocationPanel as a callback during construction.
// When the user presses Enter in the search box or clicks the Go button,
// LocationPanel calls this handler with the query text.
//
// The handler delegates to the AppController for geocoding.
func (mw *MainWindow) onLocationSearch(query string) {
	mw.controller.SearchLocation(query)
}

// onDetectLocation handles the "Detect My Location" button from LocationPanel.
//
// This is passed to LocationPanel as a callback during construction.
// When the user clicks the detect button, LocationPanel calls this handler.
//
// The handler delegates to the AppController for IP-based detection.
func (mw *MainWindow) onDetectLocation() {
	mw.controller.DetectLocation()
}

// onDateChanged handles date changes from the DatePanel widget.
//
// This is passed to DatePanel as a callback during construction.
// When the user changes the date (via buttons, calendar, or today button),
// DatePanel calls this handler with the new date.
//
// The handler delegates to the AppController for recalculation.
func (mw *MainWindow) onDateChanged(date time.Time) {
	mw.controller.UpdateDate(date)
}

// onSettingsChanged handles settings changes from the SettingsPanel widget.
//
// This is passed to SettingsPanel as a callback during construction.
// When the user changes any setting (elevation angles, checkboxes),
// SettingsPanel calls this handler with the complete new settings.
//
// The handler:
//  1. Updates local config with new settings
//  2. Updates time panel format (in case 12/24 hour changed)
//  3. Delegates to AppController for persistence and recalculation
//
// Note: This may be called during SettingsPanel construction (applySettings).
// The App controller handles this by checking if mainWindow is nil.
func (mw *MainWindow) onSettingsChanged(settings domain.Settings) {
	// Update local config
	mw.config.Settings = settings

	// Update time format immediately (before waiting for recalculation)
	mw.timePanel.SetTimeFormat(settings.TimeFormat24Hour)

	// Delegate to controller for persistence and recalculation
	mw.controller.UpdateSettings(settings)
}
