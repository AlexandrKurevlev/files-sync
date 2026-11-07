package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
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

func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	result := hasher.Sum(nil)
	return hex.EncodeToString(result), nil
}

func main() {
	filepaths, err := ListFiles("./")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, filepath := range filepaths {
		hash, err := HashFile(filepath)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(filepath, hash)
	}
}
