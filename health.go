package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type HealthStatus string

const (
	Healthy   HealthStatus = "Healthy"
	Unhealthy HealthStatus = "Unhealthy"
	Stopped   HealthStatus = "Stopped"
	Unknown   HealthStatus = "Unknown"
)

// DetermineHealthStatus returns the enum status type based on the container's status string
func DetermineHealthStatus(status string) HealthStatus {
	if strings.HasPrefix(status, "Up ") {
		return Healthy
	} else if strings.HasPrefix(status, "Exited ") {
		return Stopped
	} else {
		return Unknown
	}
}

// HealthStatusServiceRunner periodically pings an endpoint with the status of Docker containers
// running on the system. The interval for pinging the endpoint and other configurations
// are read from environment variables.
func HealthStatusServiceRunner() {
	log.Println("Starting the HealthStatusServiceRunner...")

	intervalStr := os.Getenv("PING_INTERVAL_SECONDS")
	if intervalStr == "" {
		intervalStr = "60"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("ERROR: Invalid PING_INTERVAL_SECONDS value (%s). Must be an integer. Using default value.", intervalStr)
		interval = 10
	}

	endpoint := os.Getenv("HEALTH_CHECK_ENDPOINT_URL")
	if endpoint == "" {
		log.Println("ERROR: HEALTH_CHECK_ENDPOINT_URL is not set.")
		return
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("ERROR: API_KEY is not set.")
		return
	}

	log.Printf("Starting periodic health service with an interval of %d seconds.", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pingHealthStatus(endpoint, apiKey)
		}
	}
}

// pingHealthStatus sends the Docker container statuses to the specified endpoint.
// It constructs the POST request, sets appropriate headers and sends the request.
// Any issues encountered during the process are logged.
func pingHealthStatus(endpoint, apiKey string) {
	// Fetch docker ps details
	containers, err := GetDockerPS()
	if err != nil {
		log.Println("Error fetching Docker PS data:", err)
		return
	}

	certReady, err := CheckZeroSSLStatus()
	if err != nil {
		log.Printf("tls cert not set: %s \n", err)
		certReady = false
	}

	currentHealthStatus := VMHealth{
		Containers: containers,
		CertReady:  certReady,
	}

	// Convert the containers to JSON
	data, err := json.Marshal(currentHealthStatus)
	if err != nil {
		log.Println("Error marshalling Docker PS data:", err)
		return
	}

	// Create the request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// In case of a non 2xx response from the health endpoint debug print info
	if resp.StatusCode/100 != 2 {
		log.Println("Bad response status:", resp.Status)

		// Try reading the body and print any info
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			log.Println("Response body:", string(body))
		}

	}

}

// HealthServiceHandler is an HTTP handler that responds with the status of Docker containers
// running on the system in JSON format.
func HealthServiceHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := GetDockerPS()
	if err != nil {
		log.Printf("failed to run docker ps: %s \n", err)
		http.Error(w, "failed to run docker ps", http.StatusInternalServerError)
	}

	certReady, err := CheckZeroSSLStatus()
	if err != nil {
		log.Printf("tls cert not set: %s \n", err)
		certReady = false
	}

	currentHealthStatus := VMHealth{
		Containers: containers,
		CertReady:  certReady,
		Version:    VERSION,
	}

	err = WriteJSONResponse(w, currentHealthStatus)
	if err != nil {
		log.Printf("failed to write json response: %s \n", err)
		http.Error(w, "failed to write json response", http.StatusInternalServerError)
	}
}

// DockerContainer represents details about a running Docker container
type DockerContainer struct {
	ContainerID string       `json:"container_id"`
	Image       string       `json:"image"`
	Command     string       `json:"command"`
	Created     string       `json:"created"`
	Status      string       `json:"status"`
	Ports       string       `json:"ports"`
	Names       string       `json:"names"`
	Health      HealthStatus `json:"health"`
}

type VMHealth struct {
	Containers []DockerContainer `json:"containers"`
	CertReady  bool              `json:"cert_ready"`
	Version    string            `json:"version"`
}

// GetDockerPS fetches running Docker container details using the "docker ps" command.
// It returns a slice of DockerContainer structs representing each running container.
func GetDockerPS() ([]DockerContainer, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Image}}\t{{.Command}}\t{{.CreatedAt}}\t{{.Status}}\t{{.Ports}}\t{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var containers []DockerContainer
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 7 {
			continue
		}

		containers = append(containers, DockerContainer{
			ContainerID: parts[0],
			Image:       parts[1],
			Command:     parts[2],
			Created:     parts[3],
			Status:      parts[4],
			Ports:       parts[5],
			Names:       parts[6],
			Health:      DetermineHealthStatus(parts[4]),
		})
	}

	return containers, nil
}

func CheckZeroSSLStatus() (bool, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		return false, err
	}

	filePath := filepath.Join(currentDir, "zero_ssl/acme.json")

	sizeThreshold := int64(50 * 1024) // 50 KB

	fileSize, err := getFileSize(filePath)
	if err != nil {
		return false, err
	}

	return fileSize > sizeThreshold, nil
}

// getFileSize checks if a file exists and it's size. if a file does not exist, return 0 size (not an error)
func getFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	fileSize := fileInfo.Size()
	return fileSize, nil
}
