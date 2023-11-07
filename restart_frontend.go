package main

import (
	"log"
	"net/http"
)

func restartFrontend(w http.ResponseWriter, r *http.Request) {
	runner := &CommandRunner{}
	runner.Run("docker", "compose", "restart", "frontend")

	if runner.err != nil {
		log.Println("Error restarting frontend:", runner.err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Frontend restarted successfully!"))
	return
}
