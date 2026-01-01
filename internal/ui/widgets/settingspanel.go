package widgets

import (
	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// SettingsPanel allows users to configure application settings
type SettingsPanel struct {
	groupBox           *qt.QGroupBox
	goldenElevation    *qt.QDoubleSpinBox
	blueStartElevation *qt.QDoubleSpinBox
	blueEndElevation   *qt.QDoubleSpinBox
	timeFormatCheck    *qt.QCheckBox
	autoDetectCheck    *qt.QCheckBox
	settings           domain.Settings
	onSettingsChange   func(settings domain.Settings)
}

// NewSettingsPanel creates a new settings panel
func NewSettingsPanel(settings domain.Settings, onSettingsChange func(settings domain.Settings)) *SettingsPanel {
	sp := &SettingsPanel{
		settings:         settings,
		onSettingsChange: onSettingsChange,
	}

	sp.setupUI()
	sp.applySettings(settings)
	return sp
}

// setupUI creates the settings panel UI
func (sp *SettingsPanel) setupUI() {
	sp.groupBox = qt.NewQGroupBox3("Settings")
	sp.groupBox.SetCheckable(true)
	sp.groupBox.SetChecked(false) // Start collapsed

	layout := qt.NewQGridLayout(sp.groupBox.QWidget)
	layout.SetSpacing(8)

	// Row 0: Golden Hour elevation | Blue Hour start elevation
	goldenLabel := qt.NewQLabel3("Golden Hour:")
	sp.goldenElevation = qt.NewQDoubleSpinBox2()
	sp.goldenElevation.SetRange(0, 15)
	sp.goldenElevation.SetSingleStep(0.5)
	sp.goldenElevation.SetSuffix("°")
	sp.goldenElevation.OnValueChanged(func(value float64) {
		sp.settings.GoldenHourElevation = value
		sp.notifyChange()
	})
	layout.AddWidget2(goldenLabel.QWidget, 0, 0)
	layout.AddWidget2(sp.goldenElevation.QWidget, 0, 1)

	blueStartLabel := qt.NewQLabel3("Blue Start:")
	sp.blueStartElevation = qt.NewQDoubleSpinBox2()
	sp.blueStartElevation.SetRange(-6, 0)
	sp.blueStartElevation.SetSingleStep(0.5)
	sp.blueStartElevation.SetSuffix("°")
	sp.blueStartElevation.OnValueChanged(func(value float64) {
		sp.settings.BlueHourStart = value
		sp.notifyChange()
	})
	layout.AddWidget2(blueStartLabel.QWidget, 0, 2)
	layout.AddWidget2(sp.blueStartElevation.QWidget, 0, 3)

	// Row 1: Blue Hour end elevation | 24-hour format checkbox
	blueEndLabel := qt.NewQLabel3("Blue End:")
	sp.blueEndElevation = qt.NewQDoubleSpinBox2()
	sp.blueEndElevation.SetRange(-18, -6)
	sp.blueEndElevation.SetSingleStep(0.5)
	sp.blueEndElevation.SetSuffix("°")
	sp.blueEndElevation.OnValueChanged(func(value float64) {
		sp.settings.BlueHourEnd = value
		sp.notifyChange()
	})
	layout.AddWidget2(blueEndLabel.QWidget, 1, 0)
	layout.AddWidget2(sp.blueEndElevation.QWidget, 1, 1)

	sp.timeFormatCheck = qt.NewQCheckBox3("24-hour format")
	sp.timeFormatCheck.OnStateChanged(func(state int) {
		sp.settings.TimeFormat24Hour = state == int(qt.Checked)
		sp.notifyChange()
	})
	layout.AddWidget3(sp.timeFormatCheck.QWidget, 1, 2, 1, 2)

	// Row 2: Auto-detect location (spans 4 columns)
	sp.autoDetectCheck = qt.NewQCheckBox3("Auto-detect location on startup")
	sp.autoDetectCheck.OnStateChanged(func(state int) {
		sp.settings.AutoDetectLocation = state == int(qt.Checked)
		sp.notifyChange()
	})
	layout.AddWidget3(sp.autoDetectCheck.QWidget, 2, 0, 1, 4)
}

// Widget returns the group box widget
func (sp *SettingsPanel) Widget() *qt.QGroupBox {
	return sp.groupBox
}

// applySettings applies settings to the UI
func (sp *SettingsPanel) applySettings(settings domain.Settings) {
	sp.goldenElevation.SetValue(settings.GoldenHourElevation)
	sp.blueStartElevation.SetValue(settings.BlueHourStart)
	sp.blueEndElevation.SetValue(settings.BlueHourEnd)

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

// GetSettings returns the current settings
func (sp *SettingsPanel) GetSettings() domain.Settings {
	return sp.settings
}

// notifyChange notifies the callback of settings changes
func (sp *SettingsPanel) notifyChange() {
	if sp.onSettingsChange != nil {
		sp.onSettingsChange(sp.settings)
	}
}
