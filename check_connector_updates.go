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

func check_connector_updates() error {
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
	for i, match := range matches {
		versions[i] = strings.TrimSuffix(match[1], "/")
	}
	semver.Sort(versions)
	// print all sorted versions	
	fmt.Println("All sorted versions:")	
	for _, version := range versions {
		fmt.Println(version)
	}
	latestVersion := versions[len(versions)-1]

	fmt.Printf("MongoDB Kafka connector latest Version: %s\n", latestVersion)

	// Construct download link
	downloadLink := fmt.Sprintf("%s%s/mongo-kafka-connect-%s-all.jar", url, latestVersion, latestVersion)

	// Check if file exists and download if not
	filePath := filepath.Join(downloadDir, fmt.Sprintf("mongo-kafka-connect-%s-all.jar", latestVersion))
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Println("File already exists.")
	} else {
		err = download_file(downloadLink, filePath)
		if err != nil {
			fmt.Println("Failed to download the JAR file:", err)
			return err
		}
	}

	// Read current version from .env file
	envContent, _ := os.ReadFile(envFile)
	re = regexp.MustCompile(`MONGO_KAFKA_CONNECT_VERSION=(.+)`)
	match := re.FindStringSubmatch(string(envContent))
	if len(match) > 1 && match[1] != latestVersion {
		currentVersion := match[1]
		fmt.Printf("Updating MONGO_KAFKA_CONNECT_VERSION from %s to %s\n", currentVersion, latestVersion)
		updatedContent := strings.Replace(string(envContent), currentVersion, latestVersion, 1)
		os.WriteFile(envFile, []byte(updatedContent), 0644)
		fmt.Println("Latest version of mongo-kafka-connect has been downloaded and updated in the .env file.")
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
