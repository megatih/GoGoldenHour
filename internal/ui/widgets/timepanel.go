package widgets

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// =============================================================================
// TimePanel
// =============================================================================

// TimePanel displays calculated golden hour and blue hour times.
//
// This is the primary output panel showing the results of solar calculations.
// It displays sunrise/sunset times and the four key photography periods:
//   - Morning golden hour (after sunrise)
//   - Evening golden hour (before sunset)
//   - Morning blue hour (before sunrise)
//   - Evening blue hour (after sunset)
//
// # UI Layout
//
//	┌─ Sun Times ───────────────────────────────────────────────┐
//	│ Sunrise: 07:15                  Sunset: 17:45             │
//	│ ┌─ Golden Hour ──────────┐ ┌─ Blue Hour ───────────┐      │
//	│ │ AM: 07:15 - 08:15      │ │ AM: 06:45 - 07:15     │      │
//	│ │ PM: 16:45 - 17:45      │ │ PM: 17:45 - 18:15     │      │
//	│ └────────────────────────┘ └───────────────────────┘      │
//	└───────────────────────────────────────────────────────────┘
//
// # Styling
//
// Each group has distinctive styling matching the lighting conditions:
//   - Golden Hour: Orange border (#ff9800) representing warm light
//   - Blue Hour: Blue border (#2196f3) representing cool twilight
//
// # Time Validation
//
// Some time ranges may be invalid for certain dates/locations:
//   - Polar regions during midnight sun have no blue hour
//   - Polar regions during polar night may have no sunrise/sunset
//   - Invalid ranges display "N/A" instead of times
type TimePanel struct {
	// groupBox is the outer container with "Sun Times" title.
	groupBox *qt.QGroupBox

	// goldenGroup is the nested group box for golden hour times.
	// Styled with orange border matching golden hour lighting.
	goldenGroup *qt.QGroupBox

	// blueGroup is the nested group box for blue hour times.
	// Styled with blue border matching blue hour lighting.
	blueGroup *qt.QGroupBox

	// goldenMorning displays the morning golden hour time range.
	// Shows "AM: HH:MM - HH:MM" or "AM: N/A" if invalid.
	goldenMorning *qt.QLabel

	// goldenEvening displays the evening golden hour time range.
	// Shows "PM: HH:MM - HH:MM" or "PM: N/A" if invalid.
	goldenEvening *qt.QLabel

	// blueMorning displays the morning blue hour time range.
	// Shows "AM: HH:MM - HH:MM" or "AM: N/A" if invalid.
	blueMorning *qt.QLabel

	// blueEvening displays the evening blue hour time range.
	// Shows "PM: HH:MM - HH:MM" or "PM: N/A" if invalid.
	blueEvening *qt.QLabel

	// sunriseLabel displays the sunrise time.
	sunriseLabel *qt.QLabel

	// sunsetLabel displays the sunset time.
	sunsetLabel *qt.QLabel

	// use24Hour determines the time display format.
	// true: 24-hour format (14:30), false: 12-hour format (2:30 PM)
	use24Hour bool
}

// NewTimePanel creates a new time panel with the specified time format.
//
// Parameters:
//   - use24Hour: If true, display times in 24-hour format (14:30).
//     If false, display in 12-hour format (2:30 PM).
//
// Returns a fully initialized TimePanel showing placeholder times ("--:--").
// Call SetSunTimes() to update with actual calculated values.
func NewTimePanel(use24Hour bool) *TimePanel {
	tp := &TimePanel{use24Hour: use24Hour}
	tp.setupUI()
	return tp
}

// setupUI creates and arranges all widgets in the time panel.
//
// The layout structure:
//  1. Sunrise/Sunset row at top (horizontal)
//  2. Two side-by-side group boxes below (horizontal):
//     - Golden Hour group (orange styled)
//     - Blue Hour group (blue styled)
//
// Each hour group contains AM and PM time ranges stacked vertically.
//
// # Styling
//
// Qt stylesheet is used for custom group box appearance:
//   - Colored borders (orange/blue) match the lighting type
//   - Rounded corners for modern look
//   - Colored titles positioned within the border
func (tp *TimePanel) setupUI() {
	// Create outer container with "Sun Times" title
	tp.groupBox = qt.NewQGroupBox3("Sun Times")
	mainLayout := qt.NewQVBoxLayout(tp.groupBox.QWidget)
	mainLayout.SetSpacing(8)

	// =========================================================================
	// Sunrise/Sunset Row
	// =========================================================================
	// Display these prominently at the top
	sunLayout := qt.NewQHBoxLayout2()
	tp.sunriseLabel = qt.NewQLabel3("Sunrise: --:--")
	tp.sunsetLabel = qt.NewQLabel3("Sunset: --:--")
	tp.sunriseLabel.SetStyleSheet("font-weight: bold;")
	tp.sunsetLabel.SetStyleSheet("font-weight: bold;")
	sunLayout.AddWidget(tp.sunriseLabel.QWidget)
	sunLayout.AddWidget(tp.sunsetLabel.QWidget)
	mainLayout.AddLayout(sunLayout.QLayout)

	// =========================================================================
	// Golden Hour and Blue Hour Groups (Side by Side)
	// =========================================================================
	hoursLayout := qt.NewQHBoxLayout2()
	hoursLayout.SetSpacing(8)

	// -------------------------------------------------------------------------
	// Golden Hour Group (Orange Theme)
	// -------------------------------------------------------------------------
	// Styled with warm orange color representing the golden light quality
	tp.goldenGroup = qt.NewQGroupBox3("Golden Hour")
	tp.goldenGroup.SetStyleSheet(`
		QGroupBox {
			font-weight: bold;
			border: 2px solid #ff9800;
			border-radius: 6px;
			margin-top: 10px;
			padding-top: 10px;
		}
		QGroupBox::title {
			subcontrol-origin: margin;
			left: 10px;
			padding: 0 5px;
			color: #ff9800;
		}
	`)
	goldenLayout := qt.NewQVBoxLayout(tp.goldenGroup.QWidget)
	goldenLayout.SetSpacing(4)

	// Morning and evening time range labels
	tp.goldenMorning = qt.NewQLabel3("AM: --:-- - --:--")
	tp.goldenEvening = qt.NewQLabel3("PM: --:-- - --:--")
	goldenLayout.AddWidget(tp.goldenMorning.QWidget)
	goldenLayout.AddWidget(tp.goldenEvening.QWidget)

	hoursLayout.AddWidget(tp.goldenGroup.QWidget)

	// -------------------------------------------------------------------------
	// Blue Hour Group (Blue Theme)
	// -------------------------------------------------------------------------
	// Styled with cool blue color representing the twilight light quality
	tp.blueGroup = qt.NewQGroupBox3("Blue Hour")
	tp.blueGroup.SetStyleSheet(`
		QGroupBox {
			font-weight: bold;
			border: 2px solid #2196f3;
			border-radius: 6px;
			margin-top: 10px;
			padding-top: 10px;
		}
		QGroupBox::title {
			subcontrol-origin: margin;
			left: 10px;
			padding: 0 5px;
			color: #2196f3;
		}
	`)
	blueLayout := qt.NewQVBoxLayout(tp.blueGroup.QWidget)
	blueLayout.SetSpacing(4)

	// Morning and evening time range labels
	tp.blueMorning = qt.NewQLabel3("AM: --:-- - --:--")
	tp.blueEvening = qt.NewQLabel3("PM: --:-- - --:--")
	blueLayout.AddWidget(tp.blueMorning.QWidget)
	blueLayout.AddWidget(tp.blueEvening.QWidget)

	hoursLayout.AddWidget(tp.blueGroup.QWidget)

	mainLayout.AddLayout(hoursLayout.QLayout)
}

// Widget returns the group box container for adding to parent layouts.
//
// The returned QGroupBox contains all time display widgets and can be
// added to a parent layout using layout.AddWidget(panel.Widget().QWidget).
func (tp *TimePanel) Widget() *qt.QGroupBox {
	return tp.groupBox
}

// SetSunTimes updates all displayed times from calculated sun times.
//
// This is called by MainWindow whenever sun times are recalculated due to:
//   - Location change
//   - Date change
//   - Settings change (elevation angles)
//
// Parameters:
//   - st: The calculated sun times from the solar calculator
//   - use24Hour: Time format preference (true = 24h, false = 12h)
//
// # Time Range Validation
//
// Each time range is checked with IsValid() before display. Invalid ranges
// occur in polar regions during extreme seasons (midnight sun, polar night).
// Invalid ranges show "N/A" instead of times.
//
// # Time Formatting
//
// Times are formatted using domain.FormatTime() which respects the use24Hour
// setting. Examples:
//   - 24-hour: "14:30"
//   - 12-hour: "2:30 PM"
func (tp *TimePanel) SetSunTimes(st domain.SunTimes, use24Hour bool) {
	tp.use24Hour = use24Hour

	// -------------------------------------------------------------------------
	// Sunrise and Sunset (always valid for non-polar regions)
	// -------------------------------------------------------------------------
	tp.sunriseLabel.SetText(fmt.Sprintf("Sunrise: %s", domain.FormatTime(st.Sunrise, use24Hour)))
	tp.sunsetLabel.SetText(fmt.Sprintf("Sunset: %s", domain.FormatTime(st.Sunset, use24Hour)))

	// -------------------------------------------------------------------------
	// Golden Hour Times
	// -------------------------------------------------------------------------
	// Morning golden hour occurs just after sunrise
	if st.GoldenMorning.IsValid() {
		tp.goldenMorning.SetText(fmt.Sprintf("AM: %s - %s",
			domain.FormatTime(st.GoldenMorning.Start, use24Hour),
			domain.FormatTime(st.GoldenMorning.End, use24Hour)))
	} else {
		tp.goldenMorning.SetText("AM: N/A")
	}

	// Evening golden hour occurs just before sunset
	if st.GoldenEvening.IsValid() {
		tp.goldenEvening.SetText(fmt.Sprintf("PM: %s - %s",
			domain.FormatTime(st.GoldenEvening.Start, use24Hour),
			domain.FormatTime(st.GoldenEvening.End, use24Hour)))
	} else {
		tp.goldenEvening.SetText("PM: N/A")
	}

	// -------------------------------------------------------------------------
	// Blue Hour Times
	// -------------------------------------------------------------------------
	// Morning blue hour occurs just before sunrise
	if st.BlueMorning.IsValid() {
		tp.blueMorning.SetText(fmt.Sprintf("AM: %s - %s",
			domain.FormatTime(st.BlueMorning.Start, use24Hour),
			domain.FormatTime(st.BlueMorning.End, use24Hour)))
	} else {
		tp.blueMorning.SetText("AM: N/A")
	}

	// Evening blue hour occurs just after sunset
	if st.BlueEvening.IsValid() {
		tp.blueEvening.SetText(fmt.Sprintf("PM: %s - %s",
			domain.FormatTime(st.BlueEvening.Start, use24Hour),
			domain.FormatTime(st.BlueEvening.End, use24Hour)))
	} else {
		tp.blueEvening.SetText("PM: N/A")
	}
}

// SetTimeFormat updates the stored time format preference.
//
// This stores the preference but does not update the display. A subsequent
// call to SetSunTimes() will use the new format.
//
// Note: This method is provided for settings changes, but SetSunTimes()
// takes the format as a parameter for simplicity, so this may be redundant.
func (tp *TimePanel) SetTimeFormat(use24Hour bool) {
	tp.use24Hour = use24Hour
}
