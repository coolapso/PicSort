package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/database"
)

type PicsortUI struct {
	app            fyne.App
	win            fyne.Window
	bins           *fyne.Container
	thumbnails     *widget.GridWrap
	db             *database.DB
	imagePaths     []string
	progress       *widget.ProgressBar
	progressValue  binding.Float
	progressTitle  *widget.Label
	progressFile   *widget.Label
	progressDialog dialog.Dialog
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	return &PicsortUI{
		app:           a,
		win:           w,
		progressValue: binding.NewFloat(),
		progressTitle: widget.NewLabel(""),
		progressFile:  widget.NewLabel(""),
	}
}

func (p *PicsortUI) sortingBins() {
	p.bins = container.New(layout.NewGridLayout(5))
	for i := 1; i <= 5; i++ {
		p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", i), "", nil))
	}
}

func (p *PicsortUI) NewThumbnailGrid() *widget.GridWrap {
	return widget.NewGridWrap(
		func() int {
			return len(p.imagePaths)
		},

		func() fyne.CanvasObject {
			img := canvas.NewImageFromImage(nil)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(200, 200))
			return img
		},

		func(i widget.GridWrapItemID, o fyne.CanvasObject) {
			if i >= len(p.imagePaths) {
				return
			}
			path := p.imagePaths[i]
			img := o.(*canvas.Image)
			thumb, _ := p.db.GetThumbnail(path)
			img.Image = thumb
			img.Refresh()
		},
	)
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
	p.thumbnails = p.NewThumbnailGrid()
	centerContent := container.NewBorder(nil, nil, nil, nil, p.thumbnails)

	preview := widget.NewCard("Preview", "Selected image", nil)
	topSplit := container.NewHSplit(centerContent, preview)
	topSplit.SetOffset(0.3)

	mainContent := container.NewVSplit(topSplit, p.bins)
	mainContent.SetOffset(0.8)

	p.win.SetContent(container.NewBorder(topBar, bottomBar, nil, nil, mainContent))
	p.win.Resize(fyne.NewSize(1280, 720))
}
