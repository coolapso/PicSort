package ui

import (
	_ "embed"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newURLToolbarAction(a fyne.App, icon fyne.Resource, urlStr string) widget.ToolbarItem {
	return widget.NewToolbarAction(icon, func() {
		u, _ := url.Parse(urlStr)
		_ = a.OpenURL(u)
	})
}

func (p *PicsortUI) setTopBar() {
	openDataSetButton := widget.NewButton("Open dataset", p.openFolderDialog)
	exportButton := widget.NewButton("Export", func() {})
	helpButton := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {})

	p.topBar = container.NewBorder(nil, nil, nil, helpButton,
		container.NewHBox(openDataSetButton, exportButton),
	)
}

func (p *PicsortUI) setBottomBar() {
	p.addBinButton = widget.NewToolbarAction(theme.ContentAddIcon(), func() {
		p.NewBin()
	})
	p.addBinButton.ToolbarObject().Hide()

	p.rmBinButton = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
		p.RemoveBin()
	})
	p.rmBinButton.ToolbarObject().Hide()

	p.bottomBar = widget.NewToolbar(
		p.addBinButton,
		p.rmBinButton,
		widget.NewToolbarSpacer(),
		newURLToolbarAction(p.app, Icons["logo"], "https://picsort.coolapso.sh"),
		newURLToolbarAction(p.app, Icons["sponsor"], "https://github.com/sponsors/coolapso"),
		newURLToolbarAction(p.app, Icons["bmc"], "https://buymeacoffee.com"),
		newURLToolbarAction(p.app, Icons["github"], "https://github.com/coolapso/picsort"),
		newURLToolbarAction(p.app, Icons["discord"], "https://discord.com"),
		newURLToolbarAction(p.app, Icons["mastodon"], "https://mastodon.social/@coolapso"),
		newURLToolbarAction(p.app, Icons["x"], "https://x.com"),
	)
}
