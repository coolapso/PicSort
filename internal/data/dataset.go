package data

import (
	"image"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

type Dataset struct {
	Path   string
	Images []string
}

func NewDataset(path string) (*Dataset, error) {
	d := &Dataset{Path: path}
	err := filepath.WalkDir(path, func(s string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() {
			ext := strings.ToLower(filepath.Ext(s))
			if imageExtensions[ext] {
				d.Images = append(d.Images, s)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return d, nil
}

type ThumbnailCache struct {
	sync.RWMutex
	cache map[string]image.Image
}

func NewThumbnailCache() *ThumbnailCache {
	return &ThumbnailCache{
		cache: make(map[string]image.Image),
	}
}

func (c *ThumbnailCache) Get(path string) (image.Image, bool) {
	c.RLock()
	defer c.RUnlock()
	img, found := c.cache[path]
	return img, found
}

func (c *ThumbnailCache) Set(path string, img image.Image) {
	c.Lock()
	defer c.Unlock()
	c.cache[path] = img
}

func (c *ThumbnailCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.cache = make(map[string]image.Image)
}
