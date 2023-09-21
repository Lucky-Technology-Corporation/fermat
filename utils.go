package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/storage"
)

const SECRETS_JSON = "SECRETS_JSON"

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

// Runs: gcloud auth configure-docker us-central1-docker.pkg.dev --quiet
// This allows the docker compose to reference an image on the artifact registry.
func setupArtifactRegistryAuth() error {
	return exec.Command("gcloud", "auth", "configure-docker", "us-central1-docker.pkg.dev", "--quiet").Run()
}

// runDockerCompose runs `docker compose down`, `docker compose pull` and `docker compose up -d`,
// and logs the output to docker_compose_fermat.log
func runDockerCompose() error {
	logFile, err := os.OpenFile("docker_compose_fermat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	commands := [][]string{
		{"docker", "compose", "down"},
		{"docker", "compose", "pull"},
		{"docker", "compose", "up", "-d"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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

// downloadFileFromGoogleBucket downloads a file from a Google Cloud Storage bucket
// and writes it directly to a given destination on disk.
func downloadFileFromGoogleBucket(bucketName, objectName, destinationPath string) error {
	ctx := context.Background()

	log.Println("... creating google bucket connection ...")

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	log.Println("... set object ...")

	r, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to read object: %v", err)
	}
	defer r.Close()

	log.Println("... pulling object ...")

	// Create or truncate the destination file
	out, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Stream the content directly to the file
	if _, err = io.Copy(out, r); err != nil {
		return fmt.Errorf("failed to write data to file: %v", err)
	}

	return nil
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
	cmd := exec.Command("docker", "load", "--input", tarballPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Printf("Command output: %s", output)
	return nil
}

func decryptAES(ciphertext, key, iv []byte) ([]byte, error) {
	// Create a new AES cipher block with the given key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Check that the length of the ciphertext is a multiple of the block size
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// Create a new Cipher Blocker using the block and IV
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext in-place
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpadding (removing PKCS#7 padding)
	padLength := int(ciphertext[len(ciphertext)-1])
	unpaddedData := ciphertext[:len(ciphertext)-padLength]

	return unpaddedData, nil
}

func saveInitialSecrets() error {
	secretsJson, exists := os.LookupEnv(SECRETS_JSON)
	if !exists {
		log.Println("[Warning] " + SECRETS_JSON + " environment variable is not set.")
		return nil
	}

	secrets, err := ReadSecrets(strings.NewReader(secretsJson))
	if err != nil {
		return err
	}
	return secrets.SaveSecretsToFile(SECRETS_FILE_PATH)
}

func WriteJSONResponse(w http.ResponseWriter, data interface{}) error {
	// Marshal the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set content type and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

	return nil
}
