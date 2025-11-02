package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/mod/semver"
)

const (
	titleText      = "Welcome to Picsort!"
	welcomeMessage = "Load your pictures to get started!\n\nYou can press ? or F1 at any time to see the help menu."
)

type ghRelease struct {
	Version string `json:"tag_name"`
}

func getLatestRelease() ghRelease {
	var latestRelease ghRelease
	data, err := http.Get("https://api.github.com/repos/coolapso/picsort/releases/latest")
	if err != nil {
		return ghRelease{Version: "unknown"}
	}
	response, err := io.ReadAll(data.Body)
	if err != nil {
		return ghRelease{Version: "unknown"}
	}

	err = json.Unmarshal(response, &latestRelease)
	if err != nil {
		return ghRelease{Version: "unknown"}
	}

	return latestRelease
}

func checkNewVersion(currentVersion string) (newVersionAvailable bool, newVersion string) {
	ghRelease := getLatestRelease()
	if currentVersion == "dev" {
		return false, ""
	}

	if v := semver.Compare(currentVersion, ghRelease.Version); v == 1 {
		return true, ghRelease.Version
	}

	return false, ""
}

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

func newWelcomeScreen(v string) *fyne.Container {
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
	versionText := canvas.NewText(v, &color.Gray{Y: 0xaa})
	versionText.Alignment = fyne.TextAlignCenter
	latestVersion := canvas.NewText("", &color.RGBA{R: 0x72, G: 0x2f, B: 0x37, A: 0xff})
	latestVersion.Alignment = fyne.TextAlignCenter
	if newVersionAvailable, newVersion := checkNewVersion(v); newVersionAvailable {
		latestVersion.Text = fmt.Sprintf("New version available: %s", newVersion)
	}

	return container.NewCenter(container.NewVBox(logo, titleText, welcomeText, versionText, latestVersion))
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
