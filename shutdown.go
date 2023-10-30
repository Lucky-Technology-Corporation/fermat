package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func ShutdownHandler(shutdownChan chan bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		runner := &CommandRunner{}
		runner.Run("docker", "compose", "down")
		if runner.err != nil {
			log.Println("Couldn't shutdown stuck on docker compose down:", runner.err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		commitMessage := fmt.Sprintf("swizzle commit shutdown: %s", time.Now().Format(time.RFC3339))

		runner = &CommandRunner{dir: "code"}
		runner.Run("git", "add", ".")
		runner.Run("git", "commit", "-m", commitMessage)
		runner.Run("git", "push", "-o", "nokeycheck", "origin", "master")

		w.WriteHeader(http.StatusOK)

		log.Println("Signaling to shutdown server...")

		go func() {
			shutdownChan <- true
		}()
	}
}
