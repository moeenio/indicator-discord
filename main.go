package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

const DiscordToken = "no u"

var connection *websocket.Conn

type ApplicationSettings struct {
	DiscordToken string `json:"token"`
	LastStatus   string `json:"lastStatus"`
}

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

type DiscordGatewayEventDataUpdateStatus struct {
	TimeSinceIdle *int                   `json:"since"`
	Activities    map[string]interface{} `json:"activities, omitempty"`
	Status        string                 `json:"status"`
	IsAfk         bool                   `json:"afk"`
}

func recieveIncomingPayloads() {
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
			go setupHeartbeat(heartbeatInterval)
			identify()
		}
	}
}

func setupHeartbeat(interval float64) {
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

func identify() {
	b, marshalErr := json.Marshal(DiscordGatewayPayload{2,
		DiscordGatewayEventDataIdentify{DiscordToken, 0, map[string]interface{}{
			"$os":      "Linux",
			"$browser": "Chrome",
			"$device":  "",
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

func setStatus(status string) {
	var afk bool

	if status == "invisible" || status == "offline" {
		afk = true
	} else {
		afk = false
	}

	b, marshalErr := json.Marshal(DiscordGatewayPayload{3, DiscordGatewayEventDataUpdateStatus{
		nil,
		map[string]interface{}{},
		status,
		afk,
	}, ""})
	if marshalErr != nil {
		log.Fatal("marshal:", marshalErr)
	}
	log.Println(string(b))

	err := connection.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	/*fyneApp := app.New()

	configPath := configdir.LocalConfig("dcstatus")
	makePathErr := configdir.MakePath(configPath)
	if makePathErr != nil {
		log.Fatal(makePathErr)
	}

	configFile := filepath.Join(configPath, "settings.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		promptForToken(fyneApp)
	}*/

	go systray.Run(systrayReady, systrayExit)

	var err error
	connection, _, err = websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=6&encoding=json", nil)
	if err != nil {
		log.Fatal("dial: ", err)
		return
	}

	recieveIncomingPayloads()
}

func systrayReady() {
	b, err := ioutil.ReadFile("discord.ico")
	if err != nil {
		log.Fatal("read icon:", err)
	}

	systray.SetIcon(b)
	mOnline := systray.AddMenuItem("Online", "Online")
	mIdle := systray.AddMenuItem("Idle", "Idle")
	mDnd := systray.AddMenuItem("Do Not Disturb", "Do Not Disturb")
	mInvisible := systray.AddMenuItem("Invisible", "Invisible")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Exit")

	for {
		select {
		case <-mOnline.ClickedCh:
			setStatus("online")
		case <-mIdle.ClickedCh:
			setStatus("idle")
		case <-mDnd.ClickedCh:
			setStatus("dnd")
		case <-mInvisible.ClickedCh:
			setStatus("invisible")
		case <-mExit.ClickedCh:
			setStatus("offline")
			connection.Close()
			systray.Quit()
			os.Exit(0)
		}

	}
}

func systrayExit() {

}

/*func promptForToken(fyneApp fyne.App) {
	window := fyneApp.NewWindow("Input your Discord token...")

	tokenInput := widget.NewPasswordEntry()
	window.SetContent(widget.NewLabel("Input your Discord token..."))
	window.ShowAndRun()
}*/
