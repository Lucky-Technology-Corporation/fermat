package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	IsDir    bool    `json:"isDir"`
	Children []*File `json:"children,omitempty"`
}

func listDir(root string) (*File, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	file := &File{
		Name:  info.Name(),
		Path:  root,
		IsDir: info.IsDir(),
	}

	if info.IsDir() {
		files, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}

		children := make([]*File, 0)
		for _, f := range files {
			fullPath := filepath.Join(root, f.Name())
			if strings.Contains(fullPath, "node_modules") {
				continue
			}

			child, err := listDir(filepath.Join(root, f.Name()))
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
		file.Children = children
	}

	return file, nil
}

func tableOfContents(w http.ResponseWriter, r *http.Request) {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		http.Error(w, "HOME environment variable not found", http.StatusInternalServerError)
		return
	}
	root := filepath.Join(home, "code")
	result, err := listDir(root)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list directory: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func packageJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	file, err := os.Open("code/package.json")
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusNotFound)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close file: %v", cerr)
		}
	}()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to write file content", http.StatusInternalServerError)
	}
}
