package widgets

import (
	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// SettingsPanel
// =============================================================================

// SettingsPanel provides user configuration controls for the application.
//
// This panel allows users to customize:
//   - Sun elevation angles for golden and blue hour boundaries
//   - Time display format (12-hour vs 24-hour)
//   - Auto-detect location on startup behavior
//
// # UI Layout
//
// The panel uses a collapsible group box with a 2-column grid layout:
//
//	┌─ Settings ─────────────────────────────────────────────────┐
//	│ [✓] (click to expand/collapse)                             │
//	├────────────────────────────────────────────────────────────┤
//	│ Golden Hour: [6.0°]      Blue Start: [-4.0°]               │
//	│ Blue End:    [-8.0°]     [✓] 24-hour format                │
//	│ [✓] Auto-detect location on startup                        │
//	└────────────────────────────────────────────────────────────┘
//
// # Elevation Angles
//
// Sun elevation angle settings control when golden/blue hours occur:
//
//	                    Zenith (90°)
//	                         │
//	    Golden Hour ─────────┼───────── Sun at +6° (configurable)
//	    Sunrise/Sunset ──────┼───────── Sun at 0° (horizon)
//	    Blue Hour Start ─────┼───────── Sun at -4° (configurable)
//	    Blue Hour End ───────┼───────── Sun at -8° (configurable)
//	                         │
//	                    Nadir (-90°)
//
// # Initialization Warning
//
// IMPORTANT: This panel triggers onSettingsChange callbacks during construction
// when applySettings() is called. This happens because setting widget values
// fires their OnValueChanged signals. The App must handle this by checking
// if mainWindow is nil in recalculate().
//
// # Communication
//
// Settings changes are communicated via the onSettingsChange callback.
// Each widget change immediately invokes the callback, which:
//  1. Updates the App's internal settings
//  2. Reconfigures the solar calculator
//  3. Persists settings to disk
//  4. Triggers sun time recalculation
type SettingsPanel struct {
	// groupBox is the collapsible container with "Settings" title.
	// The checkable property makes it expandable/collapsible.
	groupBox *qt.QGroupBox

	// goldenElevation sets the sun elevation angle for golden hour boundary.
	// Range: 0° to 15°, default 6°. Higher values give longer golden hour.
	goldenElevation *qt.QDoubleSpinBox

	// blueStartElevation sets when blue hour begins (sun below horizon).
	// Range: -6° to 0°, default -4°. Less negative = earlier start.
	blueStartElevation *qt.QDoubleSpinBox

	// blueEndElevation sets when blue hour ends (deeper twilight).
	// Range: -18° to -6°, default -8°. More negative = later end.
	blueEndElevation *qt.QDoubleSpinBox

	// timeFormatCheck toggles between 12-hour and 24-hour time display.
	// Checked = 24-hour (14:30), Unchecked = 12-hour (2:30 PM)
	timeFormatCheck *qt.QCheckBox

	// autoDetectCheck toggles IP-based location detection on app startup.
	// When enabled, the app queries IP-API to determine initial location.
	autoDetectCheck *qt.QCheckBox

	// settings holds the current settings values.
	// Updated in real-time as widgets change.
	settings domain.Settings

	// onSettingsChange is the callback invoked when any setting changes.
	// Receives the complete updated Settings object.
	onSettingsChange func(settings domain.Settings)
}

// NewSettingsPanel creates a new settings panel with initial values and callback.
//
// Parameters:
//   - settings: Initial settings values to display in the controls
//   - onSettingsChange: Callback invoked whenever any setting changes.
//     The App uses this to update configuration, persist, and recalculate.
//
// Returns a fully initialized SettingsPanel with the given settings applied.
//
// WARNING: This constructor triggers onSettingsChange during initialization
// because applySettings() sets widget values, which fires their change signals.
// The App handles this by checking mainWindow == nil in recalculate().
func NewSettingsPanel(settings domain.Settings, onSettingsChange func(settings domain.Settings)) *SettingsPanel {
	sp := &SettingsPanel{
		settings:         settings,
		onSettingsChange: onSettingsChange,
	}

	sp.setupUI()
	sp.applySettings(settings)
	return sp
}

// setupUI creates and arranges all widgets in the settings panel.
//
// The layout is a 4-column grid within a collapsible group box:
//
//	Row 0: [Label] [Spin] [Label] [Spin]   - Golden Hour & Blue Start
//	Row 1: [Label] [Spin] [Checkbox----]   - Blue End & Time Format
//	Row 2: [Checkbox------------------]    - Auto-detect (spans 4 cols)
//
// # miqt API Notes
//
// Constructor patterns:
//   - NewQGroupBox3("Settings"): Creates group box with title (suffix "3")
//   - NewQDoubleSpinBox2(): Creates spin box (suffix "2" = no params)
//   - NewQCheckBox3("text"): Creates checkbox with text (suffix "3")
//   - NewQLabel3("text"): Creates label with text (suffix "3")
//
// Grid layout methods:
//   - AddWidget2(widget, row, col): Single cell at (row, col)
//   - AddWidget3(widget, row, col, rowSpan, colSpan): Spanning cells
//
// # Collapsible Behavior
//
// The group box is made collapsible using SetCheckable(true). When the
// user unchecks the box, the contents are hidden, saving screen space.
func (sp *SettingsPanel) setupUI() {
	// Create collapsible group box
	// SetCheckable(true) allows expand/collapse via checkbox
	sp.groupBox = qt.NewQGroupBox3("Settings")
	sp.groupBox.SetCheckable(true)
	sp.groupBox.SetChecked(false) // Start collapsed to save space

	// Use grid layout for 2-column arrangement
	layout := qt.NewQGridLayout(sp.groupBox.QWidget)
	layout.SetSpacing(8)

	// =========================================================================
	// Row 0: Golden Hour Elevation | Blue Hour Start Elevation
	// =========================================================================
	// Golden Hour: Sun elevation angle defining the golden hour boundary
	goldenLabel := qt.NewQLabel3("Golden Hour:")
	sp.goldenElevation = qt.NewQDoubleSpinBox2()
	sp.goldenElevation.SetRange(0, 15)     // 0° (horizon) to 15° above
	sp.goldenElevation.SetSingleStep(0.5)  // Fine-grained adjustment
	sp.goldenElevation.SetSuffix("°")      // Show degree symbol
	sp.goldenElevation.OnValueChanged(func(value float64) {
		sp.settings.GoldenHourElevation = value
		sp.notifyChange()
	})
	layout.AddWidget2(goldenLabel.QWidget, 0, 0)
	layout.AddWidget2(sp.goldenElevation.QWidget, 0, 1)

	// Blue Hour Start: Sun elevation when blue hour begins
	blueStartLabel := qt.NewQLabel3("Blue Start:")
	sp.blueStartElevation = qt.NewQDoubleSpinBox2()
	sp.blueStartElevation.SetRange(-6, 0)  // 0° to -6° (civil twilight)
	sp.blueStartElevation.SetSingleStep(0.5)
	sp.blueStartElevation.SetSuffix("°")
	sp.blueStartElevation.OnValueChanged(func(value float64) {
		sp.settings.BlueHourStart = value
		sp.notifyChange()
	})
	layout.AddWidget2(blueStartLabel.QWidget, 0, 2)
	layout.AddWidget2(sp.blueStartElevation.QWidget, 0, 3)

	// =========================================================================
	// Row 1: Blue Hour End Elevation | Time Format Checkbox
	// =========================================================================
	// Blue Hour End: Sun elevation when blue hour ends
	blueEndLabel := qt.NewQLabel3("Blue End:")
	sp.blueEndElevation = qt.NewQDoubleSpinBox2()
	sp.blueEndElevation.SetRange(-18, -6) // -6° to -18° (nautical twilight)
	sp.blueEndElevation.SetSingleStep(0.5)
	sp.blueEndElevation.SetSuffix("°")
	sp.blueEndElevation.OnValueChanged(func(value float64) {
		sp.settings.BlueHourEnd = value
		sp.notifyChange()
	})
	layout.AddWidget2(blueEndLabel.QWidget, 1, 0)
	layout.AddWidget2(sp.blueEndElevation.QWidget, 1, 1)

	// Time Format: Toggle between 12-hour and 24-hour display
	sp.timeFormatCheck = qt.NewQCheckBox3("24-hour format")
	sp.timeFormatCheck.OnStateChanged(func(state int) {
		// Compare to qt.Checked constant to get boolean
		sp.settings.TimeFormat24Hour = state == int(qt.Checked)
		sp.notifyChange()
	})
	// Spans columns 2-3 for alignment
	layout.AddWidget3(sp.timeFormatCheck.QWidget, 1, 2, 1, 2)

	// =========================================================================
	// Row 2: Auto-Detect Location (Full Width)
	// =========================================================================
	// Spans all 4 columns since the label is long
	sp.autoDetectCheck = qt.NewQCheckBox3("Auto-detect location on startup")
	sp.autoDetectCheck.OnStateChanged(func(state int) {
		sp.settings.AutoDetectLocation = state == int(qt.Checked)
		sp.notifyChange()
	})
	layout.AddWidget3(sp.autoDetectCheck.QWidget, 2, 0, 1, 4)
}

// Widget returns the group box container for adding to parent layouts.
//
// The returned QGroupBox contains all settings widgets and can be
// added to a parent layout using layout.AddWidget(panel.Widget().QWidget).
func (sp *SettingsPanel) Widget() *qt.QGroupBox {
	return sp.groupBox
}

// applySettings updates all UI controls to reflect the given settings.
//
// This is called during construction to initialize the controls with
// the user's saved settings.
//
// WARNING: This method triggers OnValueChanged/OnStateChanged callbacks
// because setting widget values fires their change signals. This means
// onSettingsChange will be called during initialization.
//
// The App handles this edge case by checking if mainWindow is nil in
// recalculate(), preventing crashes during initialization.
func (sp *SettingsPanel) applySettings(settings domain.Settings) {
	// Set spin box values (triggers OnValueChanged for each)
	sp.goldenElevation.SetValue(settings.GoldenHourElevation)
	sp.blueStartElevation.SetValue(settings.BlueHourStart)
	sp.blueEndElevation.SetValue(settings.BlueHourEnd)

	// Set checkbox states (triggers OnStateChanged for each)
	// Qt checkboxes use SetCheckState with qt.Checked/qt.Unchecked constants
	if settings.TimeFormat24Hour {
		sp.timeFormatCheck.SetCheckState(qt.Checked)
	} else {
		sp.timeFormatCheck.SetCheckState(qt.Unchecked)
	}

	if settings.AutoDetectLocation {
		sp.autoDetectCheck.SetCheckState(qt.Checked)
	} else {
		sp.autoDetectCheck.SetCheckState(qt.Unchecked)
	}
}

// GetSettings returns the current settings values.
//
// This returns the internal settings struct which is kept in sync with
// the UI controls via their change handlers.
func (sp *SettingsPanel) GetSettings() domain.Settings {
	return sp.settings
}

// notifyChange invokes the settings change callback if set.
//
// This is called by all widget change handlers to propagate changes
// to the App controller. Each change invokes the callback immediately,
// providing real-time settings updates.
func (sp *SettingsPanel) notifyChange() {
	if sp.onSettingsChange != nil {
		sp.onSettingsChange(sp.settings)
	}
}
