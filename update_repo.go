package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type UpdateRepoRequest struct {
	// Base64 encoded service account credentials
	GoogleSACredentials string `json:"google_sa_credentials"`
	GoogleSourceRepo    string `json:"google_source_repo"`
}

func updateRepo(w http.ResponseWriter, r *http.Request) {
	var req UpdateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	credentials, err := base64.StdEncoding.DecodeString(req.GoogleSACredentials)
	if err != nil {
		http.Error(w, "Credentials must be base64 encoded", http.StatusBadRequest)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Can't get user's home directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")

	err = os.WriteFile(filePath, credentials, 0644)
	if err != nil {
		http.Error(w, "Failed to write credentials file", http.StatusInternalServerError)
		return
	}

	runner := &CommandRunner{dir: "code"}

	runner.Run("gcloud", "auth", "activate-service-account", "--key-file", filePath)
	runner.Run("git", "remote", "set-url", "origin", req.GoogleSourceRepo)
	runner.Run("git", "config", "--replace-all", "credential.'https://source.developers.google.com/'.helper", "'!gcloud auth git-helper --ignore-unknown $@'")

	if runner.err != nil {
		http.Error(w, runner.err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
