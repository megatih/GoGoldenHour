package widgets

import (
	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// TimePanel displays Golden Hour and Blue Hour times
type TimePanel struct {
	groupBox      *qt.QGroupBox
	goldenGroup   *qt.QGroupBox
	blueGroup     *qt.QGroupBox
	goldenMorning *qt.QLabel
	goldenEvening *qt.QLabel
	blueMorning   *qt.QLabel
	blueEvening   *qt.QLabel
	sunriseLabel  *qt.QLabel
	sunsetLabel   *qt.QLabel
	use24Hour     bool
}

// NewTimePanel creates a new time panel
func NewTimePanel(use24Hour bool) *TimePanel {
	tp := &TimePanel{use24Hour: use24Hour}
	tp.setupUI()
	return tp
}

// setupUI creates the time panel UI
func (tp *TimePanel) setupUI() {
	tp.groupBox = qt.NewQGroupBox3("Sun Times")
	mainLayout := qt.NewQVBoxLayout(tp.groupBox.QWidget)
	mainLayout.SetSpacing(8)

	// Sunrise/Sunset row
	sunLayout := qt.NewQHBoxLayout2()
	tp.sunriseLabel = qt.NewQLabel3("Sunrise: --:--")
	tp.sunsetLabel = qt.NewQLabel3("Sunset: --:--")
	tp.sunriseLabel.SetStyleSheet("font-weight: bold;")
	tp.sunsetLabel.SetStyleSheet("font-weight: bold;")
	sunLayout.AddWidget(tp.sunriseLabel.QWidget)
	sunLayout.AddWidget(tp.sunsetLabel.QWidget)
	mainLayout.AddLayout(sunLayout.QLayout)

	// Side-by-side layout for Golden Hour and Blue Hour
	hoursLayout := qt.NewQHBoxLayout2()
	hoursLayout.SetSpacing(8)

	// Golden Hour group
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

	tp.goldenMorning = qt.NewQLabel3("AM: --:-- - --:--")
	tp.goldenEvening = qt.NewQLabel3("PM: --:-- - --:--")
	goldenLayout.AddWidget(tp.goldenMorning.QWidget)
	goldenLayout.AddWidget(tp.goldenEvening.QWidget)

	hoursLayout.AddWidget(tp.goldenGroup.QWidget)

	// Blue Hour group
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

	tp.blueMorning = qt.NewQLabel3("AM: --:-- - --:--")
	tp.blueEvening = qt.NewQLabel3("PM: --:-- - --:--")
	blueLayout.AddWidget(tp.blueMorning.QWidget)
	blueLayout.AddWidget(tp.blueEvening.QWidget)

	hoursLayout.AddWidget(tp.blueGroup.QWidget)

	mainLayout.AddLayout(hoursLayout.QLayout)
}

// Widget returns the group box widget
func (tp *TimePanel) Widget() *qt.QGroupBox {
	return tp.groupBox
}

// SetSunTimes updates the displayed sun times
func (tp *TimePanel) SetSunTimes(st domain.SunTimes, use24Hour bool) {
	tp.use24Hour = use24Hour

	// Sunrise and sunset
	tp.sunriseLabel.SetText("Sunrise: " + domain.FormatTime(st.Sunrise, use24Hour))
	tp.sunsetLabel.SetText("Sunset: " + domain.FormatTime(st.Sunset, use24Hour))

	// Golden Hour
	if st.GoldenMorning.IsValid() {
		tp.goldenMorning.SetText("AM: " + domain.FormatTime(st.GoldenMorning.Start, use24Hour) +
			" - " + domain.FormatTime(st.GoldenMorning.End, use24Hour))
	} else {
		tp.goldenMorning.SetText("AM: N/A")
	}

	if st.GoldenEvening.IsValid() {
		tp.goldenEvening.SetText("PM: " + domain.FormatTime(st.GoldenEvening.Start, use24Hour) +
			" - " + domain.FormatTime(st.GoldenEvening.End, use24Hour))
	} else {
		tp.goldenEvening.SetText("PM: N/A")
	}

	// Blue Hour
	if st.BlueMorning.IsValid() {
		tp.blueMorning.SetText("AM: " + domain.FormatTime(st.BlueMorning.Start, use24Hour) +
			" - " + domain.FormatTime(st.BlueMorning.End, use24Hour))
	} else {
		tp.blueMorning.SetText("AM: N/A")
	}

	if st.BlueEvening.IsValid() {
		tp.blueEvening.SetText("PM: " + domain.FormatTime(st.BlueEvening.Start, use24Hour) +
			" - " + domain.FormatTime(st.BlueEvening.End, use24Hour))
	} else {
		tp.blueEvening.SetText("PM: N/A")
	}
}

// SetTimeFormat updates the time format
func (tp *TimePanel) SetTimeFormat(use24Hour bool) {
	tp.use24Hour = use24Hour
}
