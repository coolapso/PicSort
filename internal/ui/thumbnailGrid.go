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
	GetThumbnail(path string) (image.Image, bool)
	UpdatePreview(path string)
	GetImagePaths(bindID int) []string
}

type ThumbnailGridWrap struct {
	widget.GridWrap
	id              int
	selectionAnchor widget.GridWrapItemID
	selectedIDs     []widget.GridWrapItemID
	previousKey     fyne.KeyName
	previousKeyAt   time.Time

	dataProvider ThumbnailProvider
	imagePaths   []string
}

func (g *ThumbnailGridWrap) TypedKey(key *fyne.KeyEvent) {
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
	case fyne.KeyG:
		if g.isDoublePress(key) {
			if shiftPressed() {
				g.ScrollToBottom()
			} else {
				g.ScrollToTop()
				g.Refresh()
			}
			return
		}
		return
	case fyne.KeyHome:
		g.ScrollToTop()
		return
	case fyne.KeyEnd:
		g.ScrollToBottom()
		return
	case fyne.KeyEscape:
		g.unselectAll()
		return
	}

	g.GridWrap.TypedKey(&translatedKey)
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
	if id >= len(g.imagePaths) {
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
	if id >= len(g.imagePaths) {
		return
	}
	path := g.imagePaths[id]
	g.dataProvider.UpdatePreview(path)

	if !shiftPressed() {
		g.selectionAnchor = -1
	}

	if shiftPressed() {
		if g.selectionAnchor == -1 {
			g.selectionAnchor = id - 1
		}
		start, end := g.selectionAnchor, id
		if start > end {
			start, end = end, start
		}

		for i := start; i <= end; i++ {
			g.selectedIDs = append(g.selectedIDs, i)
		}
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
	if i >= len(g.imagePaths) {
		return
	}
	path := g.imagePaths[i]
	imgCheck := o.(*Thumbnail)

	if thumb, ok := g.dataProvider.GetThumbnail(path); ok {
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

func NewThumbnailGrid(id int, d ThumbnailProvider) *ThumbnailGridWrap {
	grid := &ThumbnailGridWrap{
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
