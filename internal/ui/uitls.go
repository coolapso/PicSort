package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/coolapso/picsort/internal/data"
	"github.com/coolapso/picsort/internal/database"
	"github.com/nfnt/resize"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

func (p *PicsortUI) setProgress(progress float64, f string) {
	fyne.Do(func() {
		p.progressFile.SetText(f)
		p.progressValue.Set(progress)
	})
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
		p.thumbCache = make(map[string]image.Image)
	})
}

func (p *PicsortUI) cacheThumbnails(total float64, processedCount *int64) {
	defer p.wg.Done()
	for imgPath := range p.jobs {
		if thumb, found := p.db.GetThumbnail(imgPath); found {
			p.thumbMutex.Lock()
			p.thumbCache[imgPath] = thumb
			p.thumbMutex.Unlock()
			atomic.AddInt64(processedCount, 1)
			progress := float64(atomic.LoadInt64(processedCount)) / total
			p.setProgress(progress, filepath.Base(imgPath))
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

		atomic.AddInt64(processedCount, 1)
		progress := float64(atomic.LoadInt64(processedCount)) / total
		p.setProgress(progress, filepath.Base(imgPath))

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
	fyne.Do(func() { p.thumbnails.Refresh() })

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
		p.win.Canvas().Focus(p.thumbnails)
	})
}

func (p *PicsortUI) updatePreview(path string) {
	go func() {
		file, err := os.Open(path)
		if err != nil {
			log.Printf("could not open file for preview %s: %v", path, err)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			log.Printf("could not decode image for preview %s: %v", path, err)
			return
		}

		fyne.Do(func() {
			p.preview.Image = img
			p.preview.Refresh()
			p.previewCard.SetSubTitle(filepath.Base(path))
		})
	}()
}
