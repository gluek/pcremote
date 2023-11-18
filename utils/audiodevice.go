package utils

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// https://learn.microsoft.com/en-us/windows/win32/coreaudio/

var err error

func AudioMessageRouter(msg mqtt.Message) error {
	switch msg.Topic() {
	case "computer/sound/device/speaker":
		speakerID, err := getDeviceIDByName("Lautsprecher")
		if err != nil {
			return err
		}
		setDefaultEndpointByID(speakerID)
	case "computer/sound/device/soundbar":
		soundbarID, err := getDeviceIDByName("Cinebar")
		if err != nil {
			return err
		}
		setDefaultEndpointByID(soundbarID)
	}
	return nil
}
