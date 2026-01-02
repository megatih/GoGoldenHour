package widgets

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	we "github.com/mappu/miqt/qt6/webengine"
)

// MapView wraps a QWebEngineView displaying a Leaflet map
type MapView struct {
	view       *we.QWebEngineView
	page       *we.QWebEnginePage
	onMapClick func(lat, lon float64)
	ready      bool
	currentLat float64
	currentLon float64
	baseURL    string // Store base URL for hash-based updates
}

// Default zoom level for the map
const defaultZoom = 13

// NewMapView creates a new map view widget
func NewMapView(onMapClick func(lat, lon float64)) *MapView {
	mv := &MapView{
		view:       we.NewQWebEngineView2(),
		onMapClick: onMapClick,
		currentLat: 51.5074, // Default: London
		currentLon: -0.1278,
	}

	mv.setupView()
	return mv
}

// buildLocationURL constructs a URL with location coordinates in the hash fragment
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
