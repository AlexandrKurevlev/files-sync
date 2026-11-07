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

func GetHashes(pathToProcess string) (map[string]string, error) {
	filepaths, err := ListFiles(pathToProcess)
	if err != nil {
		return nil, err
	}

	tasks := make(chan string)
	results := make(chan FileHash)

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range tasks {
				hash, _ := HashFile(filepath.Join(pathToProcess, path))
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

	res := make(map[string]string)
	for hash := range results {
		res[hash.Path] = hash.Hash
	}
	return res, nil
}

func CompareScans(source, dest map[string]string) (toCopy, toUpdate, toDelete []string) {
	for srcKey, srcHash := range source {
		if destHash, exists := dest[srcKey]; !exists {
			toCopy = append(toCopy, srcKey)
		} else if srcHash != destHash {
			toUpdate = append(toUpdate, srcKey)
		}
	}

	for destKey, _ := range dest {
		if _, exists := source[destKey]; !exists {
			toDelete = append(toDelete, destKey)
		}
	}

	return
}

func main() {
	sourcePath := os.Args[1]
	destPath := os.Args[2]

	source, err := GetHashes(sourcePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	dest, err := GetHashes(destPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	toCopy, toUpdate, toDelete := CompareScans(source, dest)
	fmt.Println("Файлы для КОПИРОВАНИЯ:")
	for _, file := range toCopy {
		fmt.Println("-", file)
	}
	fmt.Println("Файлы для ОБНОВЛЕНИЯ:")
	for _, file := range toUpdate {
		fmt.Println("-", file)
	}
	fmt.Println("Файлы для УДАЛЕНИЯ:")
	for _, file := range toDelete {
		fmt.Println("-", file)
	}
}
