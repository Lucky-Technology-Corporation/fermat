package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type NPMInstallRequest struct {
	Packages []string `json:"packages"`
	Save     bool     `json:"save"`
}

func npmInstallHandler(w http.ResponseWriter, r *http.Request) {
	var req NPMInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(req.Packages) == 0 {
		http.Error(w, "Must specify at least 1 package", http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()
	path := queryParams.Get("path")

	if path == "" {
		http.Error(w, "Path can't be empty", http.StatusBadRequest)
		return
	}

	dir, err := filepath.Abs(filepath.Join("code", path))
	if err != nil {
		log.Println("Couldn't convert path to absolute:", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	packageJSON := filepath.Join(dir, "package.json")
	if !fileExists(packageJSON) {
		http.Error(w, "The directory: "+dir+" doesn't contain a package.json", http.StatusBadRequest)
		return
	}

	args := []string{"install"}
	args = append(args, req.Packages...)
	if req.Save {
		args = append(args, "--save")
	}

	runner := &CommandRunner{dir: dir}
	runner.RunDockerNpmCommand(args...)
	if runner.err != nil {
		http.Error(w, "Error running command: "+runner.err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Package installed successfully!"))
}

type NPMRemoveRequest struct {
	Packages []string `json:"packages"`
}

func npmRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var req NPMRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(req.Packages) == 0 {
		http.Error(w, "Must specify at least 1 package", http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()
	path := queryParams.Get("path")

	if path == "" {
		http.Error(w, "Path can't be empty", http.StatusBadRequest)
		return
	}

	dir, err := filepath.Abs(filepath.Join("code", path))
	if err != nil {
		log.Println("Couldn't convert path to absolute:", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	packageJSON := filepath.Join(dir, "package.json")
	if !fileExists(packageJSON) {
		http.Error(w, "The directory: "+dir+" doesn't contain a package.json", http.StatusBadRequest)
		return
	}

	args := append([]string{"remove"}, req.Packages...)
	runner := &CommandRunner{dir: dir}
	runner.RunDockerNpmCommand(args...)
	if runner.err != nil {
		http.Error(w, "Error running command: "+runner.err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Package removed successfully!"))
}

func (runner *CommandRunner) RunDockerNpmCommand(args ...string) {
	dockerArgs := []string{
		"run", "--rm",
		"-v", runner.dir + ":/app",
		"-w", "/app",
		"node:18-alpine", "npm",
	}

	dockerArgs = append(dockerArgs, args...)
	runner.Run("docker", dockerArgs...)
}

func fileExists(filename string) bool {
	// Dubious implementation... Should be good enough for our purposes to check
	// if package.json exists however.
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !os.IsNotExist(err)
}
