package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"

	// "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/data"
)

type PicsortUI struct {
	app           fyne.App
	win           fyne.Window
	bins          *fyne.Container
	thumbnails    *fyne.Container
	thumbCache    *data.ThumbnailCache
	progress      *widget.ProgressBar
	progressValue binding.Float
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	return &PicsortUI{
		app:           a,
		win:           w,
		thumbCache:    data.NewThumbnailCache(),
		progressValue: binding.NewFloat(),
	}
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
	p.progress.Hide()

	topBar := p.topBar()
	bottomBar := p.bottomBar()
	p.thumbnails = container.New(layout.NewGridLayout(3))
	thumbnailPane := container.NewScroll(p.thumbnails)
	centerContent := container.NewStack(
		thumbnailPane,
		container.NewBorder(p.progress, nil, nil, nil),
	)

	preview := widget.NewCard("Preview", "Selected image", nil)
	topSplit := container.NewHSplit(centerContent, preview)
	topSplit.SetOffset(0.3)

	mainContent := container.NewVSplit(topSplit, p.bins)
	mainContent.SetOffset(0.8)

	p.win.SetContent(container.NewBorder(topBar, bottomBar, nil, nil, mainContent))
	p.win.Resize(fyne.NewSize(1280, 720))
}
