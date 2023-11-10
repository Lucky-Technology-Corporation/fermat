package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type NPMInstallRequest struct {
	Package string `json:"package"`
	Save    bool   `json:"save"`
}

func npmInstallHandler(w http.ResponseWriter, r *http.Request) {
	var req NPMInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if req.Package == "" {
		http.Error(w, "Package can't be empty", http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()
	path := queryParams.Get("path")

	if path == "" {
		http.Error(w, "Path can't be empty", http.StatusBadRequest)
		return
	}

	dir := filepath.Join("code", path)
	packageJSON := filepath.Join(dir, "package.json")
	if !fileExists(packageJSON) {
		http.Error(w, "The directory: "+dir+" doesn't contain a package.json", http.StatusBadRequest)
		return
	}

	args := []string{"install", req.Package}
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
	Package string `json:"package"`
}

func npmRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var req NPMRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if req.Package == "" {
		http.Error(w, "Package can't be empty", http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()
	path := queryParams.Get("path")

	if path == "" {
		http.Error(w, "Path can't be empty", http.StatusBadRequest)
		return
	}

	dir := filepath.Join("code", path)
	packageJSON := filepath.Join(dir, "package.json")
	if !fileExists(packageJSON) {
		http.Error(w, "The directory: "+dir+" doesn't contain a package.json", http.StatusBadRequest)
		return
	}

	runner := &CommandRunner{dir: dir}
	runner.RunDockerNpmCommand("remove", req.Package)
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
		"node:18", "npm",
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
args...}
