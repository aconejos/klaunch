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

			mognodbAvailable := check_mongodb_running()
			if mognodbAvailable != nil {
				fmt.Println("Error checking for running MongDB:", mognodbAvailable)
			} 
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
			// Call delete function
			fmt.Println("Deleting containers...")
			err := delete_connectors()
			if err != nil {
				fmt.Println("Error deleting topic:", err)
			} else {
				fmt.Println("Containers deleted successfully!")
			}
			fmt.Println("Deleting topic...")
			err = delete_topics()
			if err != nil {
				fmt.Println("Error deleting topic:", err)
			} else {
				fmt.Println("Topics deleted successfully!")
			}
		case "list":
			// Call list function
			err := list_components()
			if err != nil {
				fmt.Println("Error listing topic:", err)
			}
		case "logs":
			// Call logs function
			fmt.Println("Extracting logs...")
			// Get the current date and format it as YYYYMMDD_HHMMSS
			datePrefix := time.Now().Format("20060102_150405") // Format: YYYYMMDD_HHMMSS
				
			// Construct the filename with the date prefix
			filename := fmt.Sprintf("logs/%s_kafka_connect.log", datePrefix)
				
			// Define the command to get logs from the Docker container
			cmd := exec.Command("docker", "logs", "kafka-connect")
				
			// Execute the command and capture the output
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Failed to execute command: %v\nOutput: %s\n", err, string(output))
				return
			}
		
			// Write the output to a file
			err = os.WriteFile(filename, output, 0644)
			if err != nil {
				fmt.Printf("Failed to write to file: %v\n", err)
				return
			}

			fmt.Printf("Logs saved to %s\n", filename)
		default:
			fmt.Println("Invalid command. Possible commands: start, stop, create, delete, list, logs")
		}
	} else {
		fmt.Println("Please provide a command: start, stop, create, delete, list, logs")
	}

}
