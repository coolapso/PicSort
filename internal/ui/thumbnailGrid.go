package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

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

func (g *ThumbnailGridWrap) unselectAll() {
	g.selectedIDs = []widget.GridWrapItemID{}
	g.selectionAnchor = -1
	g.Refresh()
}

func (p *PicsortUI) NewThumbnailGrid() *ThumbnailGridWrap {
	return NewThumbnailGridWrap(
		func() int {
			return len(p.imagePaths)
		},
		func() fyne.CanvasObject {
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
