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
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/controller"
)

type PicsortUI struct {
	app        fyne.App
	win        fyne.Window
	controller *controller.Controller

	bins           *fyne.Container
	thumbnails     *ThumbnailGridWrap
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
		dialog.ShowError(err, p.win)
	})
}

func (p *PicsortUI) ReloadAll() {
	fyne.Do(func() {
		p.thumbnails.Reload()
		for _, bin := range p.binGrids {
			bin.Reload()
		}
	})
}

func (p *PicsortUI) ReloadBin(id int) {
	fyne.Do(func() {
		if id == 0 {
			p.thumbnails.Reload()
			return
		}
		p.binGrids[id].Reload()
	})
}

func (p *PicsortUI) FocusThumbnails(id int) {
	fyne.Do(func() {
		if id == 0 {
			p.win.Canvas().Focus(p.thumbnails)
			return
		}

		p.win.Canvas().Focus(p.binGrids[id])
	})
}

func (p *PicsortUI) GetWindow() fyne.Window { return p.win }

func (p *PicsortUI) UpdatePreview(i image.Image, path string) {
	fyne.Do(func() {
		p.preview.Image = i
		p.preview.Refresh()
		p.previewCard.SetSubTitle(filepath.Base(path))
	})
}

func (p *PicsortUI) sortingBins() {
	p.bins = container.New(layout.NewGridLayout(5))
	for i := 1; i <= 5; i++ {
		p.AddBin()
	}
}

func (p *PicsortUI) AddBin() {
	if len(p.bins.Objects) <= 9 {
		binCount := len(p.bins.Objects) + 1
		binGrid := NewThumbnailGrid(binCount, p.controller)
		p.binGrids[binCount] = binGrid
		card := widget.NewCard(fmt.Sprintf("Bin %d", binCount), "", binGrid)
		p.bins.Add(card)
		p.bins.Layout = layout.NewGridLayout(binCount)
		p.bins.Refresh()
	}
}

func (p *PicsortUI) RemoveBin() {
	if len(p.bins.Objects) > 1 {
		binCount := len(p.bins.Objects)
		delete(p.binGrids, binCount)
		binCount = len(p.bins.Objects) - 1
		p.bins.Remove(p.bins.Objects[binCount])
		p.bins.Layout = layout.NewGridLayout(binCount)
		p.bins.Refresh()
	}
}

func (p *PicsortUI) globalKeyBinds() {
	ctrlT := &desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl}
	p.win.Canvas().AddShortcut(ctrlT, func(shortcut fyne.Shortcut) {
		p.FocusThumbnails(0)
	})

	binKeys := []fyne.KeyName{
		fyne.Key0, fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5,
		fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9,
	}

	for i, key := range binKeys {
		shortcut := &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierControl}
		p.win.Canvas().AddShortcut(shortcut, func(s fyne.Shortcut) {
			p.FocusThumbnails(i)
		})
	}

	ctrlO := &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	p.win.Canvas().AddShortcut(ctrlO, func(s fyne.Shortcut) {
		p.openFolderDialog()
	})
}

func (p *PicsortUI) Build() {
	p.globalKeyBinds()
	p.sortingBins()
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
	p.thumbnails = NewThumbnailGrid(0, p.controller)

	p.preview = canvas.NewImageFromImage(nil)
	p.preview.FillMode = canvas.ImageFillContain
	p.previewCard = widget.NewCard("Preview", "Selected image", p.preview)
	topSplit := container.NewHSplit(p.thumbnails, p.previewCard)
	topSplit.SetOffset(0.3)

	mainContent := container.NewVSplit(topSplit, p.bins)
	mainContent.SetOffset(0.8)

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
