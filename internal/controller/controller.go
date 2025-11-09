package controller

import (
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/coolapso/picsort/internal/data"
	"github.com/coolapso/picsort/internal/database"
	"github.com/nfnt/resize"
)

var (
	errInvalidDestination = errors.New("cannot export dataset to the same location, please choose a different destination")
)

type CoreUI interface {
	ShowProgressDialog(msg string)
	SetProgress(progress float64, f string)
	ShowErrorDialog(err error)
	HideProgressDialog()
	GetBinCount() int
	LoadContent()
}

type Controller struct {
	ui          CoreUI
	db          *database.DB
	datasetRoot string
	newCached   bool
	mut         *sync.Mutex

	wg   *sync.WaitGroup
	jobs chan string
}

func (c *Controller) copyImages(imgPaths []string, datasetRoot string, binID int) error {
	total := float64(len(imgPaths))
	var copiedCount int64
	var failedCopy []string

	destinationDir := filepath.Join(datasetRoot, fmt.Sprint(binID))
	err := os.Mkdir(destinationDir, 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	for _, imgPath := range imgPaths {
		var img []byte
		fileName := filepath.Base(imgPath)
		img, err := os.ReadFile(imgPath)
		if err != nil {
			log.Println("Failed to read file:", err)
			failedCopy = append(failedCopy, fileName)
			continue
		}
		destinationPath := filepath.Join(destinationDir, fileName)
		err = os.WriteFile(destinationPath, img, 0644)
		if err != nil {
			log.Println("Failed to write file to destination:", err)
			failedCopy = append(failedCopy, fileName)
			continue
		}
		atomic.AddInt64(&copiedCount, 1)
		progress := float64(atomic.LoadInt64(&copiedCount)) / total
		c.ui.SetProgress(progress, fmt.Sprintf("bin %d/%s", binID, fileName))
	}

	if len(failedCopy) > 0 {
		return fmt.Errorf("failed to copy %d files, check logfile for more details", len(failedCopy))
	}

	return nil
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
		if _, found := c.getFromDBCache(imgPath); found {
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
		//nolint:errcheck
		c.db.SetImage(imgPath, database.CachedImage{
			Thumbnail: thumb,
			Preview:   preview,
		})

		atomic.AddInt64(processedCount, 1)
		progress := float64(atomic.LoadInt64(processedCount)) / total
		c.ui.SetProgress(progress, filepath.Base(imgPath))
		if !c.newCached {
			c.mut.Lock()
			c.newCached = true
			c.mut.Unlock()
		}
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

func (c *Controller) LoadDataset(path string) {
	c.ui.ShowProgressDialog("hang on, this may take a while...")
	c.newCached = false
	c.datasetRoot = path
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

	c.ui.LoadContent()
}

func (c *Controller) ExportDataset(dest string) {
	if dest == c.datasetRoot {
		c.ui.ShowErrorDialog(errInvalidDestination)
		return
	}
	c.ui.ShowProgressDialog("hang on, this may take a while...")
	binCount := c.ui.GetBinCount()
	datasetRoot := filepath.Join(dest, "dataset_export")
	if err := os.Mkdir(datasetRoot, 0755); err != nil {
		if !os.IsExist(err) {
			c.ui.ShowErrorDialog(err)
			return
		}
	}

	for i := range binCount {
		imgPaths, err := c.db.GetImagePaths(i)
		if err != nil {
			log.Println("error getting image paths:", err)
			c.ui.ShowErrorDialog(err)
			return
		}
		err = c.copyImages(imgPaths, datasetRoot, i)
		if err != nil {
			c.ui.HideProgressDialog()
			c.ui.ShowErrorDialog(err)
			return
		}
	}

	c.ui.HideProgressDialog()
}


// Getthumbnail to be used by controller clients and retrieve the thumbnail from the in memory cache
func (c *Controller) GetThumbnail(path string) image.Image {
	if img, ok := c.db.GetThumbnail(path); ok {
		return img
	}

	return nil
}

// Getthumbnail to be used by controller clients and retrieve the preview from the in memory cache
func (c *Controller) GetPreview(path string) image.Image {
	if img, ok := c.db.GetPreview(path); ok {
		return img
	}

	return nil
}

func (c *Controller) MoveImages(paths []string, sourceID, destID int) error {
	if sourceID == destID || destID > c.ui.GetBinCount() {
		return nil
	}

	return c.db.UpdateImages(paths, sourceID, destID)
}

func New(ui CoreUI) *Controller {
	return &Controller{
		ui:  ui,
		mut: &sync.Mutex{},
	}
}
