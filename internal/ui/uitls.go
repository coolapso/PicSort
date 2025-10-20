package ui

import (
	"image/color"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	titleText      = "Welcome to Picsort!"
	welcomeMessage = "Load your pictures to get started!\n\nYou can press ? or F1 at any time to see the help menu."
)

func shiftPressed() bool {
	return fyne.CurrentApp().Driver().(desktop.Driver).CurrentKeyModifiers() == 1
}

func translateKey(key *fyne.KeyEvent) *fyne.KeyEvent {
	translatedKey := *key
	switch key.Name {
	case fyne.KeyH:
		translatedKey.Name = fyne.KeyLeft
	case fyne.KeyJ:
		translatedKey.Name = fyne.KeyDown
	case fyne.KeyK:
		translatedKey.Name = fyne.KeyUp
	case fyne.KeyL:
		translatedKey.Name = fyne.KeyRight
	default:
	}

	return &translatedKey
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

func newHelpDialogContent() fyne.CanvasObject {
	grid := container.NewGridWithColumns(2)
	addRow := func(key, description string) {
		keyLabel := widget.NewLabel(key)
		keyLabel.TextStyle.Bold = true
		grid.Add(keyLabel)
		grid.Add(widget.NewLabel(description))
	}

	ghLink, err := url.Parse("https://github.com/coolapso/picsort")
	if err != nil {
		log.Println("Error parsing URL for help dialog:", err)
	}

	link := widget.NewHyperlink("Chekcout our Github repository", ghLink)
	link.Alignment = fyne.TextAlignCenter

	addRow("?, F1", "Show this help dialog")
	addRow("Ctrl+O", "Open dataset folder")
	addRow("Ctrl+E", "Export dataset")
	addRow("Ctrl+T", "Add a new bin")
	addRow("Ctrl+W", "Remove the last bin")
	addRow("Ctrl+0-9", "Switch to the corresponding bin tab")
	addRow("Ctrl+H/L", "Adjust preview panel size")
	addRow("H,J,K,L / Arrow Keys", "Navigate through images")
	addRow("Space", "Select / Unselect image")
	addRow("Shift + H,J,K,L / Arrow Keys", "Select multiple images")
	addRow("Escape", "Unselect all selected images")
	addRow("0 - 9", "Move selected image(s) to bin")

	globalTitle := widget.NewLabel("Picsort keyboard shortcuts")
	globalTitle.TextStyle.Bold = true
	imageGridTitle := widget.NewLabel("Image Grid")
	imageGridTitle.TextStyle.Bold = true

	moreDetailsText := widget.NewLabel("Not what you're looking for?")
	moreDetailsText.TextStyle.Bold = true
	more := container.NewHBox(moreDetailsText, link)

	return container.NewVBox(
		globalTitle,
		grid,
		more,
	)
}
