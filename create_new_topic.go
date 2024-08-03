package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func create_new_topic() error {
	url := "http://localhost:8083/connectors"

	fmt.Print("Enter the path to the configuration file: ")
	var filePath string
	fmt.Scanln(&filePath)

	file, err := ioutil.ReadFile(filePath)
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