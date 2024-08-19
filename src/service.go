package src

import (
	"fmt"
	"os/exec"
)

func stopService(serviceName string) {
	/*
		Description:
			These functions manage the lifecycle of system services using the systemctl command, allowing for the stopping and starting of specified services.
		Parameters:
			serviceName string: The name of the service to stop.
	*/

	stopCommand := fmt.Sprintf("sudo systemctl stop %s", serviceName)
	stopCmd := exec.Command("/bin/sh", "-c", stopCommand)
	if err := stopCmd.Run(); err != nil {
		fmt.Printf("Error: Failed to stop service %s\n", serviceName)
		return
	}

	fmt.Printf("Service %s stop successfully.\n", serviceName)
}

func startService(serviceName string) {
	/*
		Description:
			Starts a system service using the systemctl command on a Linux-based system.
		Parameters
			serviceName string: The name of the service to be started.
	*/

	startCommand := fmt.Sprintf("sudo systemctl start %s", serviceName)
	startCmd := exec.Command("/bin/sh", "-c", startCommand)
	if err := startCmd.Run(); err != nil {
		fmt.Printf("Error: Failed to start service %s\n", serviceName)
		return
	}

	fmt.Printf("Service %s start successfully.\n", serviceName)
}
