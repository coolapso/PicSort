package ui

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

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
	thumbCache     map[string]image.Image
	imagePaths     []string
	progress       *widget.ProgressBar
	progressValue  binding.Float
	progressTitle  *widget.Label
	progressFile   *widget.Label
	progressDialog dialog.Dialog
	focusedThumbID widget.GridWrapItemID
	preview        *canvas.Image
	previewCard    *widget.Card

	wg         *sync.WaitGroup
	jobs       chan string
	thumbMutex *sync.Mutex

	shiftPressed    bool
	selectionAnchor widget.GridWrapItemID
	selectedIndices map[widget.GridWrapItemID]struct{}
}

func New(a fyne.App, w fyne.Window) *PicsortUI {
	return &PicsortUI{
		app:             a,
		win:             w,
		thumbCache:      make(map[string]image.Image),
		progressValue:   binding.NewFloat(),
		progressTitle:   widget.NewLabel(""),
		progressFile:    widget.NewLabel(""),
		selectionAnchor: -1,
		selectedIndices: make(map[widget.GridWrapItemID]struct{}),
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
			bg := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
			img := canvas.NewImageFromImage(nil)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(200, 200))
			return container.NewMax(bg, img)
		},
		func(i widget.GridWrapItemID, o fyne.CanvasObject) {
			if i >= len(p.imagePaths) {
				return
			}
			c := o.(*fyne.Container)
			bg := c.Objects[0].(*canvas.Rectangle)
			img := c.Objects[1].(*canvas.Image)

			path := p.imagePaths[i]
			if thumb, ok := p.thumbCache[path]; ok {
				img.Image = thumb
			} else {
				if t, found := p.db.GetThumbnail(path); found {
					img.Image = t
				}
			}

			switch {
			case p.isSelected(i):
				bg.FillColor = theme.SelectionColor()
			case i == p.focusedThumbID:
				bg.FillColor = color.NRGBA{R: 50, G: 50, B: 50, A: 90}
			default:
				bg.FillColor = color.NRGBA{0, 0, 0, 0}
			}
			bg.Refresh()
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
	p.thumbnails.OnSelected = func(id widget.GridWrapItemID) {
		p.handleClickSelect(id)
	}

	centerContent := container.NewBorder(nil, nil, nil, nil, p.thumbnails)
	p.preview = canvas.NewImageFromImage(nil)
	p.preview.FillMode = canvas.ImageFillContain
	p.previewCard = widget.NewCard("Preview", "Selected image", p.preview)
	topSplit := container.NewHSplit(centerContent, p.previewCard)
	topSplit.SetOffset(0.3)

	mainContent := container.NewVSplit(topSplit, p.bins)
	mainContent.SetOffset(0.8)

	p.win.SetContent(container.NewBorder(topBar, bottomBar, nil, nil, mainContent))
	p.win.Canvas().SetOnTypedKey(p.navigation)
	p.win.Resize(fyne.NewSize(1280, 720))
}

func (p *PicsortUI) navigation(e *fyne.KeyEvent) {
	fmt.Println(e.Name)
	var extended bool
	if fyne.CurrentApp().Driver().(desktop.Driver).CurrentKeyModifiers() == 1 {
		extended = true
	}

	if len(p.imagePaths) == 0 {
		return
	}

	newID := max(p.focusedThumbID, 0)

	switch e.Name {
	case fyne.KeyH:
		if newID > 0 {
			newID--
		}
	case fyne.KeyL:
		if newID < widget.GridWrapItemID(len(p.imagePaths))-1 {
			newID++
		}
	case fyne.KeyK:
		cols := p.visibleCols()
		if cols > 0 && newID-cols >= 0 {
			newID -= cols
		}
	case fyne.KeyJ:
		cols := p.visibleCols()
		if cols > 0 && newID+cols < widget.GridWrapItemID(len(p.imagePaths)) {
			newID += cols
		}
	default:
		return
	}

	if newID == p.focusedThumbID {
		return
	}

	p.focusedThumbID = newID

	if !extended {
		p.clearSelection()
		p.addSelection(newID)
		p.selectionAnchor = newID
	} else {
		if p.selectionAnchor == -1 {
			p.selectionAnchor = newID
			p.clearSelection()
			p.addSelection(newID)
		} else {
			p.clearSelection()
			start, end := p.selectionAnchor, newID
			if start > end {
				start, end = end, start
			}
			for i := start; i <= end; i++ {
				p.addSelection(i)
			}
		}
	}

	p.updatePreview()
	p.thumbnails.Refresh()
}

func (p *PicsortUI) visibleCols() widget.GridWrapItemID {
	cellW := 200 + theme.Padding()
	totalW := p.thumbnails.Size().Width
	if totalW <= 0 {
		return 1
	}
	cols := max(int(totalW/float32(cellW)), 1)

	return widget.GridWrapItemID(cols)
}

func (p *PicsortUI) isSelected(id widget.GridWrapItemID) bool {
	_, ok := p.selectedIndices[id]
	return ok
}

func (p *PicsortUI) addSelection(id widget.GridWrapItemID) {
	p.selectedIndices[id] = struct{}{}
}

func (p *PicsortUI) clearSelection() {
	if len(p.selectedIndices) == 0 {
		return
	}
	for k := range p.selectedIndices {
		delete(p.selectedIndices, k)
	}
}

func (p *PicsortUI) handleClickSelect(id widget.GridWrapItemID) {
	if id < 0 || id >= widget.GridWrapItemID(len(p.imagePaths)) {
		return
	}

	p.clearSelection()
	p.focusedThumbID = id
	if !p.shiftPressed || p.selectionAnchor == -1 {
		p.selectionAnchor = id
		p.addSelection(id)
	}

	start, end := p.selectionAnchor, id
	if start > end {
		start, end = end, start
	}
	for i := start; i <= end; i++ {
		p.addSelection(i)
	}

	p.updatePreview()
	p.win.Canvas().Focus(p.thumbnails)
}
