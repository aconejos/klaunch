package main

import (
	"fmt"
)

func main() {
	// Call functions from check_docker_daemon.go
	daemonStatus := check_docker_daemon()
	if daemonStatus != nil {
		fmt.Println("Docker daemon error status:", daemonStatus)
	}

	// Call functions from check_connector_updates.go
	updateAvailable := check_connector_updates()
	if updateAvailable != nil {
		fmt.Println("Error checking for connector updates:", updateAvailable)
	} 
}
