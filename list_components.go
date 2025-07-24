package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type ExcludedTopic struct {
	Name string
}

type TaskStatus struct {
	ID     int    `json:"id"`
	State  string `json:"state"`
	Worker string `json:"worker_id"`
	Trace  string `json:"trace,omitempty"`
}

type ConnectorStatus struct {
	Name      string                 `json:"name"`
	Connector map[string]interface{} `json:"connector"`
	Tasks     []TaskStatus           `json:"tasks"`
}

// default topics to exclude from the list
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

func list_components(verbose bool) error {
	_, err := list_connectors(verbose)
	if err != nil {
		return err
	}
	
	_, err = list_topics()
	if err != nil {
		return err
	}
	return nil
}

func format_task_output(name string, tasks []TaskStatus, verbose bool) {
	if len(tasks) == 0 {
		fmt.Printf("│   └── No tasks\n")
		return
	}

	for i, task := range tasks {
		prefix := "├──"
		if i == len(tasks)-1 {
			prefix = "└──"
		}
		
		// Color-code based on state
		var stateDisplay string
		switch strings.ToUpper(task.State) {
		case "RUNNING":
			stateDisplay = fmt.Sprintf("✅ %s", task.State)
		case "FAILED":
			stateDisplay = fmt.Sprintf("❌ %s", task.State)
		case "PAUSED":
			stateDisplay = fmt.Sprintf("⏸️  %s", task.State)
		default:
			stateDisplay = fmt.Sprintf("⚪ %s", task.State)
		}
		
		fmt.Printf("│   %s Task %d: %s (worker: %s)\n", 
			prefix, task.ID, stateDisplay, task.Worker)
		
		// Show error details for failed tasks
		if strings.ToUpper(task.State) == "FAILED" && task.Trace != "" {
			errorMsg := extractErrorMessage(task.Trace, verbose)
			if errorMsg != "" {
				if verbose {
					fmt.Printf("│       └── %s", errorMsg)
				} else {
					fmt.Printf("│       └── Error: %s\n", errorMsg)
					fmt.Printf("│       └── Use './klaunch show --verbose' for full stack trace\n")
				}
			}
		}
	}
}

func extractErrorMessage(trace string, verbose bool) string {
	if trace == "" {
		return ""
	}
	
	// If verbose mode, return formatted full stack trace
	if verbose {
		lines := strings.Split(trace, "\n")
		var formattedTrace strings.Builder
		formattedTrace.WriteString("Full Stack Trace:\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			formattedTrace.WriteString("       ")
			formattedTrace.WriteString(line)
			formattedTrace.WriteString("\n")
		}
		
		return formattedTrace.String()
	}
	
	// Non-verbose mode: extract short error message
	lines := strings.Split(trace, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Look for common error patterns
		if strings.Contains(line, "Exception:") || 
		   strings.Contains(line, "Error:") ||
		   strings.Contains(line, "Caused by:") {
			// Extract just the error message part
			if idx := strings.Index(line, ": "); idx != -1 {
				return strings.TrimSpace(line[idx+2:])
			}
			return line
		}
		
		// If no exception pattern found, return first non-empty line
		if !strings.HasPrefix(line, "at ") && !strings.HasPrefix(line, "\tat ") {
			return line
		}
	}
	
	// Fallback: return first line if no pattern matched
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	
	return "Unknown error"
}

func list_connector_status(connectorName string) (*ConnectorStatus, error) {
	statusCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", 
		fmt.Sprintf("http://localhost:8083/connectors/%s/status", connectorName))
	output, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get status for connector %s: %v", connectorName, err)
	}

	var status ConnectorStatus
	err = json.Unmarshal(output, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status JSON for connector %s: %v", connectorName, err)
	}

	return &status, nil
}

func list_connectors(verbose bool) (string, error) {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "curl", "-s", "http://localhost:8083/connectors")
	output, err := listCmd.Output()
	if err != nil {
		return "", err
	}

	// Parse connector names from JSON array
	var connectorNames []string
	err = json.Unmarshal(output, &connectorNames)
	if err != nil {
		fmt.Println("List of connectors:")
		println(string(output))
		return string(output), nil
	}

	// Get detailed status for each connector
	fmt.Println("Connectors and Tasks:")
	for _, name := range connectorNames {
		status, err := list_connector_status(name)
		if err != nil {
			fmt.Printf("├── %s [ERROR: %v]\n", name, err)
			continue
		}

		// Get connector state
		connectorState := "UNKNOWN"
		if status.Connector != nil {
			if state, ok := status.Connector["state"].(string); ok {
				connectorState = state
			}
		}

		fmt.Printf("├── %s [%s]\n", name, connectorState)
		
		// Use the new formatting function
		format_task_output(name, status.Tasks, verbose)
	}

	return string(output), nil
}

func list_topics() ([]string, error) {
	listCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-topics", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091", "--list")
	output, err := listCmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var topics []string

	// Exclude topics that are in the excludedTopics slice
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !isExcludedTopic(line) {
			topics = append(topics, line)
		}
	}
	if topics == nil {
		fmt.Println("No topics created. Remember only topics with at least 1 message will be listed.")
		return nil, nil
	}

	return topics, nil
}

func isExcludedTopic(topic string) bool {
	for _, excluded := range excludedTopics {
		if topic == excluded.Name {
			return true
		}
	}
	return false
}
