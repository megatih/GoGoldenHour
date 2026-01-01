package widgets

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
	"github.com/megatih/GoGoldenHour/internal/domain"
)

// LocationPanel displays and allows editing of the current location
type LocationPanel struct {
	groupBox    *qt.QGroupBox
	searchInput *qt.QLineEdit
	searchBtn   *qt.QPushButton
	detectBtn   *qt.QPushButton
	latLabel    *qt.QLabel
	lonLabel    *qt.QLabel
	nameLabel   *qt.QLabel
	onSearch    func(query string)
	onDetect    func()
}

// NewLocationPanel creates a new location panel
func NewLocationPanel(onSearch func(query string), onDetect func()) *LocationPanel {
	lp := &LocationPanel{
		onSearch: onSearch,
		onDetect: onDetect,
	}

	lp.setupUI()
	return lp
}

// setupUI creates the location panel UI
func (lp *LocationPanel) setupUI() {
	lp.groupBox = qt.NewQGroupBox3("Location")
	layout := qt.NewQVBoxLayout(lp.groupBox.QWidget)
	layout.SetSpacing(6)

	// Search row
	searchRow := qt.NewQHBoxLayout2()
	lp.searchInput = qt.NewQLineEdit2()
	lp.searchInput.SetPlaceholderText("Search location...")
	lp.searchBtn = qt.NewQPushButton3("Go")
	lp.searchBtn.SetFixedWidth(50)

	// Connect search button click
	lp.searchBtn.OnClicked(func() {
		query := lp.searchInput.Text()
		if query != "" && lp.onSearch != nil {
			lp.onSearch(query)
		}
	})

	// Connect enter key in search input
	lp.searchInput.OnReturnPressed(func() {
		query := lp.searchInput.Text()
		if query != "" && lp.onSearch != nil {
			lp.onSearch(query)
		}
	})

	searchRow.AddWidget(lp.searchInput.QWidget)
	searchRow.AddWidget(lp.searchBtn.QWidget)
	layout.AddLayout(searchRow.QLayout)

	// Detect location button
	lp.detectBtn = qt.NewQPushButton3("Detect My Location")
	lp.detectBtn.OnClicked(func() {
		if lp.onDetect != nil {
			lp.onDetect()
		}
	})
	layout.AddWidget(lp.detectBtn.QWidget)

	// Coordinates display
	coordsLayout := qt.NewQHBoxLayout2()
	lp.latLabel = qt.NewQLabel3("Lat: --")
	lp.lonLabel = qt.NewQLabel3("Lon: --")
	coordsLayout.AddWidget(lp.latLabel.QWidget)
	coordsLayout.AddWidget(lp.lonLabel.QWidget)
	layout.AddLayout(coordsLayout.QLayout)

	// Location name
	lp.nameLabel = qt.NewQLabel3("--")
	lp.nameLabel.SetWordWrap(true)
	lp.nameLabel.SetStyleSheet("font-weight: bold; color: #ff9800;")
	layout.AddWidget(lp.nameLabel.QWidget)
}

// Widget returns the group box widget
func (lp *LocationPanel) Widget() *qt.QGroupBox {
	return lp.groupBox
}

// SetLocation updates the displayed location
func (lp *LocationPanel) SetLocation(loc domain.Location) {
	lp.latLabel.SetText(fmt.Sprintf("Lat: %.4f", loc.Latitude))
	lp.lonLabel.SetText(fmt.Sprintf("Lon: %.4f", loc.Longitude))
	lp.nameLabel.SetText(loc.Name)
}
