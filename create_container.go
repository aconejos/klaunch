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

	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	data, err := json.Marshal(file)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create topic: %s", resp.Status)
	}

	return nil
}