package cfg

import (
	"fmt"
	"time"
)

const (
	ContainerInfoPath  = "/ws/services/container_infos.json"
	UpdateInfoFilePath = "/ws/services/version.json"
)

var (
	ContainerIDDict = make(map[string]map[string]interface{})
	UpdateStatus    = UpdateStatusStruct{
		Status:    false,
		Timestamp: float64(0),
	}
)

type UpdateStatusStruct struct {
	Status    bool
	Timestamp float64
}

type MqttStruct struct {
	Broker         string
	RestartTopic   string
	ServiceTopic   string
	PingCheckTopic string
	UpdateTopic    string
	ClientID       string
	UserName       string
	Password       string
}

func NewMqttStruct() *MqttStruct {
	return &MqttStruct{
		Broker:         "tcp://0.0.0.0:1883",
		RestartTopic:   "avi/local/restart_me",
		ServiceTopic:   "avi/local/service",
		PingCheckTopic: "avi/threat/health",
		UpdateTopic:    "avi/threat/update",
		ClientID:       fmt.Sprintf("restart_me_%d", time.Now().UnixNano()),
		UserName:       "",
		Password:       "",
	}
}

var MqttConfig = NewMqttStruct()
