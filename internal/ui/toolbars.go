package ui

import (
	_ "embed"
	// "fmt"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newURLToolbarAction(a fyne.App, icon fyne.Resource, urlStr string) widget.ToolbarItem {
	return widget.NewToolbarAction(icon, func() {
		u, _ := url.Parse(urlStr)
		_ = a.OpenURL(u)
	})
}

func (p *PicsortUI) topBar() *fyne.Container {
	openDataSetButton := widget.NewButton("Open dataset", p.openFolderDialog)
	exportButton := widget.NewButton("Export", func() {})
	helpButton := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {})

	return container.NewBorder(nil, nil, nil, helpButton,
		container.NewHBox(openDataSetButton, exportButton),
	)
}

func (p *PicsortUI) openFolderDialog() {
	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			log.Println("Error opening folder dialog:", err)
			return
		}
		if uri == nil {
			return
		}
		go p.controller.LoadDataset(uri.Path())
	}, p.win)
	folderDialog.Resize(fyne.NewSize(800, 600)) // Set the size here
	folderDialog.Show()
}

func (p *PicsortUI) bottomBar() fyne.Widget {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			p.AddBin()
		}),

		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			p.RemoveBin()
		}),

		widget.NewToolbarSpacer(),
		newURLToolbarAction(p.app, Icons["sponsor"], "https://github.com/sponsors/coolapso"),
		newURLToolbarAction(p.app, Icons["bmc"], "https://buymeacoffee.com"),
		newURLToolbarAction(p.app, Icons["github"], "https://github.com/coolapso/picsort"),
		newURLToolbarAction(p.app, Icons["discord"], "https://discord.com"),
		newURLToolbarAction(p.app, Icons["mastodon"], "https://mastodon.social/@coolapso"),
		newURLToolbarAction(p.app, Icons["x"], "https://x.com"),
	)
}
