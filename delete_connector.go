package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

type ConnectorInfo struct {
	Name string `json:"name"`
}

func delete_connectors() error {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", "http://localhost:8083/connectors")
	output, err := listCmd.Output()
	if err != nil {
		fmt.Println("Error listing connectors:", err)
		return err
	}

	var connectors []ConnectorInfo
	err = json.Unmarshal(output, &connectors)
	if err != nil {
		fmt.Println("Error parsing connectors:", err)
		return err

	}

	
	for _, connector := range connectors {
		// curl DELETE  http://localhost:8083/connectors/mdb-kafka-connector-default
		url := "http://localhost:8083/connectors/"
		// concat url value with connector name
		url = url + connector.Name
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return err
		}

		defer resp.Body.Close()

	}
	return nil
}
