// Remote control package to execute pc routines via mqtt
// Author: Gerrit Lükens
// Date: 2023-11-18
package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"

	audio "github.com/gluek/pcremote/internal/audio"
)

var MQTTMsgLog = ""

var err error

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./assets")
	if err = viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	var broker = viper.Get("broker.ip")
	var port = viper.Get("broker.port")

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	mqttOpts.SetDefaultPublishHandler(messagePubHandler)
	mqttOpts.OnConnect = connectHandler
	mqttOpts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(mqttOpts)
	client.Connect()
	a := app.New()
	w := a.NewWindow("MQTT Messages")
	w.Resize(fyne.NewSize(400, 400))
	icon, err := fyne.LoadResourceFromPath("./assets/trayicon.png")
	if err != nil {
		panic(err)
	}

	w.SetIcon(icon)

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(icon)
	}

	text := widget.NewLabel(MQTTMsgLog)
	textCont := container.NewBorder(nil, nil, nil, nil, container.NewVScroll(text))
	w.SetContent(textCont)

	go func() {
		for range time.Tick(time.Second) {
			text.SetText(MQTTMsgLog)
		}
	}()

	w.SetCloseIntercept(func() {
		w.Hide()
	})
	a.Run()
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	newText := fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	MQTTMsgLog += newText
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Printf("Connected to %s:%s\n", viper.Get("broker.ip"), viper.Get("broker.port"))
	log.Printf("Subscribed to:\n")
	if err = audio.RegisterAudioDevices(client, &MQTTMsgLog); err != nil {
		log.Fatalf("could not register audio devices: %s\n", err)
	}

}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}
