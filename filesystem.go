package main

import (
	"os"
	"path/filepath"
	"strings"
)

type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	IsScript bool
}

func isShellScript(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".sh" || ext == ".bash" || ext == ".zsh"
}

func readDirectory(dirPath string) ([]FileItem, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var items []FileItem
	
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		
		fullPath := filepath.Join(dirPath, entry.Name())
		item := FileItem{
			Name:     entry.Name(),
			Path:     fullPath,
			IsDir:    entry.IsDir(),
			IsScript: !entry.IsDir() && isShellScript(entry.Name()),
		}
		
		if item.IsDir || item.IsScript {
			items = append(items, item)
		}
	}
	
	return items, nil
}