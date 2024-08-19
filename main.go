package main

import (
	"fmt"
	"restart_me/cfg"
	"restart_me/src"
)

func main() {

	// fmt.Println("Checking the container health")
	fmt.Println("Connected to broker: ", cfg.MqttConfig.Broker)

	// // Start the health check routine
	go src.CheckContainerHealth(cfg.ContainerIDDict, cfg.ContainerInfoPath, cfg.UpdateInfoFilePath)
	src.StartMqtt(cfg.MqttConfig)
}
