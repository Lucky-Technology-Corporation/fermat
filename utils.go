package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
)

func downloadFileFromURL(url string, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func runDockerCompose() error {
	cmd := exec.Command("docker-compose", "up", "-d")
	return cmd.Run()
}
