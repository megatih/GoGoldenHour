package widgets

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// LocationPanel
// =============================================================================

// LocationPanel provides location search and display functionality.
//
// This panel allows users to:
//   - Search for locations by name using Nominatim geocoding
//   - Auto-detect their location via IP geolocation
//   - View the current location's coordinates and name
//
// # UI Layout
//
//	┌─ Location ─────────────────────────┐
//	│ [Search location...        ] [Go]  │  <- Search input + button
//	│ [    Detect My Location        ]   │  <- Auto-detect button
//	│ Lat: 48.8566        Lon: 2.3522    │  <- Coordinate display
//	│ Paris, France                      │  <- Location name (orange, bold)
//	└────────────────────────────────────┘
//
// # Communication
//
// The panel communicates with the main application via callbacks:
//   - onSearch: Called when user submits a search query (Enter or Go button)
//   - onDetect: Called when user clicks "Detect My Location"
//
// These callbacks are invoked synchronously on the main Qt thread.
// The actual geocoding/geolocation work is done asynchronously by the App.
type LocationPanel struct {
	// groupBox is the container widget with "Location" title border.
	groupBox *qt.QGroupBox

	// searchInput is the text field for entering location search queries.
	// Supports Enter key to trigger search.
	searchInput *qt.QLineEdit

	// searchBtn triggers the search when clicked ("Go" button).
	searchBtn *qt.QPushButton

	// detectBtn triggers IP-based location detection.
	detectBtn *qt.QPushButton

	// latLabel displays the current latitude (e.g., "Lat: 48.8566").
	latLabel *qt.QLabel

	// lonLabel displays the current longitude (e.g., "Lon: 2.3522").
	lonLabel *qt.QLabel

	// nameLabel displays the human-readable location name.
	// Styled with orange color and bold font for visibility.
	nameLabel *qt.QLabel

	// onSearch is the callback invoked when user searches for a location.
	// Receives the search query string.
	onSearch func(query string)

	// onDetect is the callback invoked when user clicks auto-detect.
	onDetect func()
}

// NewLocationPanel creates a new location panel with the given callbacks.
//
// Parameters:
//   - onSearch: Callback invoked when user submits a search query.
//     The App uses this to trigger Nominatim geocoding.
//   - onDetect: Callback invoked when user clicks "Detect My Location".
//     The App uses this to trigger IP-based geolocation.
//
// Returns a fully initialized LocationPanel ready to be added to a layout.
// The panel initially shows placeholder text ("--") until SetLocation is called.
func NewLocationPanel(onSearch func(query string), onDetect func()) *LocationPanel {
	lp := &LocationPanel{
		onSearch: onSearch,
		onDetect: onDetect,
	}

	lp.setupUI()
	return lp
}

// performSearch validates and executes the location search.
//
// This is a consolidated helper that handles both Enter key and button click.
// It checks that:
//  1. The search input is not empty
//  2. The onSearch callback is set
//
// If both conditions are met, it invokes the callback with the search query.
func (lp *LocationPanel) performSearch() {
	query := lp.searchInput.Text()
	if query != "" && lp.onSearch != nil {
		lp.onSearch(query)
	}
}

// setupUI creates and arranges all widgets in the location panel.
//
// The layout is a vertical stack:
//  1. Search row: text input + "Go" button (horizontal)
//  2. Detect button: full-width "Detect My Location" button
//  3. Coordinates row: latitude and longitude labels (horizontal)
//  4. Name label: location name with special styling
//
// # miqt API Notes
//
// Constructor patterns used:
//   - NewQGroupBox3("title"): Creates group box with title (suffix "3")
//   - NewQLineEdit2(): Creates empty line edit (suffix "2" = no params)
//   - NewQPushButton3("text"): Creates button with text (suffix "3")
//   - NewQLabel3("text"): Creates label with text (suffix "3")
//   - NewQHBoxLayout2(): Creates horizontal layout (suffix "2" = no parent)
//
// Layout methods take single QWidget/QLayout argument (no stretch parameter).
func (lp *LocationPanel) setupUI() {
	// Create the group box container with "Location" title
	// NewQGroupBox3: suffix "3" = constructor with title parameter
	lp.groupBox = qt.NewQGroupBox3("Location")
	layout := qt.NewQVBoxLayout(lp.groupBox.QWidget)
	layout.SetSpacing(6)

	// =========================================================================
	// Search Row: Input field + Go button
	// =========================================================================
	// NewQHBoxLayout2: suffix "2" = no-parent constructor
	searchRow := qt.NewQHBoxLayout2()

	// NewQLineEdit2: suffix "2" = no-parameter constructor (empty input)
	lp.searchInput = qt.NewQLineEdit2()
	lp.searchInput.SetPlaceholderText("Search location...")

	// NewQPushButton3: suffix "3" = constructor with text parameter
	lp.searchBtn = qt.NewQPushButton3("Go")
	lp.searchBtn.SetFixedWidth(50)

	// Connect both button click and Enter key to the same search handler.
	// This provides a consistent UX - users can click or press Enter.
	lp.searchBtn.OnClicked(func() { lp.performSearch() })
	lp.searchInput.OnReturnPressed(func() { lp.performSearch() })

	// Add widgets to horizontal layout (miqt takes single argument, no stretch)
	searchRow.AddWidget(lp.searchInput.QWidget)
	searchRow.AddWidget(lp.searchBtn.QWidget)
	layout.AddLayout(searchRow.QLayout)

	// =========================================================================
	// Detect Location Button
	// =========================================================================
	// Full-width button for IP-based location detection
	lp.detectBtn = qt.NewQPushButton3("Detect My Location")
	lp.detectBtn.OnClicked(func() {
		if lp.onDetect != nil {
			lp.onDetect()
		}
	})
	layout.AddWidget(lp.detectBtn.QWidget)

	// =========================================================================
	// Coordinates Display Row
	// =========================================================================
	// Show latitude and longitude side by side
	coordsLayout := qt.NewQHBoxLayout2()
	lp.latLabel = qt.NewQLabel3("Lat: --")
	lp.lonLabel = qt.NewQLabel3("Lon: --")
	coordsLayout.AddWidget(lp.latLabel.QWidget)
	coordsLayout.AddWidget(lp.lonLabel.QWidget)
	layout.AddLayout(coordsLayout.QLayout)

	// =========================================================================
	// Location Name Display
	// =========================================================================
	// Display location name with golden hour theme styling (orange color)
	lp.nameLabel = qt.NewQLabel3("--")
	lp.nameLabel.SetWordWrap(true) // Handle long location names
	lp.nameLabel.SetStyleSheet("font-weight: bold; color: #ff9800;")
	layout.AddWidget(lp.nameLabel.QWidget)
}

// Widget returns the group box container for adding to parent layouts.
//
// The returned QGroupBox contains all location panel widgets and can be
// added to a parent layout using layout.AddWidget(panel.Widget().QWidget).
func (lp *LocationPanel) Widget() *qt.QGroupBox {
	return lp.groupBox
}

// SetLocation updates the displayed location information.
//
// This method is called by MainWindow when the location changes, either from:
//   - User search (geocoding result)
//   - Auto-detect (IP geolocation result)
//   - Map click (reverse geocoding result)
//
// The display is updated with:
//   - Latitude formatted to 4 decimal places (≈11m precision)
//   - Longitude formatted to 4 decimal places
//   - Location name (city, country, or coordinates if unavailable)
func (lp *LocationPanel) SetLocation(loc domain.Location) {
	lp.latLabel.SetText(fmt.Sprintf("Lat: %.4f", loc.Latitude))
	lp.lonLabel.SetText(fmt.Sprintf("Lon: %.4f", loc.Longitude))
	lp.nameLabel.SetText(loc.Name)
}
