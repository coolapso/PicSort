package ui

import (
	"image"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// The ThumbnailProvider needs to provide way to get thumbnails and its paths, as well as a way to update the preview.
type ThumbnailProvider interface {
	GetThumbnail(path string) image.Image
	UpdatePreview(path string)
	GetImagePaths(bindID int) []string
	MoveImages(paths []string, sourceID, destID int)
}

type ThumbnailGridWrap struct {
	widget.GridWrap
	w               fyne.Window
	id              int
	selectionAnchor widget.GridWrapItemID
	selectedIDs     []widget.GridWrapItemID
	previousKey     fyne.KeyName
	previousKeyAt   time.Time
	currentID       widget.GridWrapItemID

	dataProvider ThumbnailProvider
	imagePaths   []string
}

func (g *ThumbnailGridWrap) TypedKey(key *fyne.KeyEvent) {
	key = translateKey(key)
	switch key.Name {
	case fyne.KeySpace:
		g.onSelected(g.currentID)
	case fyne.KeyLeft:
		g.GridWrap.TypedKey(key)
	case fyne.KeyRight:
		g.GridWrap.TypedKey(key)
	case fyne.KeyDown:
		g.GridWrap.TypedKey(key)
	case fyne.KeyUp:
		g.GridWrap.TypedKey(key)
	case fyne.KeyG:
		if g.isDoublePress(key) {
			if shiftPressed() {
				g.ScrollToItem(g.Length() - 1)
			} else {
				g.ScrollToItem(0)
			}
			return
		}
	case fyne.KeyHome:
		g.ScrollToItem(0)
	case fyne.KeyEnd:
		g.ScrollToItem(g.Length() - 1)
	case fyne.KeyEscape:
		g.unselectAll()
	case fyne.Key1:
		g.MoveImages(1)
	case fyne.Key2:
		g.MoveImages(2)
	case fyne.Key3:
		g.MoveImages(3)
	case fyne.Key4:
		g.MoveImages(4)
	case fyne.Key5:
		g.MoveImages(5)
	case fyne.Key6:
		g.MoveImages(6)
	case fyne.Key7:
		g.MoveImages(7)
	case fyne.Key8:
		g.MoveImages(8)
	case fyne.Key9:
		g.MoveImages(9)
	case fyne.Key0:
		g.MoveImages(0)
	default:
		if g.w.Canvas().OnTypedKey() != nil {
			g.w.Canvas().OnTypedKey()(key)
		}
	}
}

func (g *ThumbnailGridWrap) TypedRune(r rune) {
	if g.w.Canvas().OnTypedRune() != nil {
		g.w.Canvas().OnTypedRune()(r)
	}
}

func (g *ThumbnailGridWrap) isDoublePress(key *fyne.KeyEvent) bool {
	if time.Since(g.previousKeyAt) < 500*time.Millisecond {
		g.previousKey = ""
		g.previousKeyAt = time.Time{}
		return true
	}

	g.previousKey = key.Name
	g.previousKeyAt = time.Now()
	return false
}

func (g *ThumbnailGridWrap) onSelected(id widget.GridWrapItemID) {
	if id < 0 || id >= len(g.imagePaths) {
		return
	}

	if idx := slices.Index(g.selectedIDs, id); idx != -1 {
		g.selectedIDs = slices.Delete(g.selectedIDs, idx, idx+1)
	} else {
		g.selectedIDs = append(g.selectedIDs, id)
	}

	g.Refresh()
}

func (g *ThumbnailGridWrap) onHighlighted(id widget.GridWrapItemID) {
	if id < 0 || id >= len(g.imagePaths) {
		return
	}
	path := g.imagePaths[id]
	g.dataProvider.UpdatePreview(path)
	g.currentID = id

	if !shiftPressed() {
		g.selectionAnchor = -1
		return
	}

	if shiftPressed() {
		if g.selectionAnchor == -1 {
			g.selectionAnchor = id - 1
		}
		start, end := g.selectionAnchor, id
		if start > end {
			start, end = end, start
		}

		var newSelection []widget.GridWrapItemID
		for i := start; i <= end; i++ {
			newSelection = append(newSelection, i)
		}
		g.selectedIDs = newSelection
		g.Refresh()
	}
}

func (g *ThumbnailGridWrap) unselectAll() {
	g.selectedIDs = []widget.GridWrapItemID{}
	g.selectionAnchor = -1
	g.Refresh()
}

func (g *ThumbnailGridWrap) Reload() {
	g.imagePaths = g.dataProvider.GetImagePaths(g.id)
	g.Refresh()
}

func (g *ThumbnailGridWrap) itemCount() int {
	return len(g.imagePaths)
}

func (g *ThumbnailGridWrap) createItem() fyne.CanvasObject {
	return NewThumbnail(nil, nil)
}

func (g *ThumbnailGridWrap) updateItem(i widget.GridWrapItemID, o fyne.CanvasObject) {
	if i < 0 || i >= len(g.imagePaths) {
		return
	}
	path := g.imagePaths[i]
	imgCheck := o.(*Thumbnail)

	if thumb := g.dataProvider.GetThumbnail(path); thumb != nil {
		imgCheck.Image = thumb
	}

	imgCheck.Checked = slices.Contains(g.selectedIDs, i)
	imgCheck.OnChanged = func(checked bool) {
		if checked {
			if !slices.Contains(g.selectedIDs, i) {
				g.selectedIDs = append(g.selectedIDs, i)
			}
		} else {
			if idx := slices.Index(g.selectedIDs, i); idx != -1 {
				g.selectedIDs = slices.Delete(g.selectedIDs, idx, idx+1)
			}
		}
	}
	imgCheck.Refresh()
}

func (g *ThumbnailGridWrap) MoveImages(destID int) {
	if len(g.imagePaths) == 0 {
		return
	}

	var toMove []string
	if len(g.selectedIDs) == 0 {
		imagePath := g.imagePaths[g.currentID]
		if imagePath == "" {
			return
		}
		toMove = []string{imagePath}
	}

	if len(g.selectedIDs) > 0 {
		for _, id := range g.selectedIDs {
			toMove = append(toMove, g.imagePaths[id])
		}
	}

	g.dataProvider.MoveImages(toMove, g.id, destID)

	// need to force the preview update from here
	// Doing this because on highlighted is not triggered when refreshing,
	// and controller has no clue of what images each grid contains
	// therefore have to force the preview to update to the next item,
	// but got to make sure that is now out of bounds and
	// will have at least one to show after the update
	if g.currentID+1 > len(g.imagePaths)-1 {
		return
	}
	g.dataProvider.UpdatePreview(g.imagePaths[g.currentID+1])
	g.ScrollToItem(g.selectionAnchor)
}

func NewThumbnailGrid(id int, w fyne.Window, d ThumbnailProvider) *ThumbnailGridWrap {
	grid := &ThumbnailGridWrap{
		w:               w,
		dataProvider:    d,
		selectionAnchor: -1,
		selectedIDs:     []widget.GridWrapItemID{},
	}

	grid.id = id
	grid.Length = grid.itemCount
	grid.CreateItem = grid.createItem
	grid.UpdateItem = grid.updateItem
	grid.OnSelected = grid.onSelected
	grid.OnHighlighted = grid.onHighlighted
	grid.OnUnselected = nil

	grid.ExtendBaseWidget(grid)

	grid.Reload()
	return grid
}
