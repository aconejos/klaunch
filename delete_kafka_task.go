package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func delete_tasks_interactive(interactive bool) error {
	// Get all defined connectors
	connectorNames, err := getConnectorNames()
	if err != nil {
		fmt.Println("Error parsing connector names:", err)
		return err
	}

	if len(connectorNames) == 0 {
		fmt.Println("No connectors found.")
		return nil
	}

	var connectorsToDelete []string

	if interactive {
		// Display connector selection menu
		fmt.Println("\nAvailable connectors:")
		for i, name := range connectorNames {
			fmt.Printf("%d. %s\n", i+1, name)
		}
		fmt.Printf("%d. Delete ALL connectors\n", len(connectorNames)+1)
		fmt.Printf("0. Cancel\n")

		fmt.Printf("\nSelect connectors to delete (comma-separated numbers, e.g., 1,3,5 or 'all'): ")
		var input string
		fmt.Scanln(&input)

		if input == "0" || input == "" {
			fmt.Println("Operation cancelled.")
			return nil
		}

		if input == "all" || input == strconv.Itoa(len(connectorNames)+1) {
			// Delete all connectors and all topics
			fmt.Println("Deleting ALL connectors and ALL topics...")
			return delete_all_tasks()
		} else {
			// Parse individual selections
			selections := strings.Split(input, ",")
			for _, sel := range selections {
				sel = strings.TrimSpace(sel)
				idx, err := strconv.Atoi(sel)
				if err != nil || idx < 1 || idx > len(connectorNames) {
					fmt.Printf("Invalid selection: %s\n", sel)
					continue
				}
				connectorsToDelete = append(connectorsToDelete, connectorNames[idx-1])
			}

			if len(connectorsToDelete) == 0 {
				fmt.Println("No valid connectors selected.")
				return nil
			}
		}
	} else {
		// Delete all connectors without interaction
		connectorsToDelete = connectorNames
	}

	// Delete selected connectors and their associated topics
	fmt.Printf("Deleting %d connectors and their associated topics...\n", len(connectorsToDelete))
	
	for _, connector := range connectorsToDelete {
		// Get associated topics for this connector
		associatedTopics, err := getConnectorTopics(connector)
		if err != nil {
			fmt.Printf("Warning: Could not get topics for connector %s: %v\n", connector, err)
		}

		// Delete the connector
		err = delete_single_connector(connector)
		if err != nil {
			fmt.Printf("Error deleting connector %s: %v\n", connector, err)
			continue
		}
		fmt.Printf("✅ Deleted connector: %s\n", connector)

		// Delete associated topics
		for _, topic := range associatedTopics {
			err = delete_single_topic(topic)
			if err != nil {
				fmt.Printf("Error deleting topic %s: %v\n", topic, err)
			} else {
				fmt.Printf("✅ Deleted topic: %s\n", topic)
			}
		}
	}

	return nil
}

func delete_all_tasks() error {
	fmt.Println("Deleting all connectors...")
	if err := delete_connectors(); err != nil {
		fmt.Println("Error deleting connectors:", err)
		return err
	}

	fmt.Println("Deleting all topics...")
	if err := delete_topics(); err != nil {
		fmt.Println("Error deleting topics:", err)
		return err
	}

	return nil
}

func delete_connectors() error {
	connectorsList, err := list_connectors(false)
	if err != nil {
		fmt.Println("Error listing connectors:", err)
		return err
	}

	// removing brackets from output
	cleanConnectorList := strings.TrimPrefix(strings.TrimSuffix(string(connectorsList), "]"), "[")

	// break the content of cleanConnectorList into a slice of strings
	connectorList := strings.Split(cleanConnectorList, ",")

	for _, connector := range connectorList {
		// remove double quotes from connector name
		connector = strings.Trim(connector, "\"")
		if connector == "" {
			return nil
		}
		
		err := delete_single_connector(connector)
		if err != nil {
			fmt.Printf("Error deleting connector %s: %v\n", connector, err)
		} else {
			fmt.Printf("✅ Deleted connector: %s\n", connector)
		}
	}
	return nil
}

func delete_topics() error {
	topicList, err := list_topics()
	if err != nil {
		fmt.Println("Error listing topics:", err)
		return err
	}

	if len(topicList) == 0 {
		fmt.Println("No user topics found to delete.")
		return nil
	}

	for _, topic := range topicList {
		err := delete_single_topic(topic)
		if err != nil {
			fmt.Printf("Error deleting topic %s: %v\n", topic, err)
		} else {
			fmt.Printf("✅ Deleted topic: %s\n", topic)
		}
	}

	return nil
}

func delete_single_connector(connectorName string) error {
	url := fmt.Sprintf("http://localhost:8083/connectors/%s", connectorName)
	
	fmt.Printf("Deleting connector: %s\n", connectorName)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete connector: %s", resp.Status)
	}

	return nil
}

func delete_single_topic(topicName string) error {
	fmt.Printf("Deleting topic: %s\n", topicName)
	
	deleteCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-topics", 
		"--delete", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091", 
		"--topic", topicName)
	
	output, err := deleteCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete topic: %v, output: %s", err, string(output))
	}

	return nil
}

func getConnectorNames() ([]string, error) {
	connectorsList, err := list_connectors(false)
	if err != nil {
		return nil, err
	}

	// Parse connector names from JSON array
	cleanConnectorList := strings.TrimPrefix(strings.TrimSuffix(string(connectorsList), "]"), "[")
	if cleanConnectorList == "" {
		return []string{}, nil
	}

	connectorList := strings.Split(cleanConnectorList, ",")
	var connectorNames []string
	
	for _, connector := range connectorList {
		connector = strings.Trim(strings.TrimSpace(connector), "\"")
		if connector != "" {
			connectorNames = append(connectorNames, connector)
		}
	}

	return connectorNames, nil
}

func getConnectorTopics(connectorName string) ([]string, error) {
	// Get connector configuration to find associated topics
	configCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", 
		fmt.Sprintf("http://localhost:8083/connectors/%s/config", connectorName))
	output, err := configCmd.Output()
	if err != nil {
		return []string{}, err
	}

	configStr := string(output)
	var topics []string

	// Look for "topics" field in the configuration
	if strings.Contains(configStr, "\"topics\"") {
		// Extract topics from configuration
		lines := strings.Split(configStr, ",")
		for _, line := range lines {
			if strings.Contains(line, "\"topics\"") {
				// Extract the topic value
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					topicValue := strings.Trim(strings.TrimSpace(parts[1]), "\"")
					if topicValue != "" {
						topics = append(topics, topicValue)
					}
				}
			}
		}
	}

	return topics, nil
}