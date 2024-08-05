package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {

	// Check for input parameters
	if len(os.Args) > 1 {
		command := os.Args[1]

		switch command {
		case "start":
			fmt.Println("Starting klaunch...")

			// Call functions from check_connector_updates.go
			updateAvailable := check_connector_updates()
			if updateAvailable != nil {
				fmt.Println("Error checking for connector updates:", updateAvailable)
			} 
			
			// Spin up docker-compose
			if updateAvailable == nil  {
				cmd := exec.Command("open", "-a", "Docker")
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error docker daemon not running", err)
				} else {
					fmt.Println("Starting docker daemon!")
					// 	give docker time to start
					time.Sleep(3 * time.Second)
				}

				// Start docker-compose
				cmd = exec.Command("docker-compose","-p", "klaunch", "up", "-d")
				err = cmd.Run()
				if err != nil {
					fmt.Println("Error starting docker-compose:", err)
				} else {
					fmt.Println("Klaunch docker-compose started successfully!")
				}
			}	
		case "stop":
			// Call stop function
			fmt.Println("Stopping klaunch...")

			// list all running containers 
			listContainersCmd := exec.Command("docker-compose","-p", "klaunch", "ps", "-aq")
			
			output, err := listContainersCmd.Output()
			if err != nil {
				fmt.Println("Error listing containers:", err)
				return
			}
			
			containerIDs := strings.Split(string(output), "\n")
			// remove all the containers
			for _, id := range containerIDs {
				if id != "" {
					// stop the container
					stopCmd := exec.Command("docker", "rm", "-f", id)
					err := stopCmd.Run()
					if err != nil {
						fmt.Println("Error deleting container:", err)
						return
					}
				}
			}
			fmt.Println("Containers removed successfully!")
		case "create":
			fmt.Println("Creating new container...")
			// Call create function 
			err := create_container()
			if err != nil {
				fmt.Println("Error creating new topic:", err)
			} else {
				fmt.Println("New topic created successfully!")
			}
		case "delete":
			// Call delete function (you'll need to implement this)
			fmt.Println("Deleting containers...")
			fmt.Println("Deleting topic...")
		case "list":
			// Call list function
			err := list_components()
			if err != nil {
				fmt.Println("Error listing topic:", err)
			}
		default:
			fmt.Println("Invalid command. Possible commands: start, stop, create, delete, list")
		}
	} else {
		fmt.Println("Please provide a command: start, stop, create, delete, list")
	}

}
