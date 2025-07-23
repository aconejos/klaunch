package config

import (
	"os"
	"path/filepath"
	"strings"
)

// GetConfigFiles returns a list of JSON configuration files from the case_configs directory
func GetConfigFiles() ([]string, error) {
	configDir := "./case_configs"
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var configFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			configFiles = append(configFiles, file.Name())
		}
	}

	return configFiles, nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	return "./case_configs/default_topic.json"
}

// BuildConfigPath constructs a full path to a config file
func BuildConfigPath(filename string) string {
	return filepath.Join("./case_configs", filename)
}
