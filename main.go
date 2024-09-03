package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {

	// show componentOrMessage value
	if len(os.Args) < 2 {
		fmt.Println("Invalid command. Possible commands: start, stop, create, delete, show, logs, help")
		return
	}
	// Check for input parameters
	command := os.Args[1]
	switch command {
	case "start":
		fmt.Println("Starting klaunch...")

		var connectorVersion string
		if len(os.Args) >= 3 {
			connectorVersion = os.Args[2]
		} else {
			connectorVersion = ""
		}

		mognodbAvailable := check_mongodb_running()
		if mognodbAvailable != nil {
			fmt.Println("Error checking for running MongDB:", mognodbAvailable)
		}
		// Call functions from check_connector_updates.go
		updateAvailable := check_connector_updates(connectorVersion)
		if updateAvailable != nil {
			fmt.Println("Error checking for connector updates:", updateAvailable)
		}

		// Spin up docker-compose
		if updateAvailable == nil {
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
			cmd = exec.Command("docker-compose", "-p", "klaunch", "up", "-d")
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
		listContainersCmd := exec.Command("docker-compose", "-p", "klaunch", "ps", "-aq")

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
	case "show":
		if len(os.Args) != 3 {
			fmt.Println("Invalid show command. Please choose 'components' or 'messages'.")
			return
		}
		componentOrMessage := os.Args[2]
		//if "components" list_component() if "messages" list_messages()
		if componentOrMessage == "components" {
			// Call list function
			err := list_components()
			if err != nil {
				fmt.Println("Error listing components:", err)
				return
			}
		} else if componentOrMessage == "messages" {
			// Call list function
			err := list_messages()
			if err != nil {
				fmt.Println("Error listing messages:", err)
				return
			}
		} else {
			fmt.Println("Invalid component or message type. Please choose 'components' or 'messages'.")
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
	case "help":
		fmt.Println(" - start [connector version]: Creates a Docker compose with all the necesary infrastrusture components. By default will connect to the release repository and download the latest version of MongoDB Kafka Connect. ")
		fmt.Println(" - stop: Deletes the Docker compose components completely. ")
		fmt.Println(" - create: Creates a connector Task based on an input config file path.(json format) ")
		fmt.Println(" - delete: Deletes all existing Tasks and topics. Infrastructure remains. ")
		fmt.Println(" - show [components - messages] ")
		fmt.Println("      Components: will list running Tasks and existing Topics. ")
		fmt.Println("      Messages: will list existing Topics and will create a consumer process to display messages on the console. ")
		fmt.Println(" - logs: Will dump a the Kafka connect log file into $repository/logs path with the following format: $timestamps_kadka_connect.log ")

	default:
		fmt.Println("Invalid command. Possible commands: start, stop, create, delete, show, logs, help")
	}

}