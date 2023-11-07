package main

import (
	"log"
	"net/http"
)

func restartDockerContainer(name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		runner := &CommandRunner{}
		runner.Run("docker", "compose", "restart", name)

		if runner.err != nil {
			log.Println("Error restarting "+name+":", runner.err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(name + " restarted successfully!"))
		return
	}
}
