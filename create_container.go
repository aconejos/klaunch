package main

import (
	"bytes"
	"fmt"
	"os"
	"net/http"
)

func create_container() error {
	url := "http://localhost:8083/connectors"

	fmt.Println("Enter the path to the configuration file: ")
	var filePath string
	fmt.Scanln(&filePath)
	// if no file path is provided use "case_configs/default_topic.json" as default
	if filePath == "" {
		filePath = "./case_configs/default_topic.json"
		fmt.Println("Taking the default configuration file case_configs/default_topic.json")
	}

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

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create topic: %s", resp.Status)
	}

	return nil
}