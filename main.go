package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
)

const VERSION = "0.0.3"
const SECRETS_FILE_PATH = "code/backend/secrets.json"
const WEBSERVER_KEYS_FILE = "webserver-keys.json"
const FEMRAT_KEYS_FILE = "fermat-keys.json"

func main() {
	defer recoverAndRestart() // If the program panics, this will attempt to restart it.

	log.SetOutput(os.Stdout)
	log.Println("========================================")
	log.Println("Initializing fermat...")
	log.Println("Running version: " + VERSION)

	firstTime := true
	firstTimeEnv := os.Getenv("FIRST_TIME")

	if firstTimeEnv == "false" {
		firstTime = false
		log.Println("[Info] Not first time. Skipping initialization steps.")
	} else {
		log.Println("[Info] First time running. Setting everything up.")
	}

	log.Println("[Info] Downloading docker-compose file...")
	err := downloadFileFromGoogleBucket(os.Getenv("BUCKET_NAME"), "docker-compose.yaml", "docker-compose.yaml")
	if err != nil {
		log.Fatalf("[Error] Failed to download and save docker-compose file: %s", err)
	}

	if firstTime {
		log.Println("[Info] Authenticating artifact registry...")
		err = setupArtifactRegistryAuth()
		if err != nil {
			log.Fatalf("[Error] Failed to authenticate with the artifact registry: %v", err)
		}

		log.Println("[Info] Setup Mongo Certificate for TLS support...")
		err = CreateAndSaveMongoCert()
		if err != nil {
			log.Fatalf("[Error] Failed to create and save certificate: %v", err)
		}

		log.Println("[Info] Writing secrets to file...")
		err = saveInitialSecrets()
		if err != nil {
			log.Fatalf("[Error] Failed to save initial secrets: %v", err)
		}
	}

	log.Println("[Info] Running docker compose...")
	err = runDockerCompose()
	if err != nil {
		log.Fatalf("[Error] Failed to run docker-compose: %v", err)
	}

	// Try switching back to webserver account
	if !firstTime {
		err := switchApplicationDefaultCredentialsToWebserver()
		if err != nil {
			log.Printf("[Error] Couldn't switch back to webserver service acccount: %v\n", err)
		}
	}

	log.Println("[Info] Setting up HTTP server...")

	done := make(chan bool, 1)
	shutdownChan := make(chan bool, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := setupHTTPServer(shutdownChan); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
		done <- true
	}()

	// Start Health Service Runner
	go HealthStatusServiceRunner()

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
func setupHTTPServer(shutdownChan chan bool) error {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

	// handlers to show default code package.json
	r.HandleFunc("/code/backend/package.json", packageJSON)
	r.HandleFunc("/code/frontend/package.json", packageJSONReact)
	r.HandleFunc("/code", getFileList)
	r.HandleFunc("/code/delete", deleteFile)

	//Get file contents
	r.HandleFunc("/code/file_contents", fileContents)
	//Write file contents
	r.HandleFunc("/code/write_file", writeCodeFile)

	r.Get("/spoof_jwt", spoofJwt)
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(VERSION))
	})

	r.HandleFunc("/commit", commitHandler)
	r.Post("/push_to_production", pushProduction)

	r.Get("/secrets", GetSecrets)
	r.Patch("/secrets", UpdateSecrets)

	r.Get("/services/health", HealthServiceHandler)

	r.Post("/update_repo", updateRepo)
	r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		err := activateServiceAccountKeys(FEMRAT_KEYS_FILE)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Failed to activate fermat service account", http.StatusInternalServerError)
			return
		}

		err = runDockerCompose()
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Failed to run docker compose", http.StatusInternalServerError)
			return
		}

		// In case we're refreshing a project without a production deployment we don't want the refresh
		// to report a failure since we'd expect this to fail with no webserver keys available.
		_ = activateServiceAccountKeys(WEBSERVER_KEYS_FILE)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Refresh success!"))
	})

	r.Post("/shutdown", ShutdownHandler(shutdownChan))
	r.Post("/restart_frontend", restartDockerContainerHandler("frontend"))
	r.Post("/restart_backend", restartDockerContainerHandler("backend"))

	// NPM commands
	r.Post("/npm/install", npmInstallHandler)
	r.Post("/npm/remove", npmRemoveHandler)

	r.Get("/tail_logs", tailLogsHandler)

	server := &http.Server{Addr: ":1234", Handler: r}

	go func() {
		<-shutdownChan
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not gracefully shutdown server: %v\n", err)
		}
	}()

	return server.ListenAndServe()
}
