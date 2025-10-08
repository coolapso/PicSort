package ui

import (
	"fmt"
	"image"
	"slices"
	"sync"

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
	thumbnails     *ThumbnailGridWrap
	db             *database.DB
	thumbCache     map[string]image.Image
	imagePaths     []string
	progress       *widget.ProgressBar
	progressValue  binding.Float
	progressTitle  *widget.Label
	progressFile   *widget.Label
	progressDialog dialog.Dialog
	preview        *canvas.Image
	previewCard    *widget.Card

	wg         *sync.WaitGroup
	jobs       chan string
	thumbMutex *sync.Mutex
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	return &PicsortUI{
		app:           a,
		win:           w,
		thumbCache:    make(map[string]image.Image),
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
	p.thumbnails.OnSelected = func(id widget.GridWrapItemID) {
		if id >= len(p.imagePaths) {
			return
		}

		if idx := slices.Index(p.thumbnails.selectedIDs, id); idx != -1 {
			p.thumbnails.selectedIDs = slices.Delete(p.thumbnails.selectedIDs, idx, idx+1)
		} else {
			p.thumbnails.selectedIDs = append(p.thumbnails.selectedIDs, id)
		}

		p.thumbnails.Refresh()
	}

	p.thumbnails.OnUnselected = nil

	p.thumbnails.OnHighlighted = func(id widget.GridWrapItemID) {
		if id >= len(p.imagePaths) {
			return
		}
		path := p.imagePaths[id]
		p.updatePreview(path)

		if !isExtendedSelection() {
			p.thumbnails.selectionAnchor = -1
		}

		if isExtendedSelection() {
			if p.thumbnails.selectionAnchor == -1 {
				p.thumbnails.selectionAnchor = id - 1
			}
			start, end := p.thumbnails.selectionAnchor, id
			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {
				p.thumbnails.selectedIDs = append(p.thumbnails.selectedIDs, i)
			}
			p.thumbnails.Refresh()
		}
	}

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
