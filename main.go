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
			// Call start function (you'll need to implement this)
			fmt.Println("Starting klaunch...")

			// Call functions from check_connector_updates.go
			updateAvailable := check_connector_updates()
			if updateAvailable != nil {
				fmt.Println("Error checking for connector updates:", updateAvailable)
			} 
			
			// Spin up docker-compose
			if updateAvailable == nil  {
				//cmd := exec.Command("docker-compose", "up", "-d")
				cmd := exec.Command("open", "-a", "Docker")
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error docker daemon not running", err)
				} else {
					fmt.Println("Starting docker daemon!")
					time.Sleep(3 * time.Second)
				}

				
				cmd = exec.Command("docker-compose", "up", "-d")
				err = cmd.Run()
				if err != nil {
					fmt.Println("Error starting docker-compose:", err)
				} else {
					fmt.Println("Docker-compose started successfully!")
				}
			}	
		case "stop":
			// Call stop function
			fmt.Println("Stopping klaunch...")

			// list all running containers remove them
			listContainersCmd := exec.Command("docker", "ps", "-aq")
			
			output, err := listContainersCmd.Output()
			if err != nil {
				fmt.Println("Error listing containers:", err)
				return
			}
			
			containerIDs := strings.Split(string(output), "\n")
			for _, id := range containerIDs {
				if id != "" {
					// stop the container
					stopCmd := exec.Command("docker", "rm", "-f", id)
					err := stopCmd.Run()
					if err != nil {
						fmt.Println("Error stopping container:", err)
						return
					}
				}
			}
			fmt.Println("Containers removed successfully!")
		case "create-topic":
			// Call create function (you'll need to implement this)
			fmt.Println("Creating new topic...")
				// Call create_new_topic function (you'll need to implement this)
			err := create_new_topic()
			if err != nil {
				fmt.Println("Error creating new topic:", err)
			} else {
				fmt.Println("New topic created successfully!")
			}
		case "delete-topic":
			// Call delete function (you'll need to implement this)
			fmt.Println("Deleting topic...")
		case "list-topics":
			// Call list function (you'll need to implement this)
			fmt.Println("Listing topics...")
		default:
			fmt.Println("Invalid command. Possible commands: start, stop, create-topic, delete-topic, list-topics")
		}
	} else {
		fmt.Println("Please provide a command: start, stop, create-topic, delete-topic, list-topics")
	}

}
