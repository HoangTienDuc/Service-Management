package src

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"restart_me/cfg"
	"time"
)

func restartContainer(containerID string) {
	/*
		Restarts a Docker container using the provided containerID.
		Parameters:
			containerID: The ID of the Docker container to restart.
	*/
	command := fmt.Sprintf("docker restart %s", containerID)
	_, err := exec.Command("/bin/sh", "-c", command).Output()
	if err != nil {
		fmt.Printf("Error: Failed to restart container %s\n", containerID)
	}
}

func stopContainer(containerID string) {
	/*
		Stops a Docker container using the provided containerID.
		Parameters:
		 containerID: The ID of the Docker container to stop.
	*/
	command := fmt.Sprintf("docker stop %s", containerID)
	_, err := exec.Command("/bin/sh", "-c", command).Output()
	if err != nil {
		fmt.Printf("Error: Failed to stop container %s\n", containerID)
	}
}

func startContainer(containerID string) {
	/*
		Starts a Docker container using the provided containerID.
		Parameters:
		 containerID: The ID of the Docker container to start.
	*/
	command := fmt.Sprintf("docker start %s", containerID)
	_, err := exec.Command("/bin/sh", "-c", command).Output()
	if err != nil {
		fmt.Printf("Error: Failed to start container %s\n", containerID)
	}
}

func handleRestartContainer(containerID string) {
	/*
		Handles the logic to determine if a container should be restarted and restarts it if necessary.
		Parameters:
			containerID: The ID of the Docker container to potentially restart.
	*/

	// Check if the container needs to be restarted based on the condition
	if !cfg.ContainerIDDict[containerID]["is_restart"].(bool) || float64(time.Now().Unix())-cfg.ContainerIDDict[containerID]["timestamp"].(float64) >= 6000 {
		// Update the container status and timestamp
		cfg.ContainerIDDict[containerID]["is_restart"] = true
		cfg.ContainerIDDict[containerID]["timestamp"] = float64(time.Now().Unix())

		// Restart the container
		restartContainer(containerID)
		fmt.Println("Restarted the container:", containerID)
	}
}

func clearContainerHistory() {
	/*
		Clears the container history stored in cfg.ContainerIDDict.
	*/
	for k := range cfg.ContainerIDDict {
		delete(cfg.ContainerIDDict, k)
	}
}

func CheckContainerHealth(containerIDDict map[string]map[string]interface{}, containerInfoPath, updateInfoFilePath string) {
	/*
		The CheckContainerHealth function continuously monitors the health of containers based on the provided container ID information and performs specific actions if certain conditions are met. It checks the container status every 30 seconds and takes appropriate actions, such as restarting the container or restoring factory settings, depending on the conditions.
		Parameters:
			containerIDDict (map[string]map[string]interface{}): A dictionary where each key is a container ID (string), and the value is another dictionary containing various metadata about the container, such as the timestamp of the last ping.
			containerInfoPath (string): The file path to the container information file, which contains detailed information about each container.
			updateInfoFilePath (string): The file path to the update information file, used to check the timestamp of the last update.
	*/
	for {
		time.Sleep(30 * time.Second)
		for containerID, pingInfo := range containerIDDict {
			pingTime, ok := pingInfo["timestamp"].(float64)
			if !ok {
				fmt.Println("Invalid ping time format")
				continue
			}

			currentTime := float64(time.Now().Unix())
			if currentTime-cfg.UpdateStatus.Timestamp < 3600 && cfg.UpdateStatus.Status {
				fmt.Println("Time to pending update")
				continue // Skip to the next iteration of the loop.
			}

			if currentTime-pingTime >= 120 {
				if _, err := os.Stat(containerInfoPath); os.IsNotExist(err) {
					fmt.Println("Container info file does not exist")
					continue
				}

				containerInfoFile, err := os.Open(containerInfoPath)
				if err != nil {
					fmt.Println("Error opening container info file:", err)
					continue
				}

				var containerInfo map[string]map[string]interface{}
				if err := json.NewDecoder(containerInfoFile).Decode(&containerInfo); err != nil {
					fmt.Println("Error decoding container info:", err)
					continue
				}

				data, ok := containerInfo[containerID]
				if !ok {
					handleRestartContainer(containerID)
					fmt.Println("Restarting the container ", containerID)
					continue
				}

				restartCount, ok := data["restart_count"].(float64)
				if !ok {
					fmt.Println("Invalid restart count format")
					continue
				}

				updateTime := float64(getUpdateTime(updateInfoFilePath))

				if currentTime-updateTime > 3600 && restartCount >= 5 {
					restoreFactorySettings()
				} else {
					handleRestartContainer(containerID)
				}

			}
		}
	}
}
