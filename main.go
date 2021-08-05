// Refrer to Discord's documentation for gateway, payloads, opcodes etc : https://discord.com/developers/docs/topics/gateway

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
	"github.com/kirsle/configdir"
)

// Global variables : websocket connection to Discord, and user token for Discord.
var connection *websocket.Conn
var discordToken string

// Types used for generating JSON payloads sent to the Discord gateway.

// Generic payload type. Any specific information is stored in EventData.
type DiscordGatewayPayload struct {
	Opcode    int         `json:"op"`
	EventData interface{} `json:"d"`
	EventName string      `json:"t,omitempty"`
}

// EventData object for the "identify" payload.
type DiscordGatewayEventDataIdentify struct {
	Token      string                 `json:"token"`
	Intents    int                    `json:"intents"`
	Properties map[string]interface{} `json:"properties"`
}

// EventData obejct for the "update status" payload.
type DiscordGatewayEventDataUpdateStatus struct {
	TimeSinceIdle *int                   `json:"since"`
	Activities    map[string]interface{} `json:"activities, omitempty"`
	Status        string                 `json:"status"`
	IsAfk         bool                   `json:"afk"`
}

// Recieve incoming payloads from Discord and handle them.
func recieveIncomingPayloads() {
	for {
		_, p, readErr := connection.ReadMessage()
		if readErr != nil {
			log.Fatal("error, cannot read from connection: ", readErr)
			return
		}

		// Decode the recieved payload
		var decodedMessage DiscordGatewayPayload
		decodeErr := json.Unmarshal(p, &decodedMessage)
		if decodeErr != nil {
			log.Fatal("failed decoding recieved payload: ", decodeErr)
			return
		}
		log.Println("recieved payload: ", decodedMessage)

		// Handle payload opcode 10 : recieved heartbeat interval
		if decodedMessage.Opcode == 10 {
			data := decodedMessage.EventData.(map[string]interface{})
			heartbeatInterval := data["heartbeat_interval"].(float64)

			go setupHeartbeat(heartbeatInterval)
			identify()
		}
	}
}

// Regularly send a "heartbeat" payload to Discord, at the given interval.
func setupHeartbeat(interval float64) {

	c := time.Tick(time.Duration(interval) * time.Millisecond)
	for range c {
		// Create the heartbeat payload
		b, marshalErr := json.Marshal(DiscordGatewayPayload{1, nil, ""})
		if marshalErr != nil {
			log.Fatal("marshal: ", marshalErr)
			return
		}

		// Send that payload via the connection
		log.Println("sending payload (heartbeat): ", string(b))
		err := connection.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			log.Fatal("error sending payload (heartbeat):", err)
			return
		}
	}
}

// Identify against the Discord gateway, with a user token and client properties.
func identify() {
	// Create the identify payload
	b, marshalErr := json.Marshal(DiscordGatewayPayload{2,
		DiscordGatewayEventDataIdentify{discordToken, 0, map[string]interface{}{
			"$os":      "Linux",
			"$browser": "Chrome",
			"$device":  "",
		}}, ""})
	if marshalErr != nil {
		log.Fatal("marshal error:", marshalErr)
		return
	}

	// Send the payload via the connection
	err := connection.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal("error sending payload:", err)
	}

	log.Println("Connected to Discord!")
}

// Sends a payload to discord to set the given status and the afk value
func setStatus(status string) {
	var afk bool
	if status == "invisible" || status == "offline" {
		afk = true
	} else {
		afk = false
	}

	// Create a payload to set the status
	b, marshalErr := json.Marshal(DiscordGatewayPayload{3,
		DiscordGatewayEventDataUpdateStatus{
			nil,
			map[string]interface{}{},
			status,
			afk,
		}, ""})
	if marshalErr != nil {
		log.Fatal("marshal error:", marshalErr)
	}

	// Send that payload via the connection
	log.Println("sending payload:", string(b))
	err := connection.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal("error sending payload:", err)
	}
}

func main() {
	// create a folder for configuration, inside of the proper user directory
	configPath := configdir.LocalConfig("indicator-discord")
	makePathErr := configdir.MakePath(configPath)
	if makePathErr != nil {
		log.Fatal(makePathErr)
	}

	// the path to the file that stores the user token, inside the configPath
	tokenFilePath := filepath.Join(configPath, "token.txt")

	// if the token file doesn't exist...
	if _, statErr := os.Stat(tokenFilePath); os.IsNotExist(statErr) {
		// ...create it
		tokenFile, createErr := os.Create(tokenFilePath)
		if createErr != nil {
			log.Fatal("create token file: ", createErr)
			return
		}

		// and make it private (rw-rw----)
		chmodErr := tokenFile.Chmod(0660)
		if chmodErr != nil {
			log.Fatal("chmod token file: ", chmodErr)
			return
		}
	}

	// stat the token file, to get its size later on
	tokenFileInfo, statErr := os.Stat(tokenFilePath)
	if statErr != nil {
		log.Fatal("stat token file: ", statErr)
		return
	}

	// get the token file's size
	tokenFileSize := tokenFileInfo.Size()
	// if the file is empty/too small, prompt to add the token and exit
	if tokenFileSize < 1 {
		log.Println("Please input your Discord token in", tokenFilePath)
		return
	}

	// read the token file
	tokenFileData, readFileErr := ioutil.ReadFile(tokenFilePath)
	if readFileErr != nil {
		log.Fatal("read token file: ", readFileErr)
		return
	}

	// remove any trailing new line read from the token file
	tokenFileString := string(tokenFileData)
	tokenFileString = strings.TrimSuffix(tokenFileString, "\n")
	tokenFileString = strings.TrimSuffix(tokenFileString, "\r\n")

	// finally, set the global token variable
	discordToken = tokenFileString

	// connect to the discord gateway
	var dialErr error
	connection, _, dialErr = websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=6&encoding=json", nil)
	if dialErr != nil {
		log.Fatal("dial: ", dialErr)
		return
	}

	// show the systray icon
	go systray.Run(systrayReady, systrayExit)

	// recieve and handle payloads from the discord gateway
	recieveIncomingPayloads()
}

func systrayReady() {
	// Read icon for the systray
	b, err := ioutil.ReadFile("discord.ico")
	if err != nil {
		log.Fatal("read icon:", err)
	}
	systray.SetIcon(b)

	// Add menu items to the systray
	mOnline := systray.AddMenuItem("Online", "Online")
	mIdle := systray.AddMenuItem("Idle", "Idle")
	mDnd := systray.AddMenuItem("Do Not Disturb", "Do Not Disturb")
	mInvisible := systray.AddMenuItem("Invisible", "Invisible")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Exit")

	// Call the setStatus function with the appropriate status text
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
		// or quit the systray, which calls systrayExit
		case <-mExit.ClickedCh:
			systray.Quit()
			os.Exit(0)
		}

	}
}

// When exiting the program, go offline and close the connection
func systrayExit() {
	setStatus("offline")
	connection.Close()
}
