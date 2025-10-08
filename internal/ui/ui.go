package ui

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	// "sync"

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
	progress       *widget.ProgressBar
	progressValue  binding.Float
	progressTitle  *widget.Label
	progressFile   *widget.Label
	progressDialog dialog.Dialog
	preview        *canvas.Image
	previewCard    *widget.Card
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	p := &PicsortUI{
		app:           a,
		win:           w,
		progressValue: binding.NewFloat(),
		progressTitle: widget.NewLabel(""),
		progressFile:  widget.NewLabel(""),
	}
	p.controller = controller.New(p)
	return p
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

func (p *PicsortUI) GetWindow() fyne.Window { return p.win }

func (p *PicsortUI) UpdatePreview(path string) {
	go func() {
		file, err := os.Open(path)
		if err != nil {
			log.Printf("could not open file for preview %s: %v", path, err)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			log.Printf("could not decode image for preview %s: %v", path, err)
			return
		}

		fyne.Do(func() {
			p.preview.Image = img
			p.preview.Refresh()
			p.previewCard.SetSubTitle(filepath.Base(path))
		})
	}()
}

func (p *PicsortUI) sortingBins() {
	p.bins = container.New(layout.NewGridLayout(5))
	for i := 1; i <= 5; i++ {
		p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", i), "", nil))
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
	p.thumbnails = NewThumbnailGrid(p.controller)

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
