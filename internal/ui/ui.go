package ui

import (
	"fmt"
	"image"
	"slices"

	// "image/color"
	// "reflect"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"

	// "fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/database"
)

type PicsortUI struct {
	app  fyne.App
	win  fyne.Window
	bins *fyne.Container
	// thumbnails     *widget.GridWrap
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

type ThumbnailGridWrap struct {
	widget.GridWrap
	selectionAnchor widget.GridWrapItemID
	selectedPaths   []string
	OnNavigated     func(widget.GridWrapItemID)
}

func NewThumbnailGridWrap(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.GridWrapItemID, fyne.CanvasObject)) *ThumbnailGridWrap {
	grid := &ThumbnailGridWrap{
		selectionAnchor: -1,
		selectedPaths:   []string{},
	}
	grid.Length = length
	grid.CreateItem = createItem
	grid.UpdateItem = updateItem
	grid.ExtendBaseWidget(grid)
	return grid
}

// Workaround for fyne main branch without OnNavigated
// func (t *ThumbnailGridWrap) getScrolledID() widget.GridWrapItemID {
// 	v := reflect.ValueOf(t).Elem().FieldByName("GridWrap")
// 	if !v.IsValid() {
// 		return -1
// 	}
//
// 	field := v.FieldByName("currentFocus")
// 	if !field.IsValid() {
// 		return -1
// 	}
//
// 	field = reflect.NewAt(field.Type(), field.Addr().UnsafePointer()).Elem()
//
// 	if id, ok := field.Interface().(widget.GridWrapItemID); ok {
// 		return id
// 	}
// 	return -1
// }

func (g *ThumbnailGridWrap) TypedKey(key *fyne.KeyEvent) {
	// beforeID := g.getScrolledID()
	translatedKey := *key
	switch translatedKey.Name {
	case fyne.KeyH:
		translatedKey.Name = fyne.KeyLeft
	case fyne.KeyJ:
		translatedKey.Name = fyne.KeyDown
	case fyne.KeyK:
		translatedKey.Name = fyne.KeyUp
	case fyne.KeyL:
		translatedKey.Name = fyne.KeyRight
	case fyne.KeyEscape:
		g.UnselectAll()
	}

	g.GridWrap.TypedKey(&translatedKey)

	// afterID := g.getScrolledID()

	// if g.OnNavigated != nil && beforeID != afterID && afterID != -1 {
	// 	g.OnNavigated(afterID)
	// }
}

func (p *PicsortUI) sortingBins() {
	p.bins = container.New(layout.NewGridLayout(5))
	for i := 1; i <= 5; i++ {
		p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", i), "", nil))
	}
}
func (p *PicsortUI) NewThumbnailGrid() *ThumbnailGridWrap {
	return NewThumbnailGridWrap(
		func() int {
			return len(p.imagePaths)
		},
		func() fyne.CanvasObject {
			img := canvas.NewImageFromImage(nil)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(200, 200))
			return container.NewStack(img)
		},
		func(i widget.GridWrapItemID, o fyne.CanvasObject) {
			if i >= len(p.imagePaths) {
				return
			}
			thumb := o.(*fyne.Container)
			img := thumb.Objects[0].(*canvas.Image)

			path := p.imagePaths[i]
			p.thumbMutex.Lock()
			defer p.thumbMutex.Unlock()
			if thumb, ok := p.thumbCache[path]; ok {
				img.Image = thumb
			} else {
				if t, found := p.db.GetThumbnail(path); found {
					img.Image = t
				}
			}
			img.Refresh()
		},
	)
}

func isExtendedSelection() bool {
	if fyne.CurrentApp().Driver().(desktop.Driver).CurrentKeyModifiers() == 1 {
		return true
	}

	return false
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
		path := p.imagePaths[id]

		if slices.Contains(p.thumbnails.selectedPaths, p.imagePaths[id]) {
			return
		}

		if isExtendedSelection() {
			if p.thumbnails.selectionAnchor == -1 {
				p.thumbnails.selectionAnchor = id
			}
			start, end := p.thumbnails.selectionAnchor, id
			if start > end {
				start, end = end, start
			}

			p.thumbnails.selectedPaths = []string{}
			for i := start; i <= end; i++ {
				p.thumbnails.selectedPaths = append(p.thumbnails.selectedPaths, p.imagePaths[i])
				p.thumbnails.Select(i)
			}
			return
		}
		p.thumbnails.selectionAnchor = id
		p.thumbnails.selectedPaths = []string{path}
	}

	p.thumbnails.OnHighlighted = func(id widget.GridWrapItemID) {
		if id >= len(p.imagePaths) {
			return
		}
		path := p.imagePaths[id]
		p.updatePreview(path)
		if isExtendedSelection() {
			p.thumbnails.Select(id)
		}
	}

	p.thumbnails.OnHovered = func(id widget.GridWrapItemID) {
		// fmt.Println("highlighted:", id)
		path := p.imagePaths[id]
		p.updatePreview(path)
		if isExtendedSelection() {
			p.thumbnails.Select(id)
		}
	}

	// OnHighlighted is on a custom branch of my fork, not yet on upstream.
	// This is the alternative in case it doesn't get merged.
	// p.thumbnails.OnNavigated = func(id widget.GridWrapItemID) {
	// 	path := p.imagePaths[id]
	// 	p.updatePreview(path)
	// 	if isExtendedSelection() {
	// 		if p.selectionAnchor == -1 {
	// 			p.selectionAnchor = id
	// 			p.thumbnails.Select(id)
	// 			p.selectedPaths = append(p.selectedPaths, path)
	// 			return
	// 		}
	// 		p.thumbnails.Select(id)
	// 	}
	// }

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
