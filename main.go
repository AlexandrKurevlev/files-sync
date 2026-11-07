package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
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

type FileHash struct {
	Path string
	Hash string
}

const numWorkers = 10

func main() {
	filepaths, err := ListFiles("./")
	if err != nil {
		fmt.Println(err)
		return
	}

	tasks := make(chan string)
	results := make(chan FileHash)

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range tasks {
				hash, _ := HashFile(path)
				results <- FileHash{Path: path, Hash: hash}
			}
		}()
	}

	go func() {
		for _, path := range filepaths {
			tasks <- path
		}
		close(tasks)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for hash := range results {
		fmt.Println(hash.Path, hash.Hash)
	}
}
