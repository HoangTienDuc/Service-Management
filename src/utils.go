package src

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func restoreFactorySettings() {
	/*
		Executes a script to restore the system to its factory settings.
	*/

	cmd := exec.Command("/bin/sh", "factory_reset.sh")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing script:", err)
		return
	}

	fmt.Println("Finished the restoreFactorySettings")
}

func getUpdateTime(updateInfoFilePath string) int64 {
	/*
		Retrieves the release date of the latest update from a JSON file.
	*/

	updateFile, err := os.Open(updateInfoFilePath)
	if err != nil {
		fmt.Println("Error opening update info file:", err)
		return 0
	}
	defer updateFile.Close()

	var updateInfo map[string]interface{}
	if err := json.NewDecoder(updateFile).Decode(&updateInfo); err != nil {
		fmt.Println("Error decoding update info:", err)
		return 0
	}

	updateDate, ok := updateInfo["release_date"].(float64)
	if !ok {
		fmt.Println("Invalid update date format")
		return 0
	}

	return int64(updateDate)
}
