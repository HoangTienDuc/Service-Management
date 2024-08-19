# Table of contents
- Introduction
- Prerequires
- Installation 
- How to use 
- Todo


## INTRODUCTION
This is a service that manages other services running using the MQTT platform.
Services include:
- Docker Container
- Systemd
- Update Service
- Factory Reset Service


## PREREQUIRES
- Ubuntu 18.04
- go version go1.22.3 linux/amd64

## INSTALLATION
```
go install github.com/eclipse/paho.mqtt.golang
```

## HOW TO USE
```
go build main.go

./main.go

```