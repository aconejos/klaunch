package main

import (
	"fmt"
	"os/exec"
	"os"
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
				cmd := exec.Command("docker-compose", "up", "-d")
				// show the contect of the command before running it
				fmt.Println(cmd.String())
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error starting docker-compose:", err)
				} else {
					fmt.Println("Docker-compose started successfully!")
				}
			}
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
		case "stop":
			// Call stop function (you'll need to implement this)
			fmt.Println("Stopping klaunch...")
		case "delete-topic":
			// Call delete function (you'll need to implement this)
			fmt.Println("Deleting topic...")
		case "list-topics":
			// Call list function (you'll need to implement this)
			fmt.Println("Listing topics...")
		default:
			fmt.Println("Invalid command. Possible commands: start, create, stop, delete, list")
		}
	} else {
		fmt.Println("Please provide a command: start, stop, create-topic, delete-topic, list-topics")
	}

}
