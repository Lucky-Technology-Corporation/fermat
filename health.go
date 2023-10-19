package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
		interval = 60
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
		fmt.Println("Error fetching Docker PS data:", err)
		return
	}

	// Convert the containers to JSON
	data, err := json.Marshal(containers)
	if err != nil {
		fmt.Println("Error marshalling Docker PS data:", err)
		return
	}

	// Create the request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response status for debugging purposes
	fmt.Println("Response status:", resp.Status)
}

// HealthServiceHandler is an HTTP handler that responds with the status of Docker containers
// running on the system in JSON format.
func HealthServiceHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := GetDockerPS()
	if err != nil {
		log.Printf("failed to run docker ps: %s \n", err)
		http.Error(w, "failed to run docker ps", http.StatusInternalServerError)
	}

	err = WriteJSONResponse(w, containers)
	if err != nil {
		log.Printf("failed to write json response: %s \n", err)
		http.Error(w, "failed to write json response", http.StatusInternalServerError)
	}
}

// DockerContainer represents details about a running Docker container
type DockerContainer struct {
	ContainerID string
	Image       string
	Command     string
	Created     string
	Status      string
	Ports       string
	Names       string
	Health      HealthStatus
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