package ui

import (
	_ "embed"
	"fmt"
	"image"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/coolapso/picsort/internal/data"
	"github.com/coolapso/picsort/internal/database"
	"github.com/nfnt/resize"
)

func newURLToolbarAction(a fyne.App, icon fyne.Resource, urlStr string) widget.ToolbarItem {
	return widget.NewToolbarAction(icon, func() {
		u, _ := url.Parse(urlStr)
		_ = a.OpenURL(u)
	})
}

func (p *PicsortUI) topBar() *fyne.Container {
	openDataSetButton := widget.NewButton("Open dataset", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Println("Error opening folder dialog:", err)
				return
			}
			if uri == nil {
				return
			}
			go p.loadThumbnails(uri.Path())
		}, p.win)
		folderDialog.Resize(fyne.NewSize(800, 600)) // Set the size here
		folderDialog.Show()
	})

	exportButton := widget.NewButton("Export", func() {})

	helpButton := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {})

	return container.NewBorder(nil, nil, nil, helpButton,
		container.NewHBox(openDataSetButton, exportButton),
	)
}

func (p *PicsortUI) dbinit(path string) {
	if p.db != nil {
		p.db.Close()
	}

	db, err := database.New(path)
	if err != nil {
		log.Println("error opening database:", err)
		fyne.Do(func() {
			dialog.ShowError(err, p.win)
			p.progressDialog.Hide()
		})
		return
	}
	p.db = db
}

func (p *PicsortUI) showErrorDialog(err error) {
	log.Println("error loading dataset:", err)
	fyne.Do(func() {
		p.progressDialog.Hide()
		dialog.ShowError(err, p.win)
	})
}

func (p *PicsortUI) showProgressDialog(msg string) {
	fyne.Do(func() {
		p.progressTitle.SetText(msg)
		p.progressDialog.Show()
		p.progress.Show()
		p.progressValue.Set(0)
		p.imagePaths = []string{}
		p.thumbCache = make(map[string]image.Image)
	})
}

func (p *PicsortUI) setProgress(progress float64, f string) {
	fyne.Do(func() {
		p.progressFile.SetText(f)
		p.progressValue.Set(progress)
	})
}

func (p *PicsortUI) cacheThumbnails(total float64, processedCount *int64) {
	defer p.wg.Done()
	for imgPath := range p.jobs {
		atomic.AddInt64(processedCount, 1)
		progress := float64(atomic.LoadInt64(processedCount)) / total
		p.setProgress(progress, filepath.Base(imgPath))
		if _, found := p.db.GetThumbnail(imgPath); found {
			continue
		}

		file, err := os.Open(imgPath)
		if err != nil {
			log.Printf("could not open file %s: %v", imgPath, err)
			continue
		}

		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			log.Printf("could not decode image %s: %v", imgPath, err)
			continue
		}

		thumb := resize.Thumbnail(200, 200, img, resize.Lanczos3)
		p.thumbMutex.Lock()
		p.thumbCache[imgPath] = thumb
		p.thumbMutex.Unlock()
	}
}

func (p *PicsortUI) loadThumbnails(path string) {
	p.showProgressDialog("Hang on, this may take a while...")
	p.dbinit(path)

	d, err := data.NewDataset(path)
	if err != nil {
		p.showErrorDialog(err)
		return
	}

	p.imagePaths = d.Images
	total := float64(len(p.imagePaths))
	var processedCount int64

	p.jobs = make(chan string, len(p.imagePaths))
	for _, path := range p.imagePaths {
		p.jobs <- path
	}
	close(p.jobs)

	p.thumbMutex = &sync.Mutex{}
	p.wg = &sync.WaitGroup{}
	numWorkers := runtime.NumCPU()
	p.wg.Add(numWorkers)

	for range numWorkers {
		go p.cacheThumbnails(total, &processedCount)
	}

	p.wg.Wait()

	if len(p.thumbCache) > 0 {
		if err := p.db.SetThumbnailsBatch(p.thumbCache); err != nil {
			log.Printf("Error during batch thumbnail write: %v", err)
		}
	}

	fyne.Do(func() {
		p.thumbnails.Refresh()
		p.progressDialog.Hide()
	})
}

func (p *PicsortUI) bottomBar() fyne.Widget {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			if len(p.bins.Objects) <= 9 {
				binCount := len(p.bins.Objects) + 1
				p.bins.Add(widget.NewCard(fmt.Sprintf("Bin %d", binCount), "", nil))
				p.bins.Layout = layout.NewGridLayout(binCount)
				p.bins.Refresh()
			}
		}),

		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			if len(p.bins.Objects) > 1 {
				binCount := len(p.bins.Objects) - 1
				p.bins.Remove(p.bins.Objects[binCount])
				p.bins.Layout = layout.NewGridLayout(binCount)
				p.bins.Refresh()
			}
		}),

		widget.NewToolbarSpacer(),
		newURLToolbarAction(p.app, Icons["sponsor"], "https://github.com/sponsors/coolapso"),
		newURLToolbarAction(p.app, Icons["bmc"], "https://buymeacoffee.com"),
		newURLToolbarAction(p.app, Icons["github"], "https://github.com/coolapso/picsort"),
		newURLToolbarAction(p.app, Icons["discord"], "https://discord.com"),
		newURLToolbarAction(p.app, Icons["mastodon"], "https://mastodon.social/@coolapso"),
		newURLToolbarAction(p.app, Icons["x"], "https://x.com"),
	)
}
