package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	defer recoverAndRestart() // If the program panics, this will attempt to restart it.

	log.SetOutput(os.Stdout)
	log.Println("========================================")
	log.Println("Initializing fermat...")

	log.Println("[Step 1] Downloading docker-compose file...")
	data, err := downloadFileFromGoogleBucket("swizzle_scripts", "docker-compose.yaml")
	if err != nil {
		log.Fatalf("[Error] Failed to download docker-compose file: %s", err)
	}

	log.Println("[Step 2] Saving docker-compose.yaml to disk...")
	err = saveBytesToFile("docker-compose.yaml", data)
	if err != nil {
		log.Fatalf("[Error] Failed to save docker-compose.yaml to disk: %s", err)
	}

	log.Println("[Step 3] Checking if repository exists...")
	repoExists, err := directoryExists("code")
	if !repoExists || err != nil {
		log.Println("[Step 4] Configuring Git...")
		gitSetup()

		log.Println("[Step 5] Cloning repository...")
		gitCloneRepo()
	}

	log.Println("[Step 6] Running docker-compose...")
	err = runDockerCompose()
	if err != nil {
		log.Fatalf("[Error] Failed to run docker-compose: %v", err)
	}

	log.Println("[Step 7] Setting up HTTP server...")
	setupHTTPServer()

	log.Println("========================================")
	log.Println("Fermat is now running!")
}

// recoverAndRestart will attempt to recover from a panic and restart the program.
func recoverAndRestart() {
	if r := recover(); r != nil {
		log.Println("[Fatal Error] Program crashed with error:", r)
		log.Println("Attempting to restart...")

		exec.Command(os.Args[0]).Run()
	}
}

// setupHTTPServer sets up the necessary HTTP routes and starts the server.
func setupHTTPServer() {
	http.Handle("/editor/", theiaProxy("8080"))
	http.Handle("/runner/", proxyPass("4411"))
	http.Handle("/database/", proxyPass("27017"))

	// handlers to show default code package.json
	http.HandleFunc("/code/package.json", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// CommitRequest represents the structure of the request body
	type CommitRequest struct {
		CommitMessage string `json:"commit_message"`
	}

	http.HandleFunc("/commit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.ContentLength > 4096 { // 4KB
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		commitMessage := fmt.Sprintf("swizzle commit: %s", time.Now().Format(time.RFC3339))
		if err == nil && len(body) > 0 {
			var req CommitRequest
			if json.Unmarshal(body, &req) == nil && req.CommitMessage != "" {
				commitMessage = req.CommitMessage
			}
		}

		cmd := exec.Command("git", "commit", "-m", commitMessage)
		cmd.Dir = "code"
		out, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Commit successful: %s", string(out))
	})

	http.HandleFunc("/push_to_production", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		commitMessage := fmt.Sprintf("swizzle commit production: %s", time.Now().Format(time.RFC3339))
		cmd := exec.Command("git", "commit", "-a", "-m", commitMessage) // '-a' adds all changes
		cmd.Dir = "code"
		out, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit: %s", err), http.StatusInternalServerError)
			return
		}

		cmd = exec.Command("git", "push", "origin", "production")
		cmd.Dir = "code"
		out, err = cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to push: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Push successful: %s", string(out))
	})

	http.ListenAndServe(":1234", nil)
}
