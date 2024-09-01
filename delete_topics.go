package main

import (
	"fmt"
	"os/exec"
)

func delete_topics() error {
	topicList, err := list_topics()
	if err != nil {
		fmt.Println("Error listing connectors:", err)
		return err
	}

	for _, topic := range topicList {
		deleteCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-topics", "--delete", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091", "--topic", topic)
		err := deleteCmd.Run()
		if err != nil {
			return err
		}

	}

	return nil
}
