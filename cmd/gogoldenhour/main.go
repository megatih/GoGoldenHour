// Package main is the entry point for the GoGoldenHour application.
//
// GoGoldenHour is a Qt6 desktop application that calculates golden hour and
// blue hour times for photographers. It displays an interactive map and allows
// users to search for locations or detect their position automatically.
//
// # Application Architecture
//
// The application follows a controller pattern:
//
//	main.go (entry point)
//	    └── app.App (controller/orchestrator)
//	            ├── ui.MainWindow (view)
//	            │       └── widgets/* (UI components)
//	            ├── solar.Calculator (calculations)
//	            ├── geolocation.IPAPIService (IP detection)
//	            ├── geocoding.NominatimService (search)
//	            └── storage.PreferencesStore (persistence)
//
// # Qt Integration
//
// This application uses miqt (https://github.com/mappu/miqt), a Go binding
// for Qt6. The miqt library handles the Qt/Go interop, including:
//   - Thread locking (Qt must run on the main OS thread)
//   - Event loop integration
//   - Signal/slot connections
//
// # GPU Compatibility
//
// Qt WebEngine (used for the map) uses Chromium internally. Some GPU drivers
// (particularly on ARM/Rockchip platforms) have compatibility issues with
// Chromium's GPU acceleration. The application disables GPU acceleration
// before Qt initialization to ensure reliable rendering on all platforms.
//
// # Startup Flow
//
//  1. Disable GPU acceleration (environment variable)
//  2. Initialize Qt application (locks OS thread)
//  3. Create App controller (loads settings, creates services)
//  4. Run the application (shows window, optionally auto-detects location)
//  5. Enter Qt event loop (handles user interactions)
//  6. Exit when user closes the window
package main

import (
	"log"
	"os"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/app"
)

// main is the entry point of the GoGoldenHour application.
//
// This function performs the following initialization steps:
//  1. Sets environment variable to disable GPU acceleration
//  2. Initializes the Qt application framework
//  3. Creates the application controller
//  4. Starts the application and Qt event loop
//
// The function exits the process with the Qt application's exit code,
// which is typically 0 for normal exit or non-zero for errors.
func main() {
	// =========================================================================
	// Step 1: GPU Compatibility Fix
	// =========================================================================
	// Disable GPU acceleration for Qt WebEngine (Chromium) to avoid rendering
	// issues on systems with problematic GPU drivers.
	//
	// This is particularly important for:
	//   - ARM/Rockchip platforms (common in single-board computers)
	//   - Virtual machines without GPU passthrough
	//   - Systems with outdated or proprietary GPU drivers
	//
	// IMPORTANT: This must be set BEFORE qt.NewQApplication() is called.
	// Once Chromium initializes, the GPU settings cannot be changed.
	os.Setenv("QTWEBENGINE_CHROMIUM_FLAGS", "--disable-gpu")

	// =========================================================================
	// Step 2: Qt Application Initialization
	// =========================================================================
	// Initialize the Qt application framework. This call:
	//   - Locks the current goroutine to the OS thread (required by Qt)
	//   - Parses command-line arguments for Qt-specific options
	//   - Sets up the Qt event loop infrastructure
	//
	// After this call, the current goroutine is permanently bound to the
	// main thread. All Qt widget operations must happen on this thread.
	qt.NewQApplication(os.Args)

	// =========================================================================
	// Step 3: Application Controller Creation
	// =========================================================================
	// Create the main application controller. This performs:
	//   - Loading user preferences from disk
	//   - Creating all services (solar calculator, geocoding, etc.)
	//   - Setting up the main window with all UI components
	//
	// Errors at this stage are fatal (e.g., cannot create preferences store).
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// =========================================================================
	// Step 4: Application Startup
	// =========================================================================
	// Start the application. This:
	//   - Shows the main window
	//   - Optionally auto-detects the user's location (if enabled in settings)
	//   - Performs initial solar calculations
	application.Run()

	// =========================================================================
	// Step 5: Qt Event Loop
	// =========================================================================
	// Enter the Qt event loop. This function blocks until the application
	// exits (user closes the window or calls QApplication::quit()).
	//
	// During this time, Qt processes:
	//   - User input events (mouse, keyboard)
	//   - Widget repaint events
	//   - Timer events
	//   - Network events (for map tiles, API requests)
	//
	// The return value is the application's exit code (0 = success).
	os.Exit(qt.QApplication_Exec())
}
