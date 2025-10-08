package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func isExtendedSelection() bool {
	return fyne.CurrentApp().Driver().(desktop.Driver).CurrentKeyModifiers() == 1
}


