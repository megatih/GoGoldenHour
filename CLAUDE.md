# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build the application (release, stripped)
make build                    # Output: build/gogoldenhour

# Build with debug symbols
make build-dev

# Build and run
make run

# Direct build without make
go build -o gogoldenhour ./cmd/gogoldenhour

# Run tests
make test
go test -v ./...

# Run go vet
make vet

# Clean build artifacts
make clean
```

## System Requirements

Requires Qt6 with WebEngine:
- **Arch Linux**: `qt6-base qt6-webengine qt6-webchannel`
- **Debian/Ubuntu**: `qt6-base-dev qt6-webengine-dev`

## Architecture

GoGoldenHour is a Qt6 desktop app built with miqt bindings. The architecture follows a controller pattern:

```
cmd/gogoldenhour/main.go
    └── internal/app/App (controller)
            ├── ui/MainWindow (view)
            │       └── widgets/* (UI components)
            ├── service/solar/Calculator (go-sampa)
            ├── service/geolocation/IPAPIService
            ├── service/geocoding/NominatimService
            └── storage/PreferencesStore
```

**App** (`internal/app/app.go`) orchestrates everything:
- Implements `ui.AppController` interface
- Owns services and coordinates data flow
- Handles async operations with `mainthread.Wait()` for Qt thread safety

**MainWindow** (`internal/ui/mainwindow.go`) manages the UI:
- Creates and arranges widget panels
- Delegates user actions to the controller
- Has nil checks in update methods to handle initialization timing

**Widgets** (`internal/ui/widgets/`):
- `mapview.go` - Qt WebEngine + Leaflet.js (no RunJavaScript in miqt, reloads HTML to update)
- `timepanel.go` - Displays golden/blue hour times in side-by-side columns (Golden Hour | Blue Hour)
- `locationpanel.go` - Search and location display
- `datepanel.go` - Horizontal date navigation with inline Today button
- `settingspanel.go` - Collapsible settings with 2-column grid layout (triggers callbacks during init, beware)

## miqt v0.12.0 API Patterns

Widget constructors with text parameter use suffix "3":
```go
qt.NewQGroupBox3("Title")
qt.NewQLabel3("Text")
qt.NewQPushButton3("Click")
```

No-argument constructors use suffix "2":
```go
qt.NewQLineEdit2()
qt.NewQDateEdit2()
```

Date handling returns pointers:
```go
currentDate := qt.QDate_CurrentDate()  // *QDate
dateEdit.SetDate(*currentDate)         // dereference
newDate := currentDate.AddDays(1)      // *QDate
```

Layout methods take single arguments (no stretch parameter):
```go
layout.AddWidget(widget.QWidget)
layout.AddLayout(sublayout.QLayout)
```

Grid layout for 2-column arrangements:
```go
layout := qt.NewQGridLayout(parent.QWidget)
layout.AddWidget2(widget.QWidget, row, col)           // Single cell
layout.AddWidget3(widget.QWidget, row, col, rowSpan, colSpan)  // Spanning
```

## Key Limitations

1. **No RunJavaScript**: miqt doesn't expose `QWebEnginePage.RunJavaScript()`. Map updates by regenerating and reloading HTML.

2. **Initialization callbacks**: SettingsPanel triggers `OnValueChanged` during `applySettings()`. The App must check `mainWindow == nil` in `recalculate()`.

3. **Qt thread safety**: Use `mainthread.Wait()` when updating Qt widgets from goroutines.

## GPU Compatibility

The app sets `QTWEBENGINE_CHROMIUM_FLAGS="--disable-gpu"` before Qt initialization to fix rendering issues on ARM/Rockchip platforms. This is done in `main.go` before `qt.NewQApplication()`.

## Domain Entities

- `domain.Location` - lat/lon/name with validation
- `domain.SunTimes` - sunrise/sunset + golden/blue hour ranges
- `domain.TimeRange` - start/end times with `IsValid()` and duration formatting
- `domain.Settings` - elevation angles, time format, auto-detect preference
