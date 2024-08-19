package src

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"restart_me/cfg"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func onConnectHandler(cfg *cfg.MqttStruct) func(client MQTT.Client) {
	/*
		Description:
			This function returns a handler that is triggered when the MQTT client successfully connects to the broker. The handler subscribes the client to specific topics provided in the cfg parameter. If any subscription fails, an error message is logged.
		Parameters:
			cfg *cfg.MqttStruct: A pointer to an MqttStruct containing the configuration for MQTT, including the topics to subscribe to.
		Returns:
			A function that takes an MQTT.Client as a parameter and subscribes to the topics defined in cfg.
	*/
	return func(client MQTT.Client) {
		fmt.Println("Connected to broker")
		if token := client.Subscribe(cfg.RestartTopic, 1, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error subscribing to topic:", token.Error())
		}

		if token := client.Subscribe(cfg.ServiceTopic, 1, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error subscribing to topic:", token.Error())
		}

		if token := client.Subscribe(cfg.PingCheckTopic, 1, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error subscribing to ping_check_topic:", token.Error())
		}
	}
}

func sendPing(client MQTT.Client) {
	/*
		Description:
			Sends a ping message to the "avi/threat/health" topic. The message includes a command, timestamp, and container ID. The message is marshaled into JSON format before being sent. If marshalling fails, an error message is logged.
		Parameters:
			client MQTT.Client: The MQTT client used to publish the ping message.
		Returns:
			None
	*/
	managementPingMsg := map[string]interface{}{
		"command":      "ping",
		"timestamp":    time.Now().Unix(),
		"container_id": "management_service",
	}
	jsonManagementPingMsg, err := json.Marshal(managementPingMsg)
	if err != nil {
		fmt.Println("Error: Could not marshal response JSON")
		return
	}
	client.Publish("avi/threat/health", 1, false, jsonManagementPingMsg)
}

func handlePing(client MQTT.Client, data map[string]interface{}) {
	/*
		Description:
			Processes incoming ping messages. It checks the container ID in the message. If the container ID is "management_service", the function exits. Otherwise, it sends a ping response and updates or creates the ping information for the specified container ID in the configuration map.
		Parameters:
			client MQTT.Client: The MQTT client used to send the ping response.
			data map[string]interface{}: A map containing the data from the incoming ping message.
		Returns:
			None
	*/
	containerID, ok := data["container_id"].(string)
	if !ok {
		fmt.Println("Error: Invalid container ID")
		return
	}
	if containerID == "management_service" {
		return
	}
	sendPing(client)
	value, exists := cfg.ContainerIDDict[containerID]
	if exists {
		updateContainerPingInfo(containerID, value)
	} else {
		createNewContainerPingInfo(containerID, value)
	}
}

func updateContainerPingInfo(containerID string, data map[string]interface{}) {
	/*
		Description:
			This function updates the ping information for a container identified by containerID in the cfg.ContainerIDDict map. It updates the timestamp to the current time and ensures that the is_restart flag is set to false. If a boot_time_timestamp is provided in data, it updates that as well.
		Parameters:
			containerID: The identifier for the container whose ping information is being updated.
			data: A map containing additional data for the container, including the boot_time_timestamp.
	*/
	cfg.ContainerIDDict[containerID]["timestamp"] = float64(time.Now().Unix())
	cfg.ContainerIDDict[containerID]["is_restart"] = false
	// Check if the key exists in the map
	bootTimeTimestamp, ok := data["boot_time_timestamp"].(float64)
	if ok {
		cfg.ContainerIDDict[containerID]["boot_time_timestamp"] = bootTimeTimestamp // Use bootTimeTimestamp variable here
	}
}

func createNewContainerPingInfo(containerID string, data map[string]interface{}) {
	/*
		Description:
			This function creates a new ping information entry for a container identified by containerID in the cfg.ContainerIDDict map. It initializes the timestamp with the current time and sets is_restart to false. If a boot_time_timestamp is provided in data, it includes it in the new entry.
		Parameters:
			containerID: The identifier for the container whose ping information is being created.
			data: A map containing additional data for the container, potentially including the boot_time_timestamp.
	*/

	fmt.Println("createNewContainerPingInfo: ", containerID)

	var pingInfo map[string]interface{}

	bootTimeTimestamp, ok := data["boot_time_timestamp"].(float64)
	if !ok {
		pingInfo = map[string]interface{}{
			"timestamp":  float64(time.Now().Unix()), // Include the comma here
			"is_restart": false,
		}
	} else {
		pingInfo = map[string]interface{}{
			"timestamp":           float64(time.Now().Unix()),
			"boot_time_timestamp": bootTimeTimestamp, // Use the retrieved value
			"is_restart":          false,
		}
	}
	cfg.ContainerIDDict[containerID] = pingInfo
	// fmt.Println("pingInfo: ", pingInfo)
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	/*
		Description:
			Handles incoming MQTT messages by parsing the payload as JSON, extracting a command, and executing the corresponding action.
		Parameters
			client MQTT.Client: The MQTT client instance that received the message.
			message MQTT.Message: The message received from the MQTT broker, containing the payload to be processed.
	*/
	fmt.Println("Received message")
	payload := message.Payload()
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		fmt.Println("Error: Invalid JSON payload")
		return
	}

	command, ok := data["command"].(string)
	if !ok {
		fmt.Println("Error: Command is not found")
		return
	}
	switch command {
	case "restart_container":
		containerID, ok := data["container_id"].(string)
		if !ok {
			fmt.Println("Error: Invalid container ID")
			return
		}
		restartContainer(containerID)
		fmt.Println("Restarted the container ", containerID)
	case "start_container":
		containerID, ok := data["container_id"].(string)
		if !ok {
			fmt.Println("Error: Invalid container ID")
			return
		}
		startContainer(containerID)
		fmt.Println("Started the container ", containerID)
	case "stop_container":
		containerID, ok := data["container_id"].(string)
		if !ok {
			fmt.Println("Error: Invalid container ID")
			return
		}
		stopContainer(containerID)
		fmt.Println("Stopped the container ", containerID)
	case "clear_container_history":
		clearContainerHistory()
		fmt.Println("Cleared all container history")
	case "ping":
		handlePing(client, data)
	case "update_pending":
		status, ok := data["status"].(bool)
		if !ok {
			fmt.Println("Error: Invalid container ID")
			return
		}
		cfg.UpdateStatus.Status = status
		if status {
			cfg.UpdateStatus.Timestamp = float64(time.Now().Unix())
		} else {
			cfg.UpdateStatus.Timestamp = float64(0)
		}

		fmt.Println("Update pending dict: ", cfg.UpdateStatus)

	case "stop_service":
		serviceName, ok := data["service_name"].(string)
		if !ok {
			fmt.Println("Error: Invalid service_name")
			return
		}
		stopService(serviceName)
	case "start_service":
		serviceName, ok := data["service_name"].(string)
		if !ok {
			fmt.Println("Error: Invalid service_name")
			return
		}
		startService(serviceName)

	default:
		fmt.Println("Error: Unknown command")
	}
}

func StartMqtt(mqttConfig *cfg.MqttStruct) {
	/*
		Description:
			Initializes and starts an MQTT client using the provided configuration, subscribes to relevant topics, and manages the lifecycle of the client.
		Parameters
			mqttConfig *cfg.MqttStruct: A pointer to a configuration structure containing MQTT settings such as broker URL, client ID, credentials, and subscription topics.
	*/

	// Initialize MQTT client options
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttConfig.Broker)
	opts.SetClientID(mqttConfig.ClientID)
	opts.SetUsername(mqttConfig.UserName)
	opts.SetPassword(mqttConfig.Password)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(onConnectHandler(mqttConfig))

	// Create and start MQTT client
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to MQTT topics
	if token := client.Subscribe(mqttConfig.RestartTopic, 1, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := client.Subscribe(mqttConfig.ServiceTopic, 1, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := client.Subscribe(mqttConfig.PingCheckTopic, 1, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := client.Subscribe(mqttConfig.UpdateTopic, 1, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Printf("Listening for messages on topics [%s, %s, %s, %s]...\n", mqttConfig.RestartTopic, mqttConfig.ServiceTopic, mqttConfig.PingCheckTopic, mqttConfig.UpdateTopic)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Disconnect MQTT client before exiting
	client.Disconnect(250)
}
