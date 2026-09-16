package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const samplePromYAML = `global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'kafka'
    static_configs:
      - targets:
          - kafka1:8091
          - kafka2:8092
          - kafka3:8093

  - job_name: 'zookeeper'
    static_configs:
      - targets:
          - zookeeper1:8091

  - job_name: 'kafka-connect'
    static_configs:
      - targets:
          - kafka-connect:8091
`

func withPromFixture(t *testing.T, body string) func() {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "volumes"), 0755); err != nil {
		t.Fatalf("mkdir volumes: %v", err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.WriteFile("volumes/prometheus.yml", []byte(body), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return func() { _ = os.Chdir(orig) }
}

func TestRewriteConnectTarget(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   string
		wantErr  bool
		wantHave string
		wantGone string
	}{
		{
			name:     "swap default to remote host",
			input:    samplePromYAML,
			target:   "host.docker.internal:8095",
			wantHave: "- host.docker.internal:8095",
			wantGone: "- kafka-connect:8091",
		},
		{
			name:     "reset to default",
			input:    strings.Replace(samplePromYAML, "kafka-connect:8091", "host.docker.internal:8095", 1),
			target:   "kafka-connect:8091",
			wantHave: "- kafka-connect:8091",
			wantGone: "- host.docker.internal:8095",
		},
		{
			name:    "missing kafka-connect job",
			input:   "scrape_configs:\n  - job_name: 'kafka'\n    static_configs:\n      - targets:\n          - kafka1:8091\n",
			target:  "foo:1",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := withPromFixture(t, tc.input)
			defer cleanup()

			err := rewriteConnectTarget(tc.target)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("rewriteConnectTarget: %v", err)
			}
			got, _ := os.ReadFile("volumes/prometheus.yml")
			gs := string(got)
			if !strings.Contains(gs, tc.wantHave) {
				t.Errorf("expected %q in output:\n%s", tc.wantHave, gs)
			}
			if tc.wantGone != "" && strings.Contains(gs, tc.wantGone) {
				t.Errorf("expected %q to be removed:\n%s", tc.wantGone, gs)
			}
			// other jobs must survive
			for _, must := range []string{"job_name: 'kafka'", "job_name: 'zookeeper'", "- kafka1:8091", "- zookeeper1:8091"} {
				if !strings.Contains(gs, must) {
					t.Errorf("expected %q to survive rewrite", must)
				}
			}
		})
	}
}

func TestReadCurrentConnectTarget(t *testing.T) {
	cleanup := withPromFixture(t, samplePromYAML)
	defer cleanup()
	got, err := readCurrentConnectTarget()
	if err != nil {
		t.Fatalf("readCurrentConnectTarget: %v", err)
	}
	if got != "kafka-connect:8091" {
		t.Errorf("got %q, want kafka-connect:8091", got)
	}
}

func TestValidateHostPort(t *testing.T) {
	ok := []string{"kafka-connect:8091", "host.docker.internal:8095", "10.0.0.1:9100"}
	bad := []string{"", "no-port", "trailing:", ":8091", "host:notaport"}
	for _, s := range ok {
		if err := validateHostPort(s); err != nil {
			t.Errorf("expected %q to validate, got %v", s, err)
		}
	}
	for _, s := range bad {
		if err := validateHostPort(s); err == nil {
			t.Errorf("expected %q to fail validation", s)
		}
	}
}
