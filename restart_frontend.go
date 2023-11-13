package main

import (
	"fmt"
	"log"
	"net/http"
)

func restartDockerContainerHandler(name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := restartDockerContainer(name); err != nil {
			log.Println("Error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(name + " restarted successfully!"))
		return
	}
}

func restartDockerContainer(container string) error {
	runner := &CommandRunner{}
	runner.Run("docker", "compose", "restart", container)

	if runner.err != nil {
		return fmt.Errorf("Error restarting %s: %v", container, runner.err)
	}

	return nil
}
