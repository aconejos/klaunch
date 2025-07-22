package main

import (
	"bytes"
	"fmt"
	"os"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func create_kafka_task() error {
	url := "http://localhost:8083/connectors"

	// Get available config files
	configFiles, err := getConfigFiles()
	if err != nil {
		fmt.Println("Error reading config files:", err)
		return err
	}

	var filePath string
	if len(configFiles) == 0 {
		fmt.Println("No configuration files found in case_configs directory")
		fmt.Println("Enter the path to the configuration file: ")
		fmt.Scanln(&filePath)
	} else {
		// Display available config files
		fmt.Println("Available configuration files:")
		for i, file := range configFiles {
			fmt.Printf("%d. %s\n", i+1, file)
		}
		fmt.Printf("%d. Select custom path or press Enter for default\n", len(configFiles)+1)
		fmt.Printf("\nSelect a configuration file (1-%d): ", len(configFiles)+1)
		
		var choice string
		fmt.Scanln(&choice)
		
		choiceNum, err := strconv.Atoi(choice)
		if err != nil || choiceNum < 1 || choiceNum > len(configFiles)+1 {
			fmt.Println("Invalid choice. Using default configuration.")
			filePath = "./case_configs/default_topic.json"
		} else if choiceNum == len(configFiles)+1 {
			// Custom path option
			fmt.Println("Enter the path to the configuration file: ")
			fmt.Scanln(&filePath)
		} else {
			// Selected from available files
			filePath = filepath.Join("./case_configs", configFiles[choiceNum-1])
		}
	}

	// if no file path is provided use "case_configs/default_topic.json" as default
	if filePath == "" {
		filePath = "./case_configs/default_topic.json"
		fmt.Println("Taking the default configuration file case_configs/default_topic.json")
	}

	fmt.Printf("Selected configuration file: %s\n", filePath)
	
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	// show the content of file
	fmt.Println("Using the following configuration:")
	fmt.Println(string(file))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(file))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create task: %s", resp.Status)
	}

	return nil
}

func getConfigFiles() ([]string, error) {
	configDir := "./case_configs"
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var configFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			configFiles = append(configFiles, file.Name())
		}
	}

	return configFiles, nil
}