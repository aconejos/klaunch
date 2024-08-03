package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type ExcludedTopic struct {
	Name string
}

var excludedTopics = []ExcludedTopic{
	{"__consumer_offsets"},
	{"_confluent-command"},
	{"_confluent-telemetry-metrics"},
	{"_confluent_balancer_api_state"},
	{"_confluent_balancer_broker_samples"},
	{"_confluent_balancer_partition_samples"},
	{"_schemas"},
	{"docker-connect-configs"},
	{"docker-connect-offsets"},
	{"docker-connect-status"},
}


func list_components() error {
	err := list_connectors() 
	if err != nil {
		return err
	}
	err = list_topics()
	if err != nil {
		return err
	}
	return nil
}


func list_connectors() error {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", "http://localhost:8083/connectors")
	output, err := listCmd.Output()
	// show the content of output
	fmt.Println("List of connectors:")
	println(string(output))
	if err != nil {
		return err
	}
	return nil
}

func list_topics() error {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-topics", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091", "--list")
	output, err := listCmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var topics []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !isExcludedTopic(line) {
			topics = append(topics, line)
		}
	}
	if topics == nil {
		fmt.Println("No topics created. Remember only topics with at least 1 message will be listed.")
		return nil
	}

	for _, topic := range topics {
		fmt.Println("List of topics:")
		fmt.Println(topic)
	}
	return nil
}


func isExcludedTopic(topic string) bool {
	for _, excluded := range excludedTopics {
		if topic == excluded.Name {
			return true
		}
	}
	return false
}