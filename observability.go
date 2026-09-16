package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	observabilityPromFile      = "volumes/prometheus.yml"
	observabilityDefaultTarget = "kafka-connect:8091"
	observabilityPromURL       = "http://localhost:9090"
	observabilityGrafanaURL    = "http://localhost:3000"
)

// matches the whole kafka-connect job block up to (but not including) the next job or EOF.
var observabilityConnectJobRE = regexp.MustCompile(`(?s)(- job_name: ['"]?kafka-connect['"]?.*?targets:\s*\n\s*-\s*)([^\s\n]+)`)

func observability_show() error {
	target, err := readCurrentConnectTarget()
	if err != nil {
		return err
	}
	fmt.Printf("kafka-connect target: %s\n", target)
	fmt.Printf("Prometheus:           %s\n", observabilityPromURL)
	fmt.Printf("Grafana:              %s  (admin/foobar)\n", observabilityGrafanaURL)
	return nil
}

func observability_set(hostport string) error {
	if err := validateHostPort(hostport); err != nil {
		return err
	}
	if err := rewriteConnectTarget(hostport); err != nil {
		return err
	}
	fmt.Printf("Updated kafka-connect scrape target to %s\n", hostport)

	// ponytail: docker restart instead of prom lifecycle API; swap if reloads become frequent.
	restart := exec.Command("docker", "restart", "prometheus")
	out, err := restart.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart prometheus: %v\n%s", err, string(out))
	}
	fmt.Println("Prometheus restarted. Give it ~15s to scrape the new target.")
	return nil
}

// ponytail: line-rewrite instead of YAML parse; template it if a second field ever needs configuring.
func rewriteConnectTarget(hostport string) error {
	content, err := os.ReadFile(observabilityPromFile)
	if err != nil {
		return fmt.Errorf("read %s: %v", observabilityPromFile, err)
	}
	if !observabilityConnectJobRE.Match(content) {
		return fmt.Errorf("could not locate kafka-connect job in %s", observabilityPromFile)
	}
	updated := observabilityConnectJobRE.ReplaceAll(content, []byte("${1}"+hostport))
	return os.WriteFile(observabilityPromFile, updated, 0644)
}

func readCurrentConnectTarget() (string, error) {
	content, err := os.ReadFile(observabilityPromFile)
	if err != nil {
		return "", fmt.Errorf("read %s: %v", observabilityPromFile, err)
	}
	m := observabilityConnectJobRE.FindSubmatch(content)
	if m == nil {
		return "", fmt.Errorf("could not locate kafka-connect job in %s", observabilityPromFile)
	}
	return string(m[2]), nil
}

func validateHostPort(hostport string) error {
	i := strings.LastIndex(hostport, ":")
	if i <= 0 || i == len(hostport)-1 {
		return fmt.Errorf("target must be host:port, got %q", hostport)
	}
	if _, err := strconv.Atoi(hostport[i+1:]); err != nil {
		return fmt.Errorf("target port must be numeric, got %q", hostport[i+1:])
	}
	return nil
}
