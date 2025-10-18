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
	GoToTab(id int)
	GetWindow() fyne.Window
	UpdatePreview(i image.Image, path string)
	GetBinCount() int
}

type Controller struct {
	ui         CoreUI
	db         *database.DB
	imageCache map[string]database.CachedImage
	cacheMutex *sync.Mutex

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
		go c.cacheImages(total, &processedCount)
	}
	c.wg.Wait()

	if len(c.imageCache) > 0 {
		if err := c.db.SetThumbnailsBatch(c.imageCache); err != nil {
			log.Printf("Error during batch thumbnail write: %v", err)
		}
	}

	c.ui.ReloadAll()
	c.ui.HideProgressDialog()
	c.ui.GoToTab(0)
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

func (c *Controller) cacheImages(total float64, processedCount *int64) {
	defer c.wg.Done()
	for imgPath := range c.jobs {
		if img, found := c.getFromDBCache(imgPath); found {
			c.cacheMutex.Lock()
			c.imageCache[imgPath] = img
			c.cacheMutex.Unlock()
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
		_ = file.Close()
		if err != nil {
			log.Printf("could not decode image %s: %v", imgPath, err)
			continue
		}

		thumb := resize.Thumbnail(200, 200, img, resize.Lanczos3)
		preview := resize.Thumbnail(800, 600, img, resize.Lanczos3)
		c.cacheMutex.Lock()
		c.imageCache[imgPath] = database.CachedImage{
			Thumbnail: thumb,
			Preview:   preview,
		}
		c.cacheMutex.Unlock()

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

// getCachedImage gets the image from the database cache for in memory caching
func (c *Controller) getFromDBCache(path string) (database.CachedImage, bool) {
	thumbFound := false
	previewFound := false
	img := database.CachedImage{}
	if c.db != nil {
		if thumb, found := c.db.GetThumbnail(path); found {
			img.Thumbnail = thumb
			thumbFound = true
		}

		if preview, found := c.db.GetPreview(path); found {
			img.Preview = preview
			previewFound = true
		}
	}

	if thumbFound && previewFound {
		return img, true
	}

	return img, false
}

// Getthumbnail to be used by controller clients and retrieve the thumbnail from the in memory cache
func (c *Controller) GetThumbnail(path string) image.Image {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	if img, ok := c.imageCache[path]; ok {
		return img.Thumbnail
	}

	return nil
}

// Getthumbnail to be used by controller clients and retrieve the preview from the in memory cache
func (c *Controller) GetPreview(path string) image.Image {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	if img, ok := c.imageCache[path]; ok {
		return img.Preview
	}

	return nil
}

func (c *Controller) UpdatePreview(path string) {
	c.ui.UpdatePreview(c.GetPreview(path), path)
}

func (c *Controller) MoveImages(paths []string, sourceID, destID int) {
	if sourceID == destID || destID > c.ui.GetBinCount() {
		return
	}

	err := c.db.UpdateImages(paths, sourceID, destID)
	if err != nil {
		c.ui.ShowErrorDialog(err)
		return
	}
	c.ui.ReloadBin(sourceID)
	c.ui.ReloadBin(destID)
}

func New(ui CoreUI) *Controller {
	return &Controller{
		ui:         ui,
		imageCache: make(map[string]database.CachedImage),
		cacheMutex: &sync.Mutex{},
	}
}
