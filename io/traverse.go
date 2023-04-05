package io

import (
	"log"
	"os"
	"path/filepath"
	"sqlow/helpers"
)

func WalkDir(dirPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func GetAllFiles(path string, isRecursive bool) []string {
	if isRecursive {
		log.Printf("Getting all files in %s...\n", path)
		files, err := WalkDir(path)
		helpers.CheckError(err)
		return files
	}
	log.Printf("Getting %s...\n", path)
	return []string{path}
}
