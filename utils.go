package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// downloadFileFromURL downloads the given file to the local file system
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

// runDockerCompose runs `docker compose down` and `docker compose up -d`
func runDockerCompose() error {
	down := exec.Command("docker", "compose", "down")
	_ = down.Run()

	up := exec.Command("docker", "compose", "up", "-d")
	return up.Run()
}

// directoryExists checks if a directory exists at the given path
func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// executeGitCommand is a utility function to execute a git command and handle errors
func executeGitCommand(dir string, args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Git command failed: %s\nOutput: %s", err, out)
		return false
	}
	return true
}
