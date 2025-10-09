package ui

import (
	"fmt"
	"image"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
	})
}

func (p *PicsortUI) FocusThumbnails() {
	fyne.Do(func() {
		p.win.Canvas().Focus(p.thumbnails)
	})
}

func (p *PicsortUI) FocusBin(index int) {
	fyne.Do(func() {
		if index < 1 || index > len(p.bins.Objects) {
			return
		}
		gridToFocus := p.binGrids[index-1]
		if gridToFocus != nil {
			p.win.Canvas().Focus(gridToFocus)
		}
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
		// binGrid := NewThumbnailGridWrap(p.controller)
		// p.binGrids[i] = binGrid
		// card := widget.NewCard(fmt.Sprintf("Bin %d", i), "", NewThumbnailGrid(p.controller))
		// p.bins.Add(card)
		p.AddBin()
	}
}

func (p *PicsortUI) AddBin() {
	if len(p.bins.Objects) <= 9 {
		binCount := len(p.bins.Objects) + 1
		binGrid := NewThumbnailGrid(binCount, p.controller)
		p.binGrids[binCount] = binGrid
		card := widget.NewCard(fmt.Sprintf("Bin %d", binCount), "", nil)
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

func (p *PicsortUI) Build() {
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
