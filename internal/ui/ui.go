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
	"fyne.io/fyne/v2/theme"

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
	selectedIDs     []widget.GridWrapItemID
	OnNavigated     func(widget.GridWrapItemID)
}

func NewThumbnailGridWrap(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.GridWrapItemID, fyne.CanvasObject)) *ThumbnailGridWrap {
	grid := &ThumbnailGridWrap{
		selectionAnchor: -1,
		selectedIDs:     []widget.GridWrapItemID{},
	}
	grid.Length = length
	grid.CreateItem = createItem
	grid.UpdateItem = updateItem
	grid.ExtendBaseWidget(grid)
	return grid
}

// Workaround for fyne main branch without OnHighlighted
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
		g.unselectAll()
	}

	g.GridWrap.TypedKey(&translatedKey)
}

func (g *ThumbnailGridWrap) unselectAll() {
	g.selectedIDs = []widget.GridWrapItemID{}
	g.selectionAnchor = -1
	g.Refresh()
}

func (p *PicsortUI) sortingBins() {
	p.bins = container.New(layout.NewGridLayout(5))
	for i := 1; i <= 5; i++ {
		p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", i), "", nil))
	}
}

type ImageCheck struct {
	widget.BaseWidget
	Image     image.Image
	Checked   bool
	OnChanged func(bool)
}

func (ic *ImageCheck) CreateRenderer() fyne.WidgetRenderer {
	r := &imageCheckRenderer{
		imageCheck: ic,
		thumb:      canvas.NewImageFromImage(ic.Image),
		checkIcon:  canvas.NewImageFromResource(theme.CheckButtonIcon()),
	}
	r.thumb.FillMode = canvas.ImageFillContain
	r.checkIcon.Hide()
	return r
}

type imageCheckRenderer struct {
	imageCheck *ImageCheck
	thumb      *canvas.Image
	checkIcon  *canvas.Image
}

func (r *imageCheckRenderer) Layout(size fyne.Size) {
	r.thumb.Resize(size)
	r.checkIcon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
	r.checkIcon.Move(fyne.NewPos(size.Width-theme.IconInlineSize()-theme.Padding(), theme.Padding()))
}

func (r *imageCheckRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 200)
}

func (r *imageCheckRenderer) Refresh() {
	r.thumb.Image = r.imageCheck.Image
	r.thumb.Refresh()
	if r.imageCheck.Checked {
		r.checkIcon.Resource = theme.CheckButtonCheckedIcon()
		r.checkIcon.Show()
	} else {
		r.checkIcon.Resource = theme.CheckButtonIcon()
		r.checkIcon.Hide()
	}
	r.checkIcon.Refresh()
	canvas.Refresh(r.imageCheck)
}

func (r *imageCheckRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.thumb, r.checkIcon}
}

func (r *imageCheckRenderer) Destroy() {}

func NewImageCheck(img image.Image, onChanged func(bool)) *ImageCheck {
	ic := &ImageCheck{
		Image:     img,
		OnChanged: onChanged,
	}

	ic.ExtendBaseWidget(ic)
	return ic
}

func (p *PicsortUI) NewThumbnailGrid() *ThumbnailGridWrap {
	return NewThumbnailGridWrap(
		func() int {
			return len(p.imagePaths)
		},
		func() fyne.CanvasObject {
			// img := canvas.NewImageFromImage(nil)
			// img.FillMode = canvas.ImageFillContain
			// img.SetMinSize(fyne.NewSize(200, 200))
			// bg := canvas.NewRectangle(color.Transparent)
			// return container.NewStack(bg, img)
			return NewImageCheck(nil, nil)
		},
		func(i widget.GridWrapItemID, o fyne.CanvasObject) {
			if i >= len(p.imagePaths) {
				return
			}
			path := p.imagePaths[i]
			imgCheck := o.(*ImageCheck)

			p.thumbMutex.Lock()
			if thumb, ok := p.thumbCache[path]; ok {
				imgCheck.Image = thumb
			} else {
				if t, found := p.db.GetThumbnail(path); found {
					imgCheck.Image = t
				}
			}
			p.thumbMutex.Unlock()

			imgCheck.Checked = slices.Contains(p.thumbnails.selectedIDs, i)
			imgCheck.OnChanged = func(checked bool) {
				if checked {
					if !slices.Contains(p.thumbnails.selectedIDs, i) {
						p.thumbnails.selectedIDs = append(p.thumbnails.selectedIDs, i)
					}
				} else {
					if idx := slices.Index(p.thumbnails.selectedIDs, i); idx != -1 {
						p.thumbnails.selectedIDs = slices.Delete(p.thumbnails.selectedIDs, idx, idx+1)
					}
				}
			}
			imgCheck.Refresh()
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
		fmt.Println(len(p.thumbnails.selectedIDs))
		fmt.Println(p.thumbnails.selectedIDs)

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

		if isExtendedSelection() {
			if p.thumbnails.selectionAnchor == -1 {
				p.thumbnails.selectionAnchor = id - 1
			}
			start, end := p.thumbnails.selectionAnchor, id
			if start > end {
				start, end = end, start
			}

			// Clear previous selection and select the new range
			p.thumbnails.selectedIDs = []widget.GridWrapItemID{}
			for i := start; i <= end; i++ {
				p.thumbnails.selectedIDs = append(p.thumbnails.selectedIDs, i)
			}
			p.thumbnails.Refresh()
		}
	}

	// p.thumbnails.OnHovered = func(id widget.GridWrapItemID) {
	// 	// fmt.Println("highlighted:", id)
	// 	path := p.imagePaths[id]
	// 	p.updatePreview(path)
	// 	if isExtendedSelection() {
	//
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
