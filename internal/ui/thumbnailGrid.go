package ui

import (
	"image"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Core UI needs to provide way to get thumbnails and its paths, as well as a way to update the preview.
type coreUI interface {
	GetThumbnail(path string) (image.Image, bool)
	UpdatePreview(path string)
	GetImagePaths() []string
}

// ThumbnailGridWrap is essentially an extension of the standard fyne.Gridwarp, 
// And is a self contained grid responsible for presenting the thumbnails and navigation within the grid
type ThumbnailGridWrap struct {
	widget.GridWrap
	selectionAnchor widget.GridWrapItemID
	selectedIDs     []widget.GridWrapItemID

	coreUI     coreUI
	imagePaths []string
}

func NewThumbnailGridWrap(coreUI coreUI) *ThumbnailGridWrap {
	grid := &ThumbnailGridWrap{
		coreUI:          coreUI,
		selectionAnchor: -1,
		selectedIDs:     []widget.GridWrapItemID{},
	}

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
	case fyne.KeyEscape:
		g.unselectAll()
	}

	g.GridWrap.TypedKey(&translatedKey)
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
	g.coreUI.UpdatePreview(path)

	if !isExtendedSelection() {
		g.selectionAnchor = -1
	}

	if isExtendedSelection() {
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
	g.imagePaths = g.coreUI.GetImagePaths()
	g.Refresh()
}

func (g *ThumbnailGridWrap) itemCount() int {
	return len(g.imagePaths)
}

func (g *ThumbnailGridWrap) createItem() fyne.CanvasObject {
	return NewImageCheck(nil, nil)
}

func (g *ThumbnailGridWrap) updateItem(i widget.GridWrapItemID, o fyne.CanvasObject) {
	if i >= len(g.imagePaths) {
		return
	}
	path := g.imagePaths[i]
	imgCheck := o.(*ImageCheck)

	if thumb, ok := g.coreUI.GetThumbnail(path); ok {
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

func NewThumbnailGrid(t coreUI) *ThumbnailGridWrap {
	return NewThumbnailGridWrap(t)
}
