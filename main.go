package main

import (
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
