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

func list_running_topics() error {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-topics", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091", "--list")
	output, err := listCmd.Output()
	// show the content of output
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
		fmt.Println("No topics deployed")
		return nil
	}

	for _, topic := range topics {
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