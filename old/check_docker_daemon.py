"""
Checks the status of the Docker daemon and attempts to restart it if it's not running.

This module provides two functions:

check_docker_status():
    Checks if the Docker daemon is running by running the "docker info" command. Returns True if the daemon is running, False otherwise.

restart_docker():
    Attempts to restart the Docker daemon using the "sudo systemctl restart docker" command. Returns True if the restart was successful, False otherwise.

The main() function checks the Docker daemon status and attempts to restart it if it's not running. It waits up to 30 seconds for the daemon to become ready after the restart.
"""
import subprocess
import time
import platform

def check_docker_status():
    try:
        subprocess.run(["docker", "info"], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return True
    except subprocess.CalledProcessError:
        return False

def start_docker_mac():
    try:
        subprocess.run(["open", "-a", "Docker"], check=True)
        return True
    except subprocess.CalledProcessError:
        return False

def main():
    if platform.system() != "Darwin":
        print("This script is designed to run on macOS.")
        return

    if check_docker_status():
        print("Docker daemon is up and running.")
    else:
        print("Docker daemon is not running. Attempting to start Docker...")
        if start_docker_mac():
            print("Docker application started. Waiting for the daemon to become ready...")
            for _ in range(60):  # Wait up to 60 seconds
                if check_docker_status():
                    print("Docker daemon is now up and running.")
                    return
                time.sleep(1)
            print("Error: Docker daemon failed to start within the expected time.")
        else:
            print("Error: Failed to start Docker application.")

if __name__ == "__main__":
    main()