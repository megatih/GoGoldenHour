# GoGoldenHour

A cross-platform desktop application for photographers to calculate Golden Hour and Blue Hour times. Built with pure Go using miqt Qt6 bindings and featuring an interactive OpenStreetMap.

**Current Version**: 0.1.3 | **License**: AGPL v3 | **Status**: Stable

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

- Go 1.24 or later
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
│   └── main.go                 # Application entry point with GPU fix
├── internal/
│   ├── app/
│   │   └── app.go              # Application controller (orchestrates all components)
│   ├── config/
│   │   └── config.go           # Configuration and shared constants (HTTP timeout)
│   ├── domain/
│   │   ├── location.go         # Location entity with validation
│   │   ├── settings.go         # Settings entity with elevation angle diagram
│   │   └── suntime.go          # Sun times and TimeRange entities
│   ├── service/
│   │   ├── geocoding/
│   │   │   └── nominatim.go    # OpenStreetMap Nominatim API client
│   │   ├── geolocation/
│   │   │   └── ipapi.go        # IP-API geolocation service
│   │   ├── solar/
│   │   │   └── calculator.go   # go-sampa solar calculations (8 custom events)
│   │   └── timezone/
│   │       └── lookup.go       # Offline timezone lookup via tzf
│   ├── storage/
│   │   └── preferences.go      # JSON settings persistence
│   └── ui/
│       ├── mainwindow.go       # Main window with splitter layout
│       └── widgets/
│           ├── datepanel.go    # Date navigation with calendar popup
│           ├── locationpanel.go # Location search and display
│           ├── mapview.go      # Embedded Leaflet.js map (data URL)
│           ├── settingspanel.go # Collapsible 2-column settings grid
│           └── timepanel.go    # Golden/Blue hour time display
├── Makefile                    # Build automation (build, run, test, vet)
├── go.mod
├── go.sum
├── README.md                   # This file
├── CLAUDE.md                   # Claude Code AI assistant instructions
├── LICENSE.md                  # AGPL v3 license
└── CHANGELOG.md                # Version history
```

> **Note**: The `web/` directory contains legacy files from an earlier implementation. The current version embeds the Leaflet.js map HTML directly in `mapview.go` using base64 data URLs for simpler deployment.

## Dependencies

| Package | Purpose |
|---------|---------|
| [github.com/mappu/miqt](https://github.com/mappu/miqt) | Qt6 bindings for Go |
| [github.com/hablullah/go-sampa](https://github.com/hablullah/go-sampa) | Solar position algorithm (supports custom elevation angles) |
| [github.com/ringsaturn/tzf](https://github.com/ringsaturn/tzf) | Timezone lookup from geographic coordinates |

## External APIs

| API | Purpose | Rate Limit |
|-----|---------|------------|
| [ip-api.com](http://ip-api.com) | IP geolocation | 45 req/min |
| [Nominatim](https://nominatim.openstreetmap.org) | Geocoding/search | 1 req/sec |

## Configuration

Settings are stored in `~/.config/GoGoldenHour/settings.json`:

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

**Communication**:
- **Go → JavaScript**: Location updates use URL hash fragment changes (`#lat,lon,zoom`), enabling smooth map panning without full page reloads
- **JavaScript → Go**: Map clicks are communicated via console message interception (`OnJavaScriptConsoleMessage`) with the format `MAPCLICK:lat,lon`

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

## Code Documentation

The codebase is extensively documented with comprehensive comments following Go documentation standards:

### Documentation Features

- **Package-level documentation**: Each package has a header explaining its purpose and role in the architecture
- **Type documentation**: All exported types have detailed descriptions of their fields and responsibilities
- **Method documentation**: Go doc style comments for all exported functions and methods
- **Inline comments**: Complex logic, algorithms, and non-obvious implementation details are explained
- **ASCII diagrams**: UI layouts and architectural relationships are visualized with text diagrams
- **miqt API patterns**: Qt binding quirks (constructor suffixes, pointer handling) are documented extensively

### Key Documentation Locations

| File                                   | Documentation Highlights                                            |
|----------------------------------------|---------------------------------------------------------------------|
| `internal/app/app.go`                  | Architecture diagram, thread safety patterns, initialization order  |
| `internal/domain/settings.go`          | Sun elevation angle diagram explaining golden/blue hour boundaries  |
| `internal/service/solar/calculator.go` | 8 custom sun events for precise golden/blue hour calculation        |
| `internal/ui/widgets/mapview.go`       | JavaScript↔Go communication workarounds (no RunJavaScript)          |
| `internal/ui/widgets/settingspanel.go` | Initialization callback warning, grid layout patterns               |

## License

AGPL v3 License - see [LICENSE.md](LICENSE.md)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [miqt](https://github.com/mappu/miqt) - Qt6 Go bindings
- [go-sampa](https://github.com/hablullah/go-sampa) - Solar position algorithm
- [Leaflet.js](https://leafletjs.com/) - Interactive maps
- [OpenStreetMap](https://www.openstreetmap.org/) - Map tiles
- [ip-api.com](http://ip-api.com/) - IP geolocation
- [Nominatim](https://nominatim.org/) - Geocoding service
