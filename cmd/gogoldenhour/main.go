package main

import (
	"log"
	"os"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/app"
)

func main() {
	// Disable GPU acceleration for Qt WebEngine to avoid rendering issues
	// on systems with problematic GPU drivers (e.g., ARM/Rockchip)
	os.Setenv("QTWEBENGINE_CHROMIUM_FLAGS", "--disable-gpu")

	// Initialize Qt application
	// This automatically locks the OS thread for Qt
	qt.NewQApplication(os.Args)

	// Create and run the application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Run the application
	application.Run()

	// Start Qt event loop
	os.Exit(qt.QApplication_Exec())
}
