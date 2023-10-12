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

func getFileList(w http.ResponseWriter, r *http.Request) {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		http.Error(w, "HOME environment variable not found", http.StatusInternalServerError)
		return
	}

	filePathFromQueryString := r.URL.Query().Get("path")

	root := filepath.Join(home, "code"+filePathFromQueryString)
	result, err := listDir(root)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list directory: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func fileContents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	path := r.URL.Query().Get("path")

	file, err := os.Open(path)
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

type WriteFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func writeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var fileRequest WriteFileRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&fileRequest)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	file, err := os.Create(fileRequest.Path) // Create the file (or overwrite if it exists)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close file: %v", cerr)
		}
	}()

	_, err = file.WriteString(fileRequest.Content)
	if err != nil {
		http.Error(w, "Failed to write to file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File written successfully!"))
	return
}

func packageJSONReact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	file, err := os.Open("code/frontend/package.json")
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

func packageJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	file, err := os.Open("code/backend/package.json")
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
