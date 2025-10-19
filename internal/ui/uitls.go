package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

func shiftPressed() bool {
	return fyne.CurrentApp().Driver().(desktop.Driver).CurrentKeyModifiers() == 1
}

func newWelcomeScreen() *fyne.Container {
	logo := canvas.NewImageFromResource(LogoIcon)
	logo.SetMinSize(fyne.NewSize(192, 192))
	logo.FillMode = canvas.ImageFillContain
	titleText := canvas.NewText(titleText, color.White)
	titleText.Alignment = fyne.TextAlignCenter
	titleText.TextStyle.Bold = true
	titleText.TextSize = 30
	titleText.TextStyle.Monospace = true
	welcomeText := widget.NewLabel(welcomeMessage)
	welcomeText.Alignment = fyne.TextAlignCenter
	welcomeText.TextStyle.Monospace = true

	return container.NewCenter(container.NewVBox(logo, titleText, welcomeText))
}
