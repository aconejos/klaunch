package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)


func delete_connectors() error {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", "http://localhost:8083/connectors")
	connectorsList, err := listCmd.Output()
	if err != nil {
		fmt.Println("Error listing connectors:", err)
		return err
	}

	// removing brackets from output
	cleanConnectorList := strings.TrimPrefix(strings.TrimSuffix(string(connectorsList), "]"), "[")

	// break the content of cleanConnectorList into a slide of strings
	connectorList := strings.Split(cleanConnectorList, ",")
	
	for _, connector := range connectorList {
		// curl DELETE  http://localhost:8083/connectors/mdb-kafka-connector-default
		url := "http://localhost:8083/connectors/"

		// remove double quotes from connector name
		connector = strings.Trim(connector, "\"")
		if connector == "" {
			return nil	
		}
		// concat url value with connector name
		url = url + connector

		// print url value
		println("Deleting: " + url)

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
