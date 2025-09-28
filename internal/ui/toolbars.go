package ui

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	openDataSetButton := widget.NewButton("Open dataset", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Println("Error opening folder dialog:", err)
				return
			}
			if uri == nil {
				return
			}
			go p.loadThumbnails(uri.Path())
		}, p.win)
		folderDialog.Resize(fyne.NewSize(800, 600)) // Set the size here
		folderDialog.Show()
	})

	exportButton := widget.NewButton("Export", func() {})

	helpButton := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {})

	return container.NewBorder(nil, nil, nil, helpButton,
		container.NewHBox(openDataSetButton, exportButton),
	)
}

func (p *PicsortUI) bottomBar() fyne.Widget {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			if len(p.bins.Objects) <= 9 {
				binCount := len(p.bins.Objects) + 1
				p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", binCount), "", nil))
				p.bins.Layout = layout.NewGridLayout(binCount)
				p.bins.Refresh()
			}
		}),

		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			if len(p.bins.Objects) > 1 {
				binCount := len(p.bins.Objects) - 1
				p.bins.Remove(p.bins.Objects[binCount])
				p.bins.Layout = layout.NewGridLayout(binCount)
				p.bins.Refresh()
			}
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
