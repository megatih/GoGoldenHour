# Changelog

All notable changes to GoGoldenHour will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3] - 2026-01-02

### Added

- **Comprehensive Code Documentation**: Added extensive documentation to all 17 source files
  - Package-level documentation with purpose, architecture role, and key patterns
  - Go doc style comments for all exported types, functions, and methods
  - Inline comments explaining complex logic and non-obvious implementation details
  - ASCII diagrams for UI layouts, architecture, and data flow
  - Extensive miqt/Qt6 API pattern documentation (constructor suffixes, pointer handling)

### Documentation Highlights

| Layer          | Files Documented | Key Topics                                          |
|----------------|------------------|-----------------------------------------------------|
| Entry Point    | 1                | GPU workaround, Qt initialization sequence          |
| App Controller | 1                | Architecture diagram, thread safety, initialization |
| Domain         | 3                | Entities, validation, sun elevation diagram         |
| Services       | 4                | Solar events, API clients, timezone lookup          |
| Storage        | 1                | JSON persistence, error handling                    |
| UI             | 7                | Widget layouts, miqt patterns, callbacks            |

### Technical

- All documentation follows Go documentation standards
- Comments are designed for `go doc` compatibility
- UI widgets include ASCII layout diagrams
- Critical patterns (mainthread.Wait, QDate pointers) are highlighted

---

## [0.1.2] - 2026-01-02

### Changed

- **Code Optimization**: Refactored codebase for better performance and maintainability
  - Replaced `fmt.Sscanf` with `strconv.ParseFloat` for more efficient float parsing
  - Replaced `fmt.Sprintf("%d")` with `strconv.Itoa` for integer-to-string conversion
  - Added consistent URL construction using `url.Values` in geocoding service

### Added

- **Helper Functions**: Extracted common patterns into reusable helpers
  - `doRequest()` in nominatim.go for HTTP request deduplication
  - `toSampaLocation()` and `extractTimeRange()` in calculator.go
  - `performSearch()` in locationpanel.go
  - `buildLocationURL()` in mapview.go with `defaultZoom` constant

- **Shared Configuration**: Added `config.DefaultHTTPTimeout` constant for consistent HTTP client configuration across services

### Removed

- Empty placeholder directories (`internal/bridge/`, `resources/`)
- Redundant `Validate()` call in `preferences.go` Save method (validation already happens on Load)
- Duplicate HTTP timeout constants (now centralized in config)

### Technical

- Reduced code duplication by ~60 lines through helper extraction
- Improved string formatting efficiency in TimePanel using `fmt.Sprintf`
- All services now use shared timeout configuration from `config.DefaultHTTPTimeout`

---

## [0.1.1] - 2026-01-02

### Fixed

- **GPU Rendering on ARM Devices**: Added `QTWEBENGINE_CHROMIUM_FLAGS="--disable-gpu"` to fix map rendering issues on ARM/Rockchip platforms with problematic GPU drivers (Vulkan/DMA buffer errors)

### Changed

- **TimePanel Layout**: Golden Hour and Blue Hour sections now display side-by-side instead of stacked vertically, reducing vertical space usage by ~100px
  - Labels changed from "Morning/Evening" to "AM/PM" for compactness
  - Removed duration display from time ranges

- **SettingsPanel Layout**: Converted from vertical layout to 2-column grid layout
  - Row 1: Golden Hour spinbox | Blue Start spinbox
  - Row 2: Blue End spinbox | 24-hour format checkbox
  - Row 3: Auto-detect location checkbox (full width)
  - Reduces vertical space by ~60px when expanded

- **DatePanel Layout**: Changed from vertical to horizontal layout
  - "Today" button now inline with navigation: `[<] [date] [>] [Today]`
  - Reduces vertical space by ~30px

### Technical

- Total vertical space reduction in right panel: ~190px
- Improves usability on smaller screens and lower resolution displays

---

## [0.1.0] - 2026-01-01

### Added

- Initial release of GoGoldenHour
- **Core Features**:
  - Golden Hour calculation (configurable sun elevation, default 6°)
  - Blue Hour calculation (configurable start/end elevations, default -4° to -8°)
  - Interactive OpenStreetMap using Qt WebEngine and Leaflet.js
  - IP-based geolocation via ip-api.com
  - Location search via OpenStreetMap Nominatim
  - Date navigation with calendar popup
  - 12-hour and 24-hour time format support

- **UI Components**:
  - Main window with splitter layout (60% map, 40% panels)
  - Location panel with search input and detect button
  - Date panel with previous/next/today navigation
  - Time panel showing Golden Hour and Blue Hour times with durations
  - Collapsible settings panel

- **Persistence**:
  - JSON-based settings storage in ~/.config/gogoldenhour/
  - Remembers last location between sessions
  - Saves all user preferences

- **Architecture**:
  - Clean separation of concerns (domain, services, UI)
  - Domain entities: Location, SunTimes, TimeRange, Settings
  - Services: Solar calculator, Geolocation, Geocoding
  - miqt Qt6 bindings for cross-platform GUI

### Technical Details

- Built with Go 1.22+
- Uses miqt v0.12.0 for Qt6 bindings
- Uses go-sampa v1.0.0 for solar calculations
- Tested on Arch Linux ARM (Khadas Edge2) with Qt 6.10.1

### Known Limitations (v0.1.0)

- ~~Map click events don't communicate back to Go~~ (Fixed in v0.1.1 via console message interception)
- ~~Location changes reload entire map HTML~~ (Fixed in v0.1.1 via URL hash fragment updates)
- MESA GPU warning on ARM devices (cosmetic, fixed in v0.1.1 with GPU disable flag)

## [Unreleased]

### Planned

- Notification/alarm for golden hour start
- Export sun times to calendar
- Dark mode theme
- System tray icon with quick access
- Multiple saved locations
- Sunrise/sunset notifications
- Unit tests for core services

---

## Version History Summary

| Version | Date       | Highlights                                             |
|---------|------------|--------------------------------------------------------|
| 0.1.3   | 2026-01-02 | Comprehensive code documentation for all 17 files     |
| 0.1.2   | 2026-01-02 | Code optimization, helper extraction, shared config   |
| 0.1.1   | 2026-01-02 | GPU fix for ARM devices, compact two-column UI layout |
| 0.1.0   | 2026-01-01 | Initial release with core functionality               |
