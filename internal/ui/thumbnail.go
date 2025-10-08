package ui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
