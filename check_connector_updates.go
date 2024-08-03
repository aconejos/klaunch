package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
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
	re := regexp.MustCompile(`<a href="([1-9]\d*(\.\d+)*)/?"`)
	matches := re.FindAllStringSubmatch(html, -1)
	versions := make([]string, len(matches))
	for i, match := range matches {
		versions[i] = strings.TrimSuffix(match[1], "/")
	}
	sort.Strings(versions)
	latestVersion := versions[len(versions)-1]

	fmt.Printf("MongoDB Kafka connector latest Version: %s\n", latestVersion)

	// Construct download link
	downloadLink := fmt.Sprintf("%s%s/mongo-kafka-connect-%s-all.jar", url, latestVersion, latestVersion)

	// Check if file exists and download if not
	filePath := filepath.Join(downloadDir, fmt.Sprintf("mongo-kafka-connect-%s-all.jar", latestVersion))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = downloadFile(downloadLink, filePath)
		if err != nil {
			fmt.Println("Failed to download the JAR file:", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("File already exists.")
	}

	// Read current version from .env file
	envContent, _ := os.ReadFile(envFile)
	re = regexp.MustCompile(`MONGO_KAFKA_CONNECT_VERSION=(.+)`)
	match := re.FindStringSubmatch(string(envContent))
	if len(match) > 1 {
		currentVersion := match[1]
		fmt.Printf("Updating MONGO_KAFKA_CONNECT_VERSION from %s to %s\n", currentVersion, latestVersion)
		updatedContent := strings.Replace(string(envContent), currentVersion, latestVersion, 1)
		os.WriteFile(envFile, []byte(updatedContent), 0644)
	}

	fmt.Println("Latest version of mongo-kafka-connect has been downloaded and updated in the .env file.")
}

func downloadFile(url string, filepath string) error {
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
