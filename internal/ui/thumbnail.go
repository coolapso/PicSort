package ui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Thumbnail struct {
	widget.BaseWidget
	Image     image.Image
	Checked   bool
	OnChanged func(bool)
}

func (ic *Thumbnail) CreateRenderer() fyne.WidgetRenderer {
	r := &thumbnailRenderer{
		thumbnail: ic,
		thumb:     canvas.NewImageFromImage(ic.Image),
		checkIcon: canvas.NewImageFromResource(theme.CheckButtonIcon()),
	}
	r.thumb.FillMode = canvas.ImageFillContain
	r.checkIcon.Hide()
	return r
}

type thumbnailRenderer struct {
	thumbnail *Thumbnail
	thumb     *canvas.Image
	checkIcon *canvas.Image
}

func (r *thumbnailRenderer) Layout(size fyne.Size) {
	r.thumb.Resize(size)
	r.checkIcon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
	r.checkIcon.Move(fyne.NewPos(size.Width-theme.IconInlineSize()-theme.Padding(), theme.Padding()))
}

func (r *thumbnailRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 200)
}

func (r *thumbnailRenderer) Refresh() {
	r.thumb.Image = r.thumbnail.Image
	r.thumb.Refresh()
	if r.thumbnail.Checked {
		r.checkIcon.Resource = theme.CheckButtonCheckedIcon()
		r.checkIcon.Show()
	} else {
		r.checkIcon.Resource = theme.CheckButtonIcon()
		r.checkIcon.Hide()
	}
	r.checkIcon.Refresh()
	canvas.Refresh(r.thumbnail)
}

func (r *thumbnailRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.thumb, r.checkIcon}
}

func (r *thumbnailRenderer) Destroy() {}

func NewThumbnail(img image.Image, onChanged func(bool)) *Thumbnail {
	ic := &Thumbnail{
		Image:     img,
		OnChanged: onChanged,
	}

	ic.ExtendBaseWidget(ic)
	return ic
}
