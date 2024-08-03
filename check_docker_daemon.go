package main

import (
	"fmt"
	"os/exec"
	"time"
)

func check_socker_status() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return true
	}
	return false
}

func start_docker_mac() bool {
	cmd := exec.Command("open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		return true
	}
	return false
}

func check_docker_daemon() error {

	if check_socker_status() {
		fmt.Println("Docker daemon is up and running.")
	} else {
		fmt.Println("Docker daemon is not running. Attempting to start Docker...")
		if start_docker_mac() {
			fmt.Println("Docker application started. Waiting for the daemon to become ready...")
			for i := 0; i < 60; i++ {
				if check_socker_status() {
					fmt.Println("Docker daemon is now up and running.")
					return nil
				}
				time.Sleep(1 * time.Second)
			}
			Error := "Error: Docker daemon failed to start within the expected time."
			return fmt.Errorf(Error)
		} else {
			Error := "Error: Failed to start Docker application."
			return fmt.Errorf(Error)
		}
	}
	return nil
}
