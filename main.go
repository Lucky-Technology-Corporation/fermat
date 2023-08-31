package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	defer recoverAndRestart() // If the program panics, this will attempt to restart it.

	log.SetOutput(os.Stdout)
	log.Println("========================================")
	log.Println("Initializing fermat...")

	log.Println("[Step 1] Downloading docker-compose file...")
	dockerData, err := downloadFileFromGoogleBucket("swizzle_scripts", "docker-compose.yaml")
	if err != nil {
		log.Fatalf("[Error] Failed to download docker-compose file: %s", err)
	}

	log.Println("[Step 2] Saving docker-compose.yaml to disk...")
	err = saveBytesToFile("docker-compose.yaml", dockerData)
	if err != nil {
		log.Fatalf("[Error] Failed to save docker-compose.yaml to disk: %s", err)
	}

	log.Println("[Step 3] Downloading pascal (theia) docker image")
	pascalData, err := downloadFileFromGoogleBucket("swizzle_scripts", "pascal.tar")
	if err != nil {
		log.Fatalf("[Error] Failed to download pascal tarball (theia): %s", err)
	}

	log.Println("[Step 4] Saving pascal to disk...")
	err = saveBytesToFile("pascal.tar", pascalData)
	if err != nil {
		log.Fatalf("[Error] Failed to save docker-compose.yaml to disk: %s", err)
	}

	log.Println("[Step 5] Load docker image from tarball")
	err = loadDockerImageFromTarball("pascal.tar")
	if err != nil {
		log.Fatalf("[Error] Failed to load docker tarball: %s", err)
	}

	log.Println("[Step 5] Running docker compose...")
	err = runDockerCompose()
	if err != nil {
		log.Fatalf("[Error] Failed to run docker-compose: %v", err)
	}

	log.Println("[Step 6] Setting up HTTP server...")

	done := make(chan bool, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := setupHTTPServer(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
		done <- true
	}()

	log.Println("========================================")
	log.Println("Fermat is now running!")

	select {
	case <-done:
		log.Println("Server finished.")
	case sig := <-signals:
		log.Printf("Received signal %s. Shutting down gracefully...\n", sig)
	}

	log.Println("Shutdown complete!")
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
func setupHTTPServer() error {
	http.Handle("/editor/", theiaProxy("3000"))
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

	http.HandleFunc("/commit", commitHandler)

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

	err := http.ListenAndServe(":1234", nil)
	if err != nil {
		return err
	}

	return nil
}
