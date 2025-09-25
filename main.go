package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "klaunch",
		Short: "Klaunch manages Docker services for infrastructure components",
		Long:  `Klaunch is a CLI tool to manage Docker infrastructure components like starting, stopping, creating, deleting, and showing logs of tasks and topics.`,
	}

	var startCmd = &cobra.Command{
		Use:   "start [connectorVersion]",
		Short: "Starts klaunch with specified connector version",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting klaunch...")

			var connectorVersion string
			if len(args) >= 1 {
				connectorVersion = args[0]
			} else {
				connectorVersion = ""
			}

			if err := check_mongodb_running(); err != nil {
				fmt.Println("Error checking for running MongoDB:", err)
				fmt.Println("\nDISREGARD in case you are using Atlas as source or destination")
			}

			if err := check_connector_updates(connectorVersion); err != nil {
				fmt.Println("Error checking for connector updates:", err)
				fmt.Println("\nValidate available network Connection")
			}

			dockerCmd := exec.Command("open", "-a", "Docker")
			err := dockerCmd.Run()
			if err != nil {
				fmt.Println("Error: Docker daemon not running", err)
			} else {
				fmt.Println("Starting Docker daemon!")
				time.Sleep(3 * time.Second)
			}

			fmt.Println("Checking to pull docker images...(this can take a few minutes)")
			composeCmd := exec.Command("docker-compose", "-p", "klaunch", "up", "-d")
			err = composeCmd.Run()
			if err != nil {
				composeCmd = exec.Command("docker", "compose", "-p", "klaunch", "up", "-d")
				err = composeCmd.Run()
				if err != nil {
					fmt.Println("Error starting docker compose:", err)
				} else {
					fmt.Println("Klaunch docker compose started successfully!")
				}
			} else {
				fmt.Println("Klaunch docker-compose started successfully!")
			}
		}}

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops klaunch and removes all containers",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping klaunch...")

			listContainersCmd := exec.Command("docker-compose", "-p", "klaunch", "ps", "-aq")
			output, err := listContainersCmd.Output()
			if err != nil {
				listContainersCmd = exec.Command("docker", "compose", "-p", "klaunch", "ps", "-aq")
				err = listContainersCmd.Run()
				if err != nil {
					fmt.Println("Error listing container:", err)
				}
			}

			containerIDs := strings.Split(string(output), "\n")
			for _, id := range containerIDs {
				if id != "" {
					stopCmd := exec.Command("docker", "rm", "-f", id)
					err := stopCmd.Run()
					if err != nil {
						fmt.Println("Error deleting container:", err)
						return
					}
				}
			}
			fmt.Println("Containers removed successfully!")
		},
	}

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates a new Kafka task",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Creating new Kafka task...")
			if err := create_kafka_task(); err != nil {
				fmt.Println("Error creating new task:", err)
			} else {
				fmt.Println("New task created successfully!")
			}
		},
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete [all|connectors|topics]",
		Short: "Deletes connectors and/or topics with interactive selection",
		Long: `Delete connectors and topics with options:
  delete all        - Delete all connectors and topics (with confirmation)
  delete connectors - Interactive connector selection
  delete topics     - Interactive topic selection
  delete           - Interactive selection for both connectors and topics`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var target string
			if len(args) > 0 {
				target = args[0]
			}

			switch target {
			case "all":
				fmt.Printf("Are you sure you want to delete ALL connectors and topics? (y/N): ")
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
					fmt.Println("Operation cancelled.")
					return
				}
				
				if err := delete_all_tasks(); err != nil {
					fmt.Println("Error deleting all tasks:", err)
				} else {
					fmt.Println("All connectors and topics deleted successfully!")
				}

			case "connectors":
				fmt.Println("Interactive task deletion...")
				if err := delete_tasks_interactive(true); err != nil {
					fmt.Println("Error in task deletion:", err)
				}

			case "topics":
				fmt.Println("Interactive task deletion...")
				if err := delete_tasks_interactive(true); err != nil {
					fmt.Println("Error in task deletion:", err)
				}

			default:
				// Interactive mode for both
				fmt.Println("What would you like to delete?")
				fmt.Println("1. Delete all")
				fmt.Println("2. Select task") 
				fmt.Println("0. Cancel")
				fmt.Print("Select option (0-2): ")

				var choice string
				fmt.Scanln(&choice)

				switch choice {
				case "1":
					if err := delete_tasks_interactive(false); err != nil {
						fmt.Println("Error in task deletion:", err)
					}
				case "2":
					if err := delete_tasks_interactive(true); err != nil {
						fmt.Println("Error in task deletion:", err)
					}
				case "0":
					fmt.Println("Operation cancelled.")
				default:
					fmt.Println("Invalid option. Operation cancelled.")
				}
			}
		},
	}

	var showCmd = &cobra.Command{
		Use:   "show [components|messages]",
		Short: "Shows components or messages",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			componentOrMessage := args[0]
			if componentOrMessage == "components" {
				if err := list_components(verbose); err != nil {
					fmt.Println("Error listing components:", err)
				}
			} else if componentOrMessage == "messages" {
				if err := list_messages(); err != nil {
					fmt.Println("Error listing messages:", err)
				}
			} else {
				fmt.Println("Invalid component or message type. Please choose 'components' or 'messages'.")
			}
		},
	}
	
	showCmd.Flags().Bool("verbose", false, "Show full stack traces for failed tasks")

	var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Extracts logs from Kafka Connect",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Extracting logs...")
			datePrefix := time.Now().Format("20060102_150405")
			filename := fmt.Sprintf("logs/%s_kafka_connect.log", datePrefix)

			dockerCmd := exec.Command("docker", "logs", "kafka-connect")
			output, err := dockerCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Failed to execute command: %v\nOutput: %s\n", err, string(output))
				return
			}

			err = os.WriteFile(filename, output, 0644)
			if err != nil {
				fmt.Printf("Failed to write to file: %v\n", err)
				return
			}

			fmt.Printf("Logs saved to %s\n", filename)
		},
	}

	rootCmd.AddCommand(startCmd, stopCmd, createCmd, deleteCmd, showCmd, logsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
