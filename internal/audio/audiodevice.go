package audio

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

// https://learn.microsoft.com/en-us/windows/win32/coreaudio/

var err error

type audioDevice struct {
	name          string
	topic         string
	id            string
	friendly_name string
}

var audioDevices []audioDevice
var mqttMsgLogPointer *string

var messageAudioHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	for _, device := range audioDevices {
		if msg.Topic() == device.topic {
			if err = device.setDefaultEndpointByID(); err != nil {
				log.Printf("could not set audio endpoint: %s", err)
			}
		}
	}
	newText := fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	*mqttMsgLogPointer += newText
}

func RegisterAudioDevices(client mqtt.Client, canvas *string) error {
	audioDevices = []audioDevice{}
	audioDevicesFromConfig := viper.GetStringMap("audiodevices")
	mqttMsgLogPointer = canvas

	for _, device := range audioDevicesFromConfig {
		deviceName := device.(map[string]any)["friendly_name"].(string)
		deviceTopic := device.(map[string]any)["topic"].(string)
		deviceID, err := getDeviceIDByName(deviceName)
		if err != nil {
			return err
		}
		log.Printf("  AudioDevice - Name: %s, Topic: %s\n", deviceName, deviceTopic)
		client.Subscribe(deviceTopic, 1, messageAudioHandler)
		audioDevices = append(audioDevices,
			audioDevice{name: deviceName, topic: deviceTopic, id: deviceID, friendly_name: ""})
	}
	return nil
}
