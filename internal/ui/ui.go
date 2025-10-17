package ui

import (
	"fmt"
	"image"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/controller"
)

type PicsortUI struct {
	app        fyne.App
	win        fyne.Window
	controller *controller.Controller

	layout         string
	tabs           *container.AppTabs
	binGrids       map[int]*ThumbnailGridWrap
	progress       *widget.ProgressBar
	progressValue  binding.Float
	progressTitle  *widget.Label
	progressFile   *widget.Label
	progressDialog dialog.Dialog
	preview        *canvas.Image
	previewCard    *widget.Card
}

func (p *PicsortUI) ShowProgressDialog(msg string) {
	fyne.Do(func() {
		p.progressTitle.SetText(msg)
		p.progressDialog.Show()
		p.progress.Show()
		p.progressValue.Set(0)
	})
}

func (p *PicsortUI) SetProgress(progress float64, f string) {
	fyne.Do(func() {
		p.progressFile.SetText(f)
		p.progressValue.Set(progress)
	})
}

func (p *PicsortUI) HideProgressDialog() {
	fyne.Do(func() {
		p.progressDialog.Hide()
	})
}

func (p *PicsortUI) ShowErrorDialog(err error) {
	fyne.Do(func() {
		d := dialog.NewError(err, p.win)
		d.SetOnClosed(func() {
			p.progressDialog.Hide()
		})
		d.Show()
	})
}

func (p *PicsortUI) ReloadAll() {
	fyne.Do(func() {
		for _, bin := range p.binGrids {
			p.ReloadBin(bin.id)
		}
	})
}

func (p *PicsortUI) ReloadBin(id int) {
	fyne.Do(func() {
		p.binGrids[id].unselectAll()
		p.binGrids[id].Reload()
		p.setTabTitle(id)
		p.tabs.Refresh()
	})
}

func (p *PicsortUI) GoToTab(id int) {
	fyne.Do(func() {
		if id < len(p.tabs.Items) {
			p.tabs.SelectIndex(id)
			p.win.Canvas().Focus(p.binGrids[id])
		}
	})
}

func (p *PicsortUI) setTabTitle(id int) {
	tabTitle := fmt.Sprintf("Bin %d", id)
	if id == 0 {
		tabTitle = "To Sort"
	}

	if p.binGrids[id].itemCount() > 0 {
		tabTitle = fmt.Sprintf("Bin %d (%d)", id, p.binGrids[id].itemCount())
		if id == 0 {
			tabTitle = fmt.Sprintf("To Sort (%d)", p.binGrids[id].itemCount())
		}
	}
	p.tabs.Items[id].Text = tabTitle
}

func (p *PicsortUI) GetWindow() fyne.Window { return p.win }

func (p *PicsortUI) UpdatePreview(i image.Image, path string) {
	fyne.Do(func() {
		p.preview.Image = i
		p.preview.Refresh()
		p.previewCard.SetSubTitle(filepath.Base(path))
	})
}

func (p *PicsortUI) initBins() {
	for i := 0; i <= 5; i++ {
		p.NewBin()
	}

}

func (p *PicsortUI) NewBin() {
	if len(p.binGrids) <= 9 {
		binCount := len(p.binGrids)
		binGrid := NewThumbnailGrid(binCount, p.controller)
		p.binGrids[binCount] = binGrid
		p.tabs.Append(container.NewTabItem("", p.binGrids[binCount]))
		p.setTabTitle(binCount)
	}
}

func (p *PicsortUI) RemoveBin() {
	if len(p.binGrids) > 1 {
		idToRemove := len(p.binGrids) - 1
		delete(p.binGrids, idToRemove)
		p.tabs.Remove(p.tabs.Items[idToRemove])
	}
}

func (p *PicsortUI) GetBinCount() int {
	return len(p.binGrids)
}

func (p *PicsortUI) globalKeyBinds() {
	ctrlT := &desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl}
	p.win.Canvas().AddShortcut(ctrlT, func(shortcut fyne.Shortcut) {
		p.GoToTab(0)
	})

	binKeys := []fyne.KeyName{
		fyne.Key0, fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5,
		fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9,
	}

	for i, key := range binKeys {
		shortcut := &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierControl}
		p.win.Canvas().AddShortcut(shortcut, func(s fyne.Shortcut) {
			p.GoToTab(i)
		})
	}

	ctrlO := &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	p.win.Canvas().AddShortcut(ctrlO, func(s fyne.Shortcut) {
		p.openFolderDialog()
	})
}

func (p *PicsortUI) Build() {
	p.globalKeyBinds()
	p.progress = widget.NewProgressBarWithData(p.progressValue)
	progressContent := container.NewVBox(
		p.progressTitle,
		p.progress,
		p.progressFile,
	)
	p.progressDialog = dialog.NewCustomWithoutButtons(
		"Preparing dataset...",
		progressContent,
		p.win,
	)
	p.progressDialog.Resize(fyne.NewSize(500, 150))

	topBar := p.topBar()
	bottomBar := p.bottomBar()
	p.tabs = container.NewAppTabs()
	p.initBins()

	p.preview = canvas.NewImageFromImage(nil)
	p.preview.FillMode = canvas.ImageFillContain
	p.previewCard = widget.NewCard("Preview", "Selected image", p.preview)
	mainContent := container.NewHSplit(p.tabs, p.previewCard)
	mainContent.SetOffset(0.3)

	p.win.SetContent(container.NewBorder(topBar, bottomBar, nil, nil, mainContent))
	p.win.Resize(fyne.NewSize(1280, 720))
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	p := &PicsortUI{
		app:           a,
		win:           w,
		progressValue: binding.NewFloat(),
		progressTitle: widget.NewLabel(""),
		progressFile:  widget.NewLabel(""),
		binGrids:      make(map[int]*ThumbnailGridWrap),
	}
	p.controller = controller.New(p)
	return p
}
