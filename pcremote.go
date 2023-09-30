// Remote Control Package to execute computer functions via mqtt
package main

import (
	"fmt"
	"local/pcremote/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var canvasText = ""

func main() {
	var broker = "192.168.0.5"
	var port = 1883
	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	mqttOpts.SetDefaultPublishHandler(messagePubHandler)
	mqttOpts.OnConnect = connectHandler
	mqttOpts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(mqttOpts)
	client.Connect()
	a := app.New()
	w := a.NewWindow("MQTT Messages")
	w.Resize(fyne.NewSize(400, 400))
	icon, icoErr := fyne.LoadResourceFromPath("assets/remote4w.png")
	if icoErr != nil {
		fmt.Println(icoErr)
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

	text := widget.NewLabel(canvasText)
	textCont := container.NewMax(container.NewVScroll(text))
	w.SetContent(textCont)

	go func() {
		for range time.Tick(time.Second) {
			text.SetText(canvasText)
		}
	}()

	w.SetCloseIntercept(func() {
		w.Hide()
	})
	a.Run()
}

// MQTT Callback Functions
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	newText := fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	canvasText += newText
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
	topic1 := "computer/sound/device/speaker"
	client.Subscribe(topic1, 1, nil)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

//Message Consumer

func MessageRouter(msg mqtt.Message) {
	switch msg.Topic() {
	case "computer/sound/device/speaker":
		utils.AudioMessageRouter(msg)
	default:
		fmt.Println("Could not route mqtt message")
	}
}
