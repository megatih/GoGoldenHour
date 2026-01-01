package widgets

import (
	"time"

	qt "github.com/mappu/miqt/qt6"
)

// DatePanel allows date selection
type DatePanel struct {
	groupBox     *qt.QGroupBox
	dateEdit     *qt.QDateEdit
	prevBtn      *qt.QPushButton
	nextBtn      *qt.QPushButton
	todayBtn     *qt.QPushButton
	onDateChange func(date time.Time)
}

// NewDatePanel creates a new date panel
func NewDatePanel(onDateChange func(date time.Time)) *DatePanel {
	dp := &DatePanel{
		onDateChange: onDateChange,
	}

	dp.setupUI()
	return dp
}

// setupUI creates the date panel UI
func (dp *DatePanel) setupUI() {
	dp.groupBox = qt.NewQGroupBox3("Date")
	layout := qt.NewQHBoxLayout(dp.groupBox.QWidget)
	layout.SetSpacing(6)

	// Previous button
	dp.prevBtn = qt.NewQPushButton3("<")
	dp.prevBtn.SetFixedWidth(40)
	dp.prevBtn.OnClicked(func() {
		dp.changeDate(-1)
	})
	layout.AddWidget(dp.prevBtn.QWidget)

	// Date edit
	dp.dateEdit = qt.NewQDateEdit2()
	dp.dateEdit.SetCalendarPopup(true)
	dp.dateEdit.SetDisplayFormat("MMMM d, yyyy")
	currentDate := qt.QDate_CurrentDate()
	dp.dateEdit.SetDate(*currentDate)
	dp.dateEdit.OnDateChanged(func(date qt.QDate) {
		dp.notifyDateChange()
	})
	layout.AddWidget(dp.dateEdit.QWidget)

	// Next button
	dp.nextBtn = qt.NewQPushButton3(">")
	dp.nextBtn.SetFixedWidth(40)
	dp.nextBtn.OnClicked(func() {
		dp.changeDate(1)
	})
	layout.AddWidget(dp.nextBtn.QWidget)

	// Today button (inline)
	dp.todayBtn = qt.NewQPushButton3("Today")
	dp.todayBtn.OnClicked(func() {
		currentDate := qt.QDate_CurrentDate()
		dp.dateEdit.SetDate(*currentDate)
	})
	layout.AddWidget(dp.todayBtn.QWidget)
}

// Widget returns the group box widget
func (dp *DatePanel) Widget() *qt.QGroupBox {
	return dp.groupBox
}

// SetDate sets the current date
func (dp *DatePanel) SetDate(date time.Time) {
	qdate := qt.NewQDate2(date.Year(), int(date.Month()), date.Day())
	dp.dateEdit.SetDate(*qdate)
}

// GetDate returns the currently selected date
func (dp *DatePanel) GetDate() time.Time {
	qdate := dp.dateEdit.Date()
	return time.Date(qdate.Year(), time.Month(qdate.Month()), qdate.Day(), 0, 0, 0, 0, time.Local)
}

// changeDate changes the date by the given number of days
func (dp *DatePanel) changeDate(days int) {
	currentDate := dp.dateEdit.Date()
	newDate := currentDate.AddDays(int64(days))
	dp.dateEdit.SetDate(*newDate)
}

// notifyDateChange notifies the callback of a date change
func (dp *DatePanel) notifyDateChange() {
	if dp.onDateChange != nil {
		dp.onDateChange(dp.GetDate())
	}
}
