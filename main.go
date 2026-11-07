package main

import (
	"io/fs"
	"path/filepath"
)

func ListFiles(rootPath string) ([]string, error) {
	res := make([]string, 0)
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			relPath, _ := filepath.Rel(rootPath, path)
			res = append(res, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func main() {
}
