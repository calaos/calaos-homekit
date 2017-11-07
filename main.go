package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
)

type WebSocketConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

type Configuration struct {
	WebSocketServer WebSocketConfig
	PinCode         string
}

type CalaosJsonMsg struct {
	Msg   string `json:"msg"`
	MsgID string `json:"msg_id"`
}

type CalaosJsonMsgLogin struct {
	Msg  string `json:"msg"`
	Data struct {
		Success string `json:"success"`
	} `json:"data"`
	MsgID string `json:"msg_id"`
}

type CalaosJsonMsgEvent struct {
	Msg  string `json:"msg"`
	Data struct {
		EventRaw string `json:"event_raw"`
		Data     struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"data"`
		TypeStr string `json:"type_str"`
		Type    string `json:"type"`
	} `json:"data"`
}

type CalaosIO struct {
	Visible string `json:"visible"`
	VarType string `json:"var_type"`
	ID      string `json:"id"`
	IoType  string `json:"io_type"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	GuiType string `json:"gui_type"`
	State   string `json:"state"`
	Rw      string `json:"rw,omitempty"`
}
type CalaosHome struct {
	Type string     `json:"type"`
	Hits string     `json:"hits"`
	Name string     `json:"name"`
	IOs  []CalaosIO `json:"items"`
}

type CalaosJsonMsgHome struct {
	Msg  string `json:"msg"`
	Data struct {
		Home    []CalaosHome  `json:"home"`
		Cameras []interface{} `json:"cameras"`
		Audio   []interface{} `json:"audio"`
	} `json:"data"`
	MsgID string `json:"msg_id"`
}

type CalaosJsonSetState struct {
	Msg   string `json:"msg"`
	MsgID int    `json:"msg_id"`
	Data  struct {
		Id    string `json:"id"`
		Value string `json:"value"`
	} `json:"data"`
}

var loggedin bool
var home CalaosJsonMsgHome
var configFilename string
var config Configuration
var hapIOs []HapIO
var calaosIOs []CalaosIO
var websocketClient *websocket.Conn

func getIOFromId(id string) *CalaosIO {
	for i := range home.Data.Home {
		for j := range home.Data.Home[i].IOs {
			if home.Data.Home[i].IOs[j].ID == id {
				return &home.Data.Home[i].IOs[j]
			}
		}

	}
	return nil
}

func getNameFromId(id string) string {
	for i := range home.Data.Home {
		for j := range home.Data.Home[i].IOs {
			if home.Data.Home[i].IOs[j].ID == id {
				return home.Data.Home[i].IOs[j].Name
			}
		}

	}
	return ""

}

func setupCalaosHome(acc *accessory.Accessory) {
	for i := range home.Data.Home {
		for j := range home.Data.Home[i].IOs {
			cio := home.Data.Home[i].IOs[j]
			io, err := NewIO(cio)
			if err == nil {
				hapIOs = append(hapIOs, io)
				acc.AddService(io.Service())
			}
		}

	}
}

func CalaosUpdate(cio CalaosIO) {

	msg := CalaosJsonSetState{}
	msg.MsgID = 1
	msg.Msg = "set_state"
	msg.Data.Id = cio.ID
	msg.Data.Value = cio.State

	str, _ := json.Marshal(msg)

	if err := websocketClient.WriteMessage(websocket.TextMessage, []byte(str)); err != nil {
		log.Println("Write message error")
		return
	}
}

func main() {

	flag.StringVar(&configFilename, "config", "./config.json", "Get the config to use. default value is ./config.json")
	flag.Parse()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	log.Println("Opening Configuration filename : " + configFilename)
	file, err := os.Open(configFilename)
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Println("error:", err)
	}
	log.Println("Configuration : ")
	log.Println(config.WebSocketServer)

	loggedin = false

	log.Println("Opening :", "ws://"+config.WebSocketServer.Host+":"+strconv.Itoa(config.WebSocketServer.Port)+"/api")

	websocketClient, _, err = websocket.DefaultDialer.Dial("ws://"+config.WebSocketServer.Host+":"+strconv.Itoa(config.WebSocketServer.Port)+"/api", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer websocketClient.Close()

	done := make(chan struct{})

	// Send login message through Calaos websocket API
	msg := "{ \"msg\": \"login\", \"msg_id\": \"1\", \"data\": { \"cn_user\": \"" + config.WebSocketServer.User + "\", \"cn_pass\": \"" + config.WebSocketServer.Password + "\" } }"
	if err = websocketClient.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		log.Println("Write message error")
		return
	}

	go func() {
		defer websocketClient.Close()
		defer close(done)
		// Infinite loop
		for {
			_, message, err := websocketClient.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			//log.Printf("Receive message on Websocket: %s", message)
			// Try to decode JSON message
			var msg CalaosJsonMsg
			err = json.Unmarshal([]byte(message), &msg)
			if err != nil {
				log.Println("error:", err)
				continue
			}
			// Login message
			if msg.Msg == "login" {
				var loginMsg CalaosJsonMsgLogin
				err = json.Unmarshal([]byte(message), &loginMsg)
				if err != nil {
					log.Println("error:", err)
					continue
				}
				// Login success
				if loginMsg.Data.Success == "true" {
					loggedin = true
					log.Printf("Logged in")
					// We are logged in, send get_home message to get all IO states
					getHomeMsg := "{ \"msg\": \"get_home\", \"msg_id\": \"2\" }"
					if err = websocketClient.WriteMessage(websocket.TextMessage, []byte(getHomeMsg)); err != nil {
						log.Println("Write message error")
						continue
					}
				} else {
					loggedin = false
				}
			}

			// If we received and we are logged in
			if loggedin {
				// Msg event received
				if msg.Msg == "event" {
					var eventMsg CalaosJsonMsgEvent
					err = json.Unmarshal([]byte(message), &eventMsg)
					if err != nil {
						log.Println("error:", err)
						continue
					}
					// Get the calaos IO from the Map
					cio := getIOFromId(eventMsg.Data.Data.ID)
					if cio != nil {
						cio.State = eventMsg.Data.Data.State
						// Iterate HAP IOs to find the same ID thand Calaos IO
						for _, io := range hapIOs {
							if io.ID() == cio.ID {
								// Update HAP Io
								io.Update(cio)
							}
						}
					}
				}
				// Receive get_home message
				if msg.Msg == "get_home" {
					err = json.Unmarshal([]byte(message), &home)
					if err != nil {
						log.Println("error:", err)
					}
					// Set Accessory infos
					info := accessory.Info{
						Name:         "Calaos Gateway",
						Manufacturer: "Calaos",
					}
					// Create a new accessory of type Bridge
					bridge := accessory.New(info, accessory.TypeBridge)

					// Set the PIN code
					config := hc.Config{Pin: config.PinCode}
					// Get a copy of all Calaos IOs
					calaosIOs = home.Data.Home[0].IOs

					setupCalaosHome(bridge)
					// Associate Bridge and info to a new Ip transport
					transport, err := hc.NewIPTransport(config, bridge)
					if err != nil {
						log.Println(err)
						continue
					} else {
						log.Println("Start HAP")
						go transport.Start()
					}
				}
			}
		}
	}()

	// Wait for Ctrl + c to qui app and close websocket properly
	for {
		select {
		case <-sigc:
			log.Println("interrupt")
			err := websocketClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			websocketClient.Close()
			return
		}
	}

}
