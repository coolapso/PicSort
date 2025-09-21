package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/coolapso/picsort/internal/ui"
)

func main() {
	a := app.NewWithID("picsort")
	w := a.NewWindow("PicSort")

	picsortUI := ui.New(a, w)
	picsortUI.Build()
	w.ShowAndRun()
}
