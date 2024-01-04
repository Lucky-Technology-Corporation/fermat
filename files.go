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
		//directory may not exist, return empty directory instead of an error
		fmt.Printf("Failed to list directory: %s", err)

		result = &File{
			Name:     "helpers",
			Path:     root,
			IsDir:    true,
			Children: []*File{},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	path := r.URL.Query().Get("path")

	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf(err.Error())
		http.Error(w, "Failed to delete file "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File deleted successfully!"))
	return
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

type WriteCodeFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func writeCodeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var req WriteCodeFileRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Can't get user's home directory", http.StatusInternalServerError)
		return
	}

	path := filepath.Join(home, "code", req.Path)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		http.Error(w, "Failed to create directories", http.StatusInternalServerError)
		return
	}

	file, err := os.Create(path) // Create the file (or overwrite if it exists)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close file: %v", cerr)
		}
	}()

	_, err = file.WriteString(req.Content)
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
