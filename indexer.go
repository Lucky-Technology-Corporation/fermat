package main

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path/filepath"
)

func indexFile(filePath string, index bleve.Index) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filePath, err)
		return
	}
	index.Index(filePath, string(content))
}

func recursiveIndex(dir string, index bleve.Index) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			indexFile(path, index)
		}
		return nil
	})
}

func fileIndexerWatcher() {
	// Create or open index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("fermat_swizzle.bleve", mapping)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer index.Close()

	// Initial indexing
	recursiveIndex(".", index)
	fmt.Println("Initial indexing complete.")

	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	// Recursively add directories to the watcher
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	// Process events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("Modified file:", event.Name)
				indexFile(event.Name, index)
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Println("Removed file:", event.Name)
				index.Delete(event.Name)
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				fmt.Println("Renamed file:", event.Name)
				index.Delete(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error:", err)
		}
	}
}
