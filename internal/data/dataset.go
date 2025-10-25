package data

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
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
