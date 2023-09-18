package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type UpdateGoogleCredentialsRequest struct {
	// Base64 encoded google credentials
	Credentials string `json:"credentials"`
}

func updateGoogleCredentials(w http.ResponseWriter, r *http.Request) {
	var req UpdateGoogleCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	credentials, err := base64.StdEncoding.DecodeString(req.Credentials)
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

	err = ioutil.WriteFile(filePath, credentials, 0644)
	if err != nil {
		http.Error(w, "Failed to write credentials file", http.StatusInternalServerError)
		return
	}

	// Re-auth with new credentials
	var cmdOut bytes.Buffer
	cmd := exec.Command("gcloud", "auth", "activate-service-account", "--key-file", filePath)
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdOut

	if err = cmd.Run(); err != nil {
		log.Println("Failed to run gcloud auth command")
		log.Println(cmdOut.String())

		http.Error(w, "Failed to re-authenticate using new credentials file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
