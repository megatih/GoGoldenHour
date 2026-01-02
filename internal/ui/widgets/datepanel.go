package widgets

import (
	"time"

	qt "github.com/mappu/miqt/qt6"
)

// =============================================================================
// DatePanel
// =============================================================================

// DatePanel provides date selection and navigation for solar calculations.
//
// This panel allows users to select a date for which golden/blue hour times
// will be calculated. The panel provides multiple ways to navigate dates:
//   - Previous/Next buttons for single-day navigation
//   - Calendar popup for selecting any date
//   - Today button to quickly return to current date
//
// # UI Layout
//
//	┌─ Date ─────────────────────────────────────────────┐
//	│ [<] [    January 2, 2026    ▼] [>] [  Today  ]     │
//	└────────────────────────────────────────────────────┘
//	 ▲        ▲                    ▲        ▲
//	 │        │                    │        └── Reset to today
//	 │        │                    └── Next day
//	 │        └── Date with calendar popup
//	 └── Previous day
//
// # Date Handling
//
// The panel converts between Go's time.Time and Qt's QDate:
//   - Qt QDate methods return *QDate (pointers)
//   - Must dereference when calling SetDate: dateEdit.SetDate(*qdate)
//   - time.Time uses 1-indexed months, QDate uses 1-indexed months (compatible)
//
// # Communication
//
// Date changes are communicated via the onDateChange callback. This callback
// is invoked whenever the date changes (button click, calendar selection, etc.).
// The App uses this to recalculate sun times for the new date.
type DatePanel struct {
	// groupBox is the container widget with "Date" title border.
	groupBox *qt.QGroupBox

	// dateEdit is the date picker with calendar popup support.
	// Displays dates in "MMMM d, yyyy" format (e.g., "January 2, 2026").
	dateEdit *qt.QDateEdit

	// prevBtn navigates to the previous day ("<" button).
	prevBtn *qt.QPushButton

	// nextBtn navigates to the next day (">" button).
	nextBtn *qt.QPushButton

	// todayBtn resets the date to today's date.
	todayBtn *qt.QPushButton

	// onDateChange is the callback invoked when the date changes.
	// Receives the new date as time.Time.
	onDateChange func(date time.Time)
}

// NewDatePanel creates a new date panel with the given callback.
//
// Parameters:
//   - onDateChange: Callback invoked when the selected date changes.
//     The App uses this to recalculate sun times for the new date.
//
// Returns a fully initialized DatePanel with today's date selected.
func NewDatePanel(onDateChange func(date time.Time)) *DatePanel {
	dp := &DatePanel{
		onDateChange: onDateChange,
	}

	dp.setupUI()
	return dp
}

// setupUI creates and arranges all widgets in the date panel.
//
// The layout is horizontal: [<] [Date Picker] [>] [Today]
//
// # miqt API Notes
//
// Date handling quirks:
//   - QDate_CurrentDate() returns *QDate (pointer)
//   - SetDate() requires dereferenced value: SetDate(*qdate)
//   - AddDays() also returns *QDate, must dereference
//
// Constructor patterns:
//   - NewQGroupBox3("Date"): Creates group box with title (suffix "3")
//   - NewQDateEdit2(): Creates date edit (suffix "2" = no params)
//   - NewQPushButton3("text"): Creates button with text (suffix "3")
func (dp *DatePanel) setupUI() {
	// Create group box container with horizontal layout
	dp.groupBox = qt.NewQGroupBox3("Date")
	layout := qt.NewQHBoxLayout(dp.groupBox.QWidget)
	layout.SetSpacing(6)

	// =========================================================================
	// Previous Day Button
	// =========================================================================
	dp.prevBtn = qt.NewQPushButton3("<")
	dp.prevBtn.SetFixedWidth(40)
	dp.prevBtn.OnClicked(func() {
		dp.changeDate(-1) // Go back one day
	})
	layout.AddWidget(dp.prevBtn.QWidget)

	// =========================================================================
	// Date Picker with Calendar Popup
	// =========================================================================
	// NewQDateEdit2: suffix "2" = no-parameter constructor
	dp.dateEdit = qt.NewQDateEdit2()
	dp.dateEdit.SetCalendarPopup(true) // Enable dropdown calendar
	dp.dateEdit.SetDisplayFormat("MMMM d, yyyy") // e.g., "January 2, 2026"

	// Set initial date to today
	// IMPORTANT: QDate_CurrentDate() returns *QDate (pointer)
	// Must dereference when calling SetDate()
	currentDate := qt.QDate_CurrentDate()
	dp.dateEdit.SetDate(*currentDate)

	// Connect date change signal to our callback handler
	dp.dateEdit.OnDateChanged(func(date qt.QDate) {
		dp.notifyDateChange()
	})
	layout.AddWidget(dp.dateEdit.QWidget)

	// =========================================================================
	// Next Day Button
	// =========================================================================
	dp.nextBtn = qt.NewQPushButton3(">")
	dp.nextBtn.SetFixedWidth(40)
	dp.nextBtn.OnClicked(func() {
		dp.changeDate(1) // Go forward one day
	})
	layout.AddWidget(dp.nextBtn.QWidget)

	// =========================================================================
	// Today Button (Quick Reset)
	// =========================================================================
	// Inline with navigation buttons for compact layout
	dp.todayBtn = qt.NewQPushButton3("Today")
	dp.todayBtn.OnClicked(func() {
		// Reset to current date
		// Same pattern: dereference the *QDate pointer
		currentDate := qt.QDate_CurrentDate()
		dp.dateEdit.SetDate(*currentDate)
	})
	layout.AddWidget(dp.todayBtn.QWidget)
}

// Widget returns the group box container for adding to parent layouts.
//
// The returned QGroupBox contains all date panel widgets and can be
// added to a parent layout using layout.AddWidget(panel.Widget().QWidget).
func (dp *DatePanel) Widget() *qt.QGroupBox {
	return dp.groupBox
}

// SetDate sets the displayed date from a Go time.Time value.
//
// This method converts time.Time to QDate for Qt compatibility.
// Called by MainWindow when the App updates the date programmatically.
//
// # Type Conversion
//
// Go time.Time → Qt QDate:
//   - NewQDate2(year, month, day) creates QDate (suffix "2" = 3 int params)
//   - Must dereference: SetDate(*qdate)
func (dp *DatePanel) SetDate(date time.Time) {
	qdate := qt.NewQDate2(date.Year(), int(date.Month()), date.Day())
	dp.dateEdit.SetDate(*qdate)
}

// GetDate returns the currently selected date as Go time.Time.
//
// This method converts Qt's QDate to Go's time.Time for use in the
// domain layer. The returned time is at midnight (00:00:00) in local timezone.
//
// # Type Conversion
//
// Qt QDate → Go time.Time:
//   - Extract year, month, day from QDate
//   - Build time.Time with time.Date()
//   - Month conversion: QDate.Month() (int) → time.Month (type cast)
func (dp *DatePanel) GetDate() time.Time {
	qdate := dp.dateEdit.Date()
	return time.Date(qdate.Year(), time.Month(qdate.Month()), qdate.Day(), 0, 0, 0, 0, time.Local)
}

// changeDate adjusts the date by the specified number of days.
//
// This is called by the previous/next buttons to navigate dates.
//
// Parameters:
//   - days: Number of days to add (negative for previous, positive for next)
//
// # miqt Note
//
// AddDays() returns *QDate (pointer), must dereference when setting.
func (dp *DatePanel) changeDate(days int) {
	currentDate := dp.dateEdit.Date()
	newDate := currentDate.AddDays(int64(days))
	dp.dateEdit.SetDate(*newDate)
}

// notifyDateChange invokes the date change callback if set.
//
// This is called whenever the date changes, whether from:
//   - Previous/Next button clicks
//   - Calendar popup selection
//   - Today button click
//   - Programmatic SetDate() calls (which trigger OnDateChanged)
//
// The callback receives the date converted to Go time.Time format.
func (dp *DatePanel) notifyDateChange() {
	if dp.onDateChange != nil {
		dp.onDateChange(dp.GetDate())
	}
}
