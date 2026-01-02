// Package widgets provides the individual UI components for GoGoldenHour.
//
// Each widget in this package is a self-contained UI component that:
//   - Manages its own Qt widgets and layout
//   - Communicates with the main window via callbacks
//   - Has no direct dependencies on other widgets
//
// Available widgets:
//   - MapView: Interactive Leaflet map in Qt WebEngine
//   - LocationPanel: Location search and display
//   - DatePanel: Date navigation with calendar
//   - TimePanel: Golden/blue hour time display
//   - SettingsPanel: User preferences configuration
//
// # miqt Qt6 API Patterns
//
// All widgets use miqt for Qt6 bindings. Common patterns:
//
// Constructor suffixes:
//   - NewQWidget2(): No parameters (suffix "2")
//   - NewQLabel3("text"): With text parameter (suffix "3")
//   - NewQGroupBox3("title"): With title parameter (suffix "3")
//
// Layout methods:
//   - layout.AddWidget(widget.QWidget): Takes single QWidget argument
//   - layout.AddLayout(sublayout.QLayout): Takes single QLayout argument
//
// Grid layout:
//   - AddWidget2(widget, row, col): Single cell
//   - AddWidget3(widget, row, col, rowSpan, colSpan): Spanning cells
//
// Date handling:
//   - QDate methods return *QDate (pointer)
//   - Must dereference when calling SetDate: dateEdit.SetDate(*qdate)
package widgets

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	we "github.com/mappu/miqt/qt6/webengine"
)

// =============================================================================
// MapView
// =============================================================================

// MapView wraps a QWebEngineView displaying an interactive Leaflet map.
//
// The map is implemented using Leaflet.js (a JavaScript mapping library)
// embedded in Qt WebEngine. This approach was chosen because:
//   - Leaflet provides smooth, interactive maps with tile caching
//   - OpenStreetMap tiles are free and don't require an API key
//   - Qt WebEngine handles all the rendering and JavaScript execution
//
// # Communication with JavaScript
//
// Since miqt doesn't expose QWebEnginePage.RunJavaScript(), the widget uses
// alternative communication methods:
//
// Go → JavaScript (location updates):
//   - URL hash fragments: data:text/html;base64,...#lat,lon,zoom
//   - JavaScript listens for 'hashchange' events
//   - Smooth panning without page reload
//
// JavaScript → Go (map clicks):
//   - JavaScript calls console.log("MAPCLICK:lat,lon")
//   - Qt OnJavaScriptConsoleMessage intercepts the message
//   - Go parses coordinates and invokes the callback
//
// # Embedded HTML
//
// The complete map HTML (including Leaflet library references) is embedded
// as a base64-encoded data URL. This avoids the need for external HTML files
// and ensures the map works immediately on load.
type MapView struct {
	// view is the Qt WebEngine view that displays the map.
	view *we.QWebEngineView

	// page is the web page associated with the view.
	// Used to intercept JavaScript console messages for click handling.
	page *we.QWebEnginePage

	// onMapClick is the callback invoked when the user clicks on the map.
	// The callback receives the latitude and longitude of the clicked point.
	onMapClick func(lat, lon float64)

	// ready indicates whether the map has finished loading.
	// Set to true when the OnLoadFinished signal fires with ok=true.
	ready bool

	// currentLat and currentLon track the current map center.
	// Used when building URLs for location updates.
	currentLat float64
	currentLon float64

	// baseURL is the data URL containing the map HTML.
	// Location updates append a hash fragment: baseURL#lat,lon,zoom
	baseURL string
}

// defaultZoom is the initial and default zoom level for the map.
// Zoom level 13 shows approximately city-level detail (a few kilometers).
const defaultZoom = 13

// NewMapView creates a new map view widget with the given click handler.
//
// Parameters:
//   - onMapClick: Callback invoked when user clicks on the map (lat, lon)
//
// Returns a fully initialized MapView ready to be added to a layout.
// The map initially shows London (51.5074, -0.1278) until SetLocation is called.
func NewMapView(onMapClick func(lat, lon float64)) *MapView {
	mv := &MapView{
		// NewQWebEngineView2(): No-param constructor (suffix "2")
		view:       we.NewQWebEngineView2(),
		onMapClick: onMapClick,
		currentLat: 51.5074, // Default: London
		currentLon: -0.1278,
	}

	mv.setupView()
	return mv
}

// buildLocationURL constructs a URL with location coordinates in the hash fragment.
//
// This is the key mechanism for updating the map location without a full page
// reload. The JavaScript in the map HTML listens for 'hashchange' events and
// updates the map view accordingly.
//
// URL format: data:text/html;base64,...#latitude,longitude,zoom
//
// Parameters:
//   - lat: Latitude of the map center
//   - lon: Longitude of the map center
//   - zoom: Zoom level (typically 13)
//
// Returns the complete URL with hash fragment.
func (mv *MapView) buildLocationURL(lat, lon float64, zoom int) string {
	return fmt.Sprintf("%s#%f,%f,%d", mv.baseURL, lat, lon, zoom)
}

// setupView initializes the web engine view
func (mv *MapView) setupView() {
	// Set minimum size for the map
	mv.view.SetMinimumSize2(400, 400)

	// Create a custom page directly (required for overriding virtual methods)
	mv.page = we.NewQWebEnginePage()
	mv.view.SetPage(mv.page)

	// Intercept console messages for map click events
	mv.page.OnJavaScriptConsoleMessage(func(super func(level we.QWebEnginePage__JavaScriptConsoleMessageLevel, message string, lineNumber int, sourceID string), level we.QWebEnginePage__JavaScriptConsoleMessageLevel, message string, lineNumber int, sourceID string) {
		// Check for map click message
		if strings.HasPrefix(message, "MAPCLICK:") {
			parts := strings.Split(strings.TrimPrefix(message, "MAPCLICK:"), ",")
			if len(parts) == 2 {
				lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				lon, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err1 == nil && err2 == nil && mv.onMapClick != nil {
					mv.onMapClick(lat, lon)
				}
			}
		}
		// Call parent handler for other messages
		super(level, message, lineNumber, sourceID)
	})

	// Connect to load finished signal
	mv.view.OnLoadFinished(func(ok bool) {
		mv.ready = ok
	})

	// Load the map HTML
	mv.loadMapHTML()
}

// loadMapHTML loads the map HTML content using data URL
func (mv *MapView) loadMapHTML() {
	html := mv.createMapHTML()
	encoded := base64.StdEncoding.EncodeToString([]byte(html))
	mv.baseURL = "data:text/html;base64," + encoded

	// Load with initial coordinates in hash
	mv.page.SetUrl(qt.NewQUrl3(mv.buildLocationURL(mv.currentLat, mv.currentLon, defaultZoom)))
}

// createMapHTML creates the complete HTML for the map
func (mv *MapView) createMapHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoGoldenHour Map</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        html, body { height: 100%; margin: 0; padding: 0; }
        #map { height: 100%; width: 100%; }
        .golden-marker {
            background: linear-gradient(135deg, #ff9800 0%, #ff5722 100%);
            border: 3px solid #fff;
            border-radius: 50%;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
            width: 20px;
            height: 20px;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
        // Parse initial coordinates from URL hash
        function parseHash() {
            var hash = window.location.hash.substring(1);
            if (hash) {
                var parts = hash.split(',');
                if (parts.length >= 2) {
                    var lat = parseFloat(parts[0]);
                    var lon = parseFloat(parts[1]);
                    var zoom = parts.length >= 3 ? parseInt(parts[2]) : 13;
                    if (!isNaN(lat) && !isNaN(lon)) {
                        return { lat: lat, lon: lon, zoom: zoom };
                    }
                }
            }
            return { lat: 51.5074, lon: -0.1278, zoom: 13 }; // Default: London
        }

        // Get initial position from hash
        var initial = parseHash();

        // Initialize map
        var map = L.map('map').setView([initial.lat, initial.lon], initial.zoom);

        // Add OpenStreetMap tiles
        L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        }).addTo(map);

        // Custom icon for the marker
        var goldenIcon = L.divIcon({
            className: 'golden-marker',
            iconSize: [20, 20],
            iconAnchor: [10, 10]
        });

        // Add initial marker
        var currentMarker = L.marker([initial.lat, initial.lon], {icon: goldenIcon}).addTo(map);

        // Update marker and center map
        function setLocation(lat, lon, zoom) {
            currentMarker.setLatLng([lat, lon]);
            map.setView([lat, lon], zoom || map.getZoom());
        }

        // Handle hash changes (location updates from Go)
        window.addEventListener('hashchange', function() {
            var pos = parseHash();
            setLocation(pos.lat, pos.lon, pos.zoom);
        });

        // Handle map clicks - notify Go via console message
        map.on('click', function(e) {
            var lat = e.latlng.lat;
            var lon = e.latlng.lng;
            currentMarker.setLatLng([lat, lon]);
            // Send click event to Go via console message
            console.log('MAPCLICK:' + lat + ',' + lon);
        });
    </script>
</body>
</html>`
}

// Widget returns the underlying QWidget
func (mv *MapView) Widget() *qt.QWidget {
	return mv.view.QWidget
}

// SetLocation updates the map location using hash fragment (no page reload)
func (mv *MapView) SetLocation(lat, lon float64) {
	mv.currentLat = lat
	mv.currentLon = lon

	// Update via hash change to avoid full page reload
	mv.page.SetUrl(qt.NewQUrl3(mv.buildLocationURL(lat, lon, defaultZoom)))
}

// CenterMap centers the map on the given coordinates
func (mv *MapView) CenterMap(lat, lon float64, zoom int) {
	mv.currentLat = lat
	mv.currentLon = lon

	mv.page.SetUrl(qt.NewQUrl3(mv.buildLocationURL(lat, lon, zoom)))
}

// IsReady returns true if the map is loaded and ready
func (mv *MapView) IsReady() bool {
	return mv.ready
}
