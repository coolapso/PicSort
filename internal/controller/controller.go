package controller

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"github.com/coolapso/picsort/internal/data"
	"github.com/coolapso/picsort/internal/database"
	"github.com/nfnt/resize"
)

type CoreUI interface {
	ShowProgressDialog(msg string)
	SetProgress(progress float64, f string)
	HideProgressDialog()
	ReloadAll()
	ReloadBin(id int)
	ShowErrorDialog(err error)
	FocusThumbnails(id int)
	GetWindow() fyne.Window
	UpdatePreview(i image.Image, path string)
}

type Controller struct {
	ui         CoreUI
	db         *database.DB
	thumbCache map[string]image.Image
	thumbMutex *sync.Mutex
	// imagePaths []string

	wg   *sync.WaitGroup
	jobs chan string
}

func (c *Controller) LoadDataset(path string) {
	c.ui.ShowProgressDialog("hang on, this may take a while...")
	if err := c.dbinit(path); err != nil {
		c.ui.ShowErrorDialog(err)
		return
	}

	d, err := data.NewDataset(path)
	if err != nil {
		c.ui.ShowErrorDialog(err)
		return
	}

	imagePaths := d.Images
	c.ui.ReloadAll()

	total := float64(len(imagePaths))
	var processedCount int64

	c.jobs = make(chan string, len(imagePaths))
	for _, p := range imagePaths {
		c.jobs <- p
	}
	close(c.jobs)

	c.wg = &sync.WaitGroup{}
	numWorkers := runtime.NumCPU()
	c.wg.Add(numWorkers)

	for range numWorkers {
		go c.cacheThumbnails(total, &processedCount)
	}
	c.wg.Wait()

	if len(c.thumbCache) > 0 {
		if err := c.db.SetThumbnailsBatch(c.thumbCache); err != nil {
			log.Printf("Error during batch thumbnail write: %v", err)
		}
	}

	c.ui.ReloadAll()
	c.ui.HideProgressDialog()
	c.ui.FocusThumbnails(0)
}

func (c *Controller) dbinit(path string) error {
	if c.db != nil {
		c.db.Close()
	}

	db, err := database.New(path)
	if err != nil {
		log.Println("error opening database:", err)
		return err
	}

	c.db = db
	return nil
}

func (c *Controller) cacheThumbnails(total float64, processedCount *int64) {
	defer c.wg.Done()
	for imgPath := range c.jobs {
		if _, found := c.GetThumbnail(imgPath); found {
			atomic.AddInt64(processedCount, 1)
			progress := float64(atomic.LoadInt64(processedCount)) / total
			c.ui.SetProgress(progress, filepath.Base(imgPath))
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
		c.thumbMutex.Lock()
		c.thumbCache[imgPath] = thumb
		c.thumbMutex.Unlock()

		atomic.AddInt64(processedCount, 1)
		progress := float64(atomic.LoadInt64(processedCount)) / total
		c.ui.SetProgress(progress, filepath.Base(imgPath))
	}
}

func (c *Controller) GetImagePaths(binID int) []string {
	if c.db == nil {
		return nil
	}
	paths, err := c.db.GetImagePaths(binID)
	if err != nil {
		message := fmt.Errorf("failed to get image paths: %v", err)
		log.Println(message)
		c.ui.ShowErrorDialog(message)
		return nil
	}
	return paths
}

func (c *Controller) GetThumbnail(path string) (image.Image, bool) {
	c.thumbMutex.Lock()
	defer c.thumbMutex.Unlock()
	if thumb, ok := c.thumbCache[path]; ok {
		return thumb, true
	}

	if c.db != nil {
		if thumb, found := c.db.GetThumbnail(path); found {
			c.thumbCache[path] = thumb
			return thumb, true
		}
	}

	return nil, false
}

func (c *Controller) UpdatePreview(path string) {
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

		c.ui.UpdatePreview(img, path)
	}()
}

func (c *Controller) ThumbMutex() *sync.Mutex {
	return c.thumbMutex
}

func New(ui CoreUI) *Controller {
	return &Controller{
		ui:         ui,
		thumbCache: make(map[string]image.Image),
		thumbMutex: &sync.Mutex{},
	}
}
