package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"net/http"
)

func create_container() error {
	url := "http://localhost:8083/connectors"

	fmt.Print("Enter the path to the configuration file: ")
	var filePath string
	fmt.Scanln(&filePath)
	// if no file path is provided use "case_configs/default_topic.json" as default
	if filePath == "" {
		filePath = "./case_configs/default_topic.json"
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	data, err := json.Marshal(file)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// display the request
	//fmt.Println("Request:")
	//fmt.Println(req)
	//fmt.Println("Request body:")
	//fmt.Println(string(data))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create topic: %s", resp.Status)
	}

	return nil
}