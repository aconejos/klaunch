package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

func check_connector_updates(inputVersion string) error {
	pwd, _ := os.Getwd()

	// Define variables
	url := "https://repo1.maven.org/maven2/org/mongodb/kafka/mongo-kafka-connect/"
	downloadDir := filepath.Join(pwd, "volumes")
	envFile := filepath.Join(pwd, ".env")

	// Get HTML content
	resp, _ := http.Get(url)
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// Extract latest version
	re := regexp.MustCompile(`<a href="(\d*(\.\d+)*)/?"`)
	matches := re.FindAllStringSubmatch(html, -1)
	versions := make([]string, len(matches))
	// semantic version strings must begin with a leading "v", as in "v1.0.0"
	for i, match := range matches {
		versions[i] = fmt.Sprintf("v%s", strings.TrimSuffix(match[1], "/"))
	}
	// sort versions semmatically
	semver.Sort(versions)
	latestVersion := versions[len(versions)-1]

	// clean up v
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	if len(inputVersion) == 0 {
		fmt.Printf("MongoDB Kafka connector latest Version: %s\n", latestVersion)
	} else {
		fmt.Printf("Using MongoDB Kafka connector Version: %s\n", inputVersion)
		latestVersion = inputVersion

	}

	// Construct download link
	downloadLink := fmt.Sprintf("%s%s/mongo-kafka-connect-%s-all.jar", url, latestVersion, latestVersion)

	// Check if file exists and download if not
	filePath := filepath.Join(downloadDir, fmt.Sprintf("mongo-kafka-connect-%s-all.jar", latestVersion))
	err := download_file(downloadLink, filePath)
	if err != nil {
		fmt.Println("Check the list of existing versions: ", url)
		return err
	}

	// Read current version from .env file
	envContent, _ := os.ReadFile(envFile)
	re = regexp.MustCompile(`MONGO_KAFKA_CONNECT_VERSION=(.+)`)
	match := re.FindStringSubmatch(string(envContent))
	
	// show match value 
	//fmt.Printf("MONGO_KAFKA_CONNECT_VERSION=%s\n", match[1])
	//fmt.Printf("I will updated to this =%s\n", latestVersion)

	if len(match) > 1 && match[1] != latestVersion {
		currentVersion := match[1]
		fmt.Printf("Updating MONGO_KAFKA_CONNECT_VERSION from %s to %s\n", currentVersion, latestVersion)
		updatedContent := strings.Replace(string(envContent), currentVersion, latestVersion, 1)
		os.WriteFile(envFile, []byte(updatedContent), 0644)
		fmt.Println("Choosen version of mongo-kafka-connect has been downloaded and updated in the .env file.")
	}

	return nil
}

func download_file(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
