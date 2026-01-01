# GoGoldenHour

A cross-platform desktop application for photographers to calculate Golden Hour and Blue Hour times. Built with pure Go using miqt Qt6 bindings and featuring an interactive OpenStreetMap.

## Features

- **Golden Hour Calculation**: Displays morning and evening golden hour times based on sun elevation
- **Blue Hour Calculation**: Shows twilight periods for that magical blue light
- **Interactive Map**: OpenStreetMap with Leaflet.js for location selection
- **IP Geolocation**: Auto-detect your location on startup
- **Location Search**: Search for any location using OpenStreetMap Nominatim
- **Date Navigation**: View sun times for any date with easy navigation
- **Customizable Settings**:
  - Adjustable elevation angles for golden/blue hour definitions
  - 12-hour or 24-hour time format
  - Auto-detect location on startup toggle
- **Persistent Preferences**: Settings and last location saved between sessions

## Screenshots

```
+-----------------------------------------------------------+
| GoGoldenHour - Golden & Blue Hour Calculator              |
+----------------------------+------------------------------+
|                            | LOCATION                     |
|                            | [Search...           ] [Go]  |
|     INTERACTIVE MAP        | Lat: 51.5074  Lon: -0.1278   |
|    (OpenStreetMap)         | London, United Kingdom       |
|                            +------------------------------+
|        [Marker]            | DATE                         |
|                            | [<] January 1, 2026 [>] [Today]
|                            +------------------------------+
|                            | SUN TIMES                    |
|                            | Sunrise: 08:06  Sunset: 16:02|
|                            | Golden Hour   | Blue Hour    |
|                            | AM: 08:06-08:53| AM: 07:33-08:06
|                            | PM: 15:15-16:02| PM: 16:02-16:35
+----------------------------+------------------------------+
| [x] Settings                                              |
| Golden: [6°] Blue Start: [-4°] | Blue End: [-8°] [x] 24h  |
| [x] Auto-detect location on startup                       |
+-----------------------------------------------------------+
```

## Requirements

- Go 1.22 or later
- Qt 6.x with WebEngine support
- Linux, Windows, or macOS

### Linux Dependencies (Debian/Ubuntu)

```bash
sudo apt install qt6-base-dev qt6-webengine-dev
```

### Linux Dependencies (Arch)

```bash
sudo pacman -S qt6-base qt6-webengine
```

## Installation

```bash
# Clone the repository
git clone https://github.com/megatih/GoGoldenHour.git
cd GoGoldenHour

# Build the application
go build -o gogoldenhour ./cmd/gogoldenhour

# Run
./gogoldenhour
```

## Project Structure

```
GoGoldenHour/
├── cmd/gogoldenhour/
│   └── main.go                 # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go              # Application orchestration
│   ├── bridge/
│   │   └── jsbridge.go         # Go-JavaScript bridge (placeholder)
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── domain/
│   │   ├── location.go         # Location entity
│   │   ├── settings.go         # Settings entity
│   │   └── suntime.go          # Sun time calculations entity
│   ├── service/
│   │   ├── geocoding/
│   │   │   └── nominatim.go    # Location search via Nominatim
│   │   ├── geolocation/
│   │   │   └── ipapi.go        # IP-based location detection
│   │   └── solar/
│   │       └── calculator.go   # Solar position calculations
│   ├── storage/
│   │   └── preferences.go      # Persistent storage
│   └── ui/
│       ├── mainwindow.go       # Main window layout
│       └── widgets/
│           ├── datepanel.go    # Date selection widget
│           ├── locationpanel.go # Location display/search
│           ├── mapview.go      # WebEngine map widget
│           ├── settingspanel.go # Settings configuration
│           └── timepanel.go    # Golden/Blue hour display
├── web/
│   ├── css/
│   │   └── map.css             # Map styling
│   ├── js/
│   │   ├── bridge.js           # WebChannel bridge
│   │   └── map.js              # Map logic
│   └── map.html                # Leaflet.js map
├── go.mod
├── go.sum
├── README.md
├── LICENSE.md
└── CHANGELOG.md
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [github.com/mappu/miqt](https://github.com/mappu/miqt) | Qt6 bindings for Go |
| [github.com/hablullah/go-sampa](https://github.com/hablullah/go-sampa) | Solar position algorithm (supports custom elevation angles) |

## External APIs

| API | Purpose | Rate Limit |
|-----|---------|------------|
| [ip-api.com](http://ip-api.com) | IP geolocation | 45 req/min |
| [Nominatim](https://nominatim.openstreetmap.org) | Geocoding/search | 1 req/sec |

## Configuration

Settings are stored in `~/.config/gogoldenhour/settings.json`:

```json
{
  "golden_hour_elevation": 6,
  "blue_hour_start": -4,
  "blue_hour_end": -8,
  "time_format_24_hour": true,
  "auto_detect_location": true,
  "last_location": {
    "latitude": 51.5074,
    "longitude": -0.1278,
    "name": "London, United Kingdom"
  }
}
```

### Default Settings

| Setting | Default | Range | Description |
|---------|---------|-------|-------------|
| Golden Hour Elevation | 6° | 0° to 15° | Sun angle above horizon |
| Blue Hour Start | -4° | 0° to -6° | Civil twilight begins |
| Blue Hour End | -8° | -6° to -18° | Nautical twilight |
| Time Format | 24-hour | 12h/24h | Display format |
| Auto-detect Location | Yes | Yes/No | Detect on startup |

## Technical Notes

### Qt6 Bindings (miqt)

This project uses [miqt](https://github.com/mappu/miqt) v0.12.0 for Qt6 bindings. Key API patterns:

```go
// Widget constructors with text use suffix "3"
qt.NewQGroupBox3("Title")
qt.NewQLabel3("Text")
qt.NewQPushButton3("Button")

// Constructors without arguments use suffix "2"
qt.NewQLineEdit2()
qt.NewQDateEdit2()

// Date handling
currentDate := qt.QDate_CurrentDate()  // Returns *QDate
dateEdit.SetDate(*currentDate)         // Dereference needed
```

### Map Implementation

Due to Qt Location not being available in miqt, the map uses:
- Qt WebEngine for the browser component
- Leaflet.js for interactive mapping
- OpenStreetMap tiles

**Communication**: Map clicks are communicated to Go via console message interception (`OnJavaScriptConsoleMessage`). Location updates from Go to JavaScript use URL hash fragment changes, enabling smooth map panning without full page reloads.

### Solar Calculations

Uses go-sampa with custom elevation events:

```go
customEvents := []sampa.CustomSunEvent{
    {Name: "GoldenStart", BeforeTransit: true,
     Elevation: func(_ sampa.SunPosition) float64 { return 6.0 }},
    {Name: "BlueEnd", BeforeTransit: true,
     Elevation: func(_ sampa.SunPosition) float64 { return -8.0 }},
}
```

## Troubleshooting

### GPU Rendering Issues on ARM Devices

If you see errors like `MESA: error: drmPrimeHandleToFD() failed` or `Backend texture is not a Vulkan texture`, the application automatically disables GPU acceleration for Qt WebEngine. This is handled internally via the `QTWEBENGINE_CHROMIUM_FLAGS` environment variable.

## License

MIT License - see [LICENSE.md](LICENSE.md)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [miqt](https://github.com/mappu/miqt) - Qt6 Go bindings
- [go-sampa](https://github.com/hablullah/go-sampa) - Solar position algorithm
- [Leaflet.js](https://leafletjs.com/) - Interactive maps
- [OpenStreetMap](https://www.openstreetmap.org/) - Map tiles
- [ip-api.com](http://ip-api.com/) - IP geolocation
- [Nominatim](https://nominatim.org/) - Geocoding service
