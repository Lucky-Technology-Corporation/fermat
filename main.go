package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle the GET request
	query := r.URL.Query().Get("query")
	response := fmt.Sprintf("Received query: %s", query)
	w.Write([]byte(response))
}

func main() {
	downloadURL := "https://storage.googleapis.com/swizzle_scripts_pub/docker-compose.yaml"
	err := downloadFileFromURL(downloadURL, "docker-compose.yaml")
	if err != nil {
		log.Fatalf("Failed to download docker-compose file: %v", err)
	}

	repoExists, err := directoryExists("code")

	// configure git, clone repo only if needed
	if repoExists == false || err != nil {
		gitUsername := os.Getenv("GIT_USERNAME")
		if gitUsername == "" {
			gitUsername = "Swizzle User"
		}

		gitEmail := os.Getenv("GIT_EMAIL")
		if gitEmail == "" {
			gitEmail = "default@swizzle.co"
		}

		cmd := exec.Command("git config --global user.name \"%s\"", gitUsername)
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		cmd = exec.Command("git config --global user.email \"%s\"", gitEmail)
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		cmd = exec.Command("git clone ssh://sean@useblade.com@source.developers.google.com:2022/p/swizzle-prod/r/swizzle-webserver-template code")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}

	// docker compose down -> up (reset if already running, otherwise start)
	err = runDockerCompose()
	if err != nil {
		log.Fatalf("Failed to run docker-compose: %v", err)
	}

	// reverse proxy, rewriter
	http.Handle("/editor/", theiaProxy("8080"))
	http.Handle("/runner/", proxyPass("4411"))
	http.Handle("/database/", proxyPass("27017"))

	// handlers to show default code package.json
	http.HandleFunc("/code/package.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

		file, err := os.Open("code/package.json") // Replace with your filename
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Write the file contents to the response
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
		if r.Method != "POST" {
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
		cmd.Dir = "/code"
		out, err := cmd.CombinedOutput()

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Commit successful: %s", out)
	})

	http.HandleFunc("/push_to_production", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Commit changes on the current branch
		commitMessage := fmt.Sprintf("swizzle commit production: %s", time.Now().Format(time.RFC3339))
		cmd := exec.Command("git", "commit", "-a", "-m", commitMessage) // '-a' adds all changes
		cmd.Dir = "/code"
		out, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit: %s", err), http.StatusInternalServerError)
			return
		}

		// Push changes to the 'production' branch on the remote
		cmd = exec.Command("git", "push", "origin", "HEAD:production")
		cmd.Dir = "/code"
		out, err = cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to push to production: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Pushed to production successfully: %s", out)
	})

	http.HandleFunc("/code/table_of_contents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

		root := "./code" // start at the directory you want

		var files []string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() { // We only want to collect file names, not directory names
				relativePath := strings.TrimPrefix(path, root)
				files = append(files, relativePath)
			}
			return nil
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("Error walking the directory: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, file := range files {
			fmt.Fprintln(w, file)
		}
	})

	server := &http.Server{
		Addr:         ":1234",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on :1234")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("server failure: %s", err)
		return
	}
}
