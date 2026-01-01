package widgets

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
	we "github.com/mappu/miqt/qt6/webengine"
)

// MapView wraps a QWebEngineView displaying a Leaflet map
type MapView struct {
	view       *we.QWebEngineView
	onMapClick func(lat, lon float64)
	ready      bool
	currentLat float64
	currentLon float64
}

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

// setupView initializes the web engine view
func (mv *MapView) setupView() {
	// Set minimum size for the map
	mv.view.SetMinimumSize2(400, 400)

	// Get the page for configuration
	page := mv.view.Page()

	// Connect to load finished signal
	mv.view.OnLoadFinished(func(ok bool) {
		mv.ready = ok
		if ok {
			// Map is ready, we could set up initial state here
		}
	})

	// Load the map HTML
	mv.loadMapHTML(page)
}

// loadMapHTML loads the map HTML content
func (mv *MapView) loadMapHTML(page *we.QWebEnginePage) {
	// Create inline HTML with embedded Leaflet
	html := mv.createMapHTML()

	// Load HTML directly (resources are loaded from CDN)
	page.SetHtml(html)
}

// createMapHTML creates the complete HTML for the map
func (mv *MapView) createMapHTML() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoGoldenHour Map</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        html, body { height: 100%%; margin: 0; padding: 0; }
        #map { height: 100%%; width: 100%%; }
        .golden-marker {
            background: linear-gradient(135deg, #ff9800 0%%, #ff5722 100%%);
            border: 3px solid #fff;
            border-radius: 50%%;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
            width: 20px;
            height: 20px;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
        // Initialize map with marker
        var map = L.map('map').setView([%f, %f], 13);

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
        var currentMarker = L.marker([%f, %f], {icon: goldenIcon}).addTo(map);

        // Handle map clicks (click detection not available without WebChannel)
        map.on('click', function(e) {
            var lat = e.latlng.lat;
            var lon = e.latlng.lng;
            currentMarker.setLatLng([lat, lon]);
            // Note: Click notification to Go requires WebChannel which is not yet implemented
        });
    </script>
</body>
</html>`, mv.currentLat, mv.currentLon, mv.currentLat, mv.currentLon)
}

// Widget returns the underlying QWidget
func (mv *MapView) Widget() *qt.QWidget {
	return mv.view.QWidget
}

// SetLocation sets the map location and marker by reloading the map
func (mv *MapView) SetLocation(lat, lon float64) {
	mv.currentLat = lat
	mv.currentLon = lon
	// Reload the map with new coordinates
	mv.loadMapHTML(mv.view.Page())
}

// CenterMap centers the map on the given coordinates
func (mv *MapView) CenterMap(lat, lon float64, zoom int) {
	mv.SetLocation(lat, lon)
}

// IsReady returns true if the map is loaded and ready
func (mv *MapView) IsReady() bool {
	return mv.ready
}
