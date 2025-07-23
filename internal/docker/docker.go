package docker

import (
	"os/exec"
	"strings"
)

// StartDocker opens the Docker application
func StartDocker() error {
	dockerCmd := exec.Command("open", "-a", "Docker")
	return dockerCmd.Run()
}

// StartCompose starts docker-compose services
func StartCompose() error {
	composeCmd := exec.Command("docker-compose", "-p", "klaunch", "up", "-d")
	err := composeCmd.Run()
	if err != nil {
		// Try with docker compose (newer syntax)
		composeCmd = exec.Command("docker", "compose", "-p", "klaunch", "up", "-d")
		return composeCmd.Run()
	}
	return nil
}

// ListContainers lists all containers for the klaunch project
func ListContainers() ([]string, error) {
	listContainersCmd := exec.Command("docker-compose", "-p", "klaunch", "ps", "-aq")
	output, err := listContainersCmd.Output()
	if err != nil {
		// Try with docker compose (newer syntax)
		listContainersCmd = exec.Command("docker", "compose", "-p", "klaunch", "ps", "-aq")
		output, err = listContainersCmd.Output()
		if err != nil {
			return nil, err
		}
	}

	containerIDs := strings.Split(string(output), "\n")
	var validIDs []string
	for _, id := range containerIDs {
		if id != "" {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs, nil
}

// RemoveContainer removes a specific container by ID
func RemoveContainer(containerID string) error {
	stopCmd := exec.Command("docker", "rm", "-f", containerID)
	return stopCmd.Run()
}

// GetContainerLogs gets logs from a specific container
func GetContainerLogs(containerName string) ([]byte, error) {
	dockerCmd := exec.Command("docker", "logs", containerName)
	return dockerCmd.CombinedOutput()
}
