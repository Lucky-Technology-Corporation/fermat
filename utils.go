package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
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

// downloadFileFromGoogleBucket is a quick downloader script to grab the docker compose and any other deps
func downloadFileFromGoogleBucket(bucketName, objectName string) ([]byte, error) {
	ctx := context.Background()

	log.Println("... creating google bucket connection ...")

	client, err := storage.NewClient(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	log.Println("... set object ...")

	r, err := obj.NewReader(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read object: %v", err)
	}
	defer r.Close()

	log.Println("... pulling object ...")

	data, err := io.ReadAll(r)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read data: %v", err)
	}

	return data, nil
}

// SaveBytesToFile saves a given []byte to a specified filename.
func saveBytesToFile(filename string, data []byte) error {
	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(data)
	return err
}

// loadDockerImageFromTarball is a helper function that will docker load -i [tarball] and logs any output
func loadDockerImageFromTarball(tarballPath string) error {
	cmd := exec.Command("docker", "load", "-i", tarballPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Printf("Command output: %s", output)
	return nil
}
