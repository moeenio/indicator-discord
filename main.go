package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

const DiscordToken = "NzkxNjk3MTY5MjMyODIyMjkz.X-S7cg.XH03rmhyR7ccJR17rd9eypZhQt4"

type DiscordGatewayPayload struct {
	Opcode    int         `json:"op"`
	EventData interface{} `json:"d"`
	EventName string      `json:"t,omitempty"`
}

type DiscordGatewayEventDataIdentify struct {
	Token      string                 `json:"token"`
	Intents    int                    `json:"intents"`
	Properties map[string]interface{} `json:"properties"`
}

func setupHeartbeat(connection *websocket.Conn, interval float64) {
	log.Println("setting up heartbeat")
	c := time.Tick(time.Duration(interval) * time.Millisecond)
	for range c {
		b, marshalErr := json.Marshal(DiscordGatewayPayload{1, nil, ""})
		if marshalErr != nil {
			log.Fatal("marshal: ", marshalErr)
			return
		}
		log.Println("sending heartbeat payload", string(b))

		err := connection.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			log.Fatal("heartbeat: ", err)
			return
		}
	}
}

func identify(connection *websocket.Conn) {
	b, marshalErr := json.Marshal(DiscordGatewayPayload{2,
		DiscordGatewayEventDataIdentify{DiscordToken, 0, map[string]interface{}{
			"$os":      "Linux",
			"$browser": "StatusSet",
			"$device":  "StatusSet",
		}}, ""})
	if marshalErr != nil {
		log.Fatal("marshal:", marshalErr)
		return
	}

	err := connection.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal("payload/send/identify:", err)
	}

	log.Println("Connected to Discord!")
}

func main() {
	go systray.Run(systrayReady, systrayExit)

	connection, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=6&encoding=json", nil)
	if err != nil {
		log.Fatal("dial: ", err)
		return
	}

	for {
		_, p, readErr := connection.ReadMessage()
		if readErr != nil {
			log.Fatal("read: ", readErr)
			return
		}

		var decodedMessage DiscordGatewayPayload
		decodeErr := json.Unmarshal(p, &decodedMessage)
		if decodeErr != nil {
			log.Fatal("payload/recieve: ", decodeErr)
			return
		}

		if decodedMessage.Opcode == 10 {
			data := decodedMessage.EventData.(map[string]interface{})

			heartbeatInterval := data["heartbeat_interval"].(float64)
			log.Println("recieved heartbeat interval", heartbeatInterval)
			go setupHeartbeat(connection, heartbeatInterval)
			identify(connection)
		}
	}
}

func systrayReady() {
	b, err := ioutil.ReadFile("discord.ico")
	if err != nil {
		log.Fatal("read icon:", err)
	}

	systray.SetIcon(b)
	systray.SetTitle("Discord Status")
	mOnline := systray.AddMenuItem("Online", "Online")

	for {
		select {
		case <-mOnline.ClickedCh:
			log.Fatalln("should set to Online")
		}
	}
}

func systrayExit() {

}
