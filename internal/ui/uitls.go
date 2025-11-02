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

	if v := semver.Compare(ghRelease.Version, currentVersion); v == 1 {
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

func newHelpSection(shortcuts map[string]string) *fyne.Container {
	grid := container.NewGridWithColumns(2)
	for key, description := range shortcuts {
		keyLabel := widget.NewLabel(key)
		keyLabel.TextStyle.Bold = true
		grid.Add(keyLabel)
		grid.Add(widget.NewLabel(description))
	}

	return grid
}

func newHelpDialogContent() fyne.CanvasObject {
	ghLink, err := url.Parse("https://github.com/coolapso/picsort")
	if err != nil {
		log.Println("Error parsing URL for help dialog:", err)
	}
	link := widget.NewHyperlink("Checkout our Github repository", ghLink)
	link.Alignment = fyne.TextAlignCenter

	globalShortcuts := map[string]string{
		"?, F1":    "Show this help dialog",
		"Ctrl+O":   "Open dataset folder",
		"Ctrl+E":   "Export dataset",
		"Ctrl+T":   "Add a new bin",
		"Ctrl+W":   "Remove the last bin",
		"Ctrl+0-9": "Switch to the corresponding bin tab",
		"Ctrl+H/L": "Adjust preview panel size",
	}

	movementShortcuts := map[string]string{
		"H,J,K,L / Arrow Keys": "Navigate through images",
		"T":                    "Go to the first visible image",
		"B":                    "Go to the last visible image",
		"M":                    "Go to the middle image",
		"MM":                   "Go to the middle of the dataset",
		"Shift + M":            "Go to the middle of the dataset",
		"PgDown, D":            "Scroll down to next set of images",
		"PgUp,U":               "Scroll Up to next set of images",
		"Home, GG":             "Go got top of the dataset",
		"End, Shift + GG":      "Go to the bottom of the dataset",
	}

	selectionShortcuts := map[string]string{
		"Space":                        "Select / Unselect image",
		"Shift + H,J,K,L / Arrow Keys": "Select multiple images",
		"Escape":                       "Unselect all selected images",
		"0 - 9":                        "Move selected image(s) to bin",
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Global", container.NewScroll(newHelpSection(globalShortcuts))),
		container.NewTabItem("Movement", container.NewScroll(newHelpSection(movementShortcuts))),
		container.NewTabItem("Selection", container.NewScroll(newHelpSection(selectionShortcuts))),
	)

	moreDetailsText := widget.NewLabel("Not what you're looking for?")
	moreDetailsText.TextStyle.Bold = true
	more := container.NewHBox(moreDetailsText, link)

	borderContent := container.NewBorder(nil, container.NewVBox(more, widget.NewSeparator()), nil, nil, tabs)
	sizer := canvas.NewRectangle(color.Transparent)
	sizer.SetMinSize(fyne.NewSize(700, 400))

	return container.NewStack(sizer, borderContent)
}
