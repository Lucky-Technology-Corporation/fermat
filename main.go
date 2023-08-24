package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	downloadURL := "https://storage.googleapis.com/swizzle_scripts_pub/docker-compose.yaml"
	err := downloadFileFromURL(downloadURL, "docker-compose.yaml")
	if err != nil {
		log.Fatalf("Failed to download docker-compose file: %v", err)
	}

	err = runDockerCompose()
	if err != nil {
		log.Fatalf("Failed to run docker-compose: %v", err)
	}

	http.Handle("/editor/", proxyPass("8080"))
	http.Handle("/runner/", proxyPass("4411"))
	http.Handle("/database/", proxyPass("27017"))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on :8080")
	server.ListenAndServe()
}
