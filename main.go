package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	defer recoverAndRestart() // If the program panics, this will attempt to restart it.

	log.SetOutput(os.Stdout)
	log.Println("========================================")
	log.Println("Initializing fermat...")

	log.Println("[Step 1] Downloading docker-compose file...")
	err := downloadFileFromGoogleBucket("swizzle_scripts", "docker-compose.yaml", "docker-compose.yaml")
	if err != nil {
		log.Fatalf("[Error] Failed to download and save docker-compose file: %s", err)
	}

	log.Println("[Step 2] Authenticating artifact registry...")
	err = setupArtifactRegistryAuth()
	if err != nil {
		log.Fatalf("[Error] Failed to authenticate with the artifact registry: %v", err)
	}

	log.Println("[Step 3] Running docker compose...")
	err = runDockerCompose()
	if err != nil {
		log.Fatalf("[Error] Failed to run docker-compose: %v", err)
	}

	log.Println("[Step 4] Setting up HTTP server...")

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
	mux := http.NewServeMux()

	mux.Handle("/editor/", theiaProxy("3000"))
	mux.Handle("/runner/", proxyPass("4411"))
	mux.Handle("/database/", proxyPass("27017"))

	// handlers to show default code package.json
	mux.HandleFunc("/code/package.json", packageJSON)
	mux.HandleFunc("/table_of_contents", tableOfContents)

	mux.HandleFunc("/spoof_jwt", spoofJwt)

	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/push_to_production", pushProduction)

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		err := runDockerCompose()
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Failed to run docker compose", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Refresh success!"))
	})

	err := http.ListenAndServe(":1234", corsMiddleware(mux))
	if err != nil {
		return err
	}

	return nil
}
