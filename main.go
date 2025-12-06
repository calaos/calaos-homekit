package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/gorilla/websocket"
	"github.com/vcaesar/murmur"

	log "github.com/sirupsen/logrus"
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
	BridgeName      string
}

type CalaosJsonMsg struct {
	Msg   string `json:"msg"`
	MsgID string `json:"msg_id"`
}

type CalaosJsonMsgLoginRequest struct {
	Msg   string `json:"msg"`
	MsgID string `json:"msg_id"`
	Data  struct {
		CNUser string `json:"cn_user"`
		CNPass string `json:"cn_pass"`
	} `json:"data"`
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
	IoStyle string `json:"io_style,omitempty"`
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
	MsgID string `json:"msg_id"`
	Data  struct {
		Id    string `json:"id"`
		Value string `json:"value"`
	} `json:"data"`
}

type CalaosJsonGetHomeRequest struct {
	Msg   string `json:"msg"`
	MsgID string `json:"msg_id"`
}

var loggedin bool
var home CalaosJsonMsgHome
var configFilename string
var config Configuration

var accessories map[uint64]CalaosAccessory

//var hapIOs []HapIO

var calaosIOs []CalaosIO
var websocketClient *WebSocketClient
var hapServerStarted bool

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

func setupCalaosHome() {
	for i := range home.Data.Home {
		for j := range home.Data.Home[i].IOs {

			cio := home.Data.Home[i].IOs[j]
			var acc CalaosAccessory
			id := uint64(murmur.Sum32(cio.ID))
			if cio.Visible != "false" {
				switch cio.GuiType {
				case "temp":
					acc = NewTemperatureSensor(cio, id)

				case "analog_in":
					if cio.IoStyle == "humidity" {
						acc = NewHumiditySensor(cio, id)
					}

				case "light_dimmer":
					acc = NewLightDimmer(cio, id)

				case "light":
					if cio.IoStyle == "" {
						acc = NewLightDimmer(cio, id)
					}

				//TODO:
				// case "shutter":
				// 	acc = NewWindowCovering(cio, id)

				case "shutter_smart":
					acc = NewSmartShutter(cio, id)
				}
				if acc != nil {
					accessories[id] = acc
				}
			}
		}
	}
}

func CalaosUpdate(cio CalaosIO) {

	msg := CalaosJsonSetState{}
	msg.MsgID = "user_cmd"
	msg.Msg = "set_state"
	msg.Data.Id = cio.ID
	msg.Data.Value = cio.State

	str, _ := json.Marshal(msg)

	if err := websocketClient.WriteMessage(websocket.TextMessage, []byte(str)); err != nil {
		log.Error("Write message error")
		return
	}
}

func connectedCb(ctx context.Context) {
	done := make(chan struct{})

	// Send login message through Calaos websocket API
	loginMsg := CalaosJsonMsgLoginRequest{
		Msg:   "login",
		MsgID: "1",
	}
	loginMsg.Data.CNUser = config.WebSocketServer.User
	loginMsg.Data.CNPass = config.WebSocketServer.Password

	msgBytes, err := json.Marshal(loginMsg)
	if err != nil {
		log.Errorf("Failed to marshal login message: %v", err)
		return
	}
	if err := websocketClient.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		log.Error("Write message error")
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
				log.Error("error:", err)
				continue
			}
			// Login message
			if msg.Msg == "login" {
				var loginMsg CalaosJsonMsgLogin
				err = json.Unmarshal([]byte(message), &loginMsg)
				if err != nil {
					log.Error("error:", err)
					continue
				}
				// Login success
				if loginMsg.Data.Success == "true" {
					loggedin = true
					log.Printf("Logged in")
					// We are logged in, send get_home message to get all IO states
					getHomeMsg := CalaosJsonGetHomeRequest{
						Msg:   "get_home",
						MsgID: "2",
					}
					getHomeBytes, err := json.Marshal(getHomeMsg)
					if err != nil {
						log.Errorf("Failed to marshal get_home message: %v", err)
						continue
					}
					if err = websocketClient.WriteMessage(websocket.TextMessage, getHomeBytes); err != nil {
						log.Error("Write message error")
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
						log.Error("error:", err)
						continue
					}
					// Get the calaos IO from the Map
					cio := getIOFromId(eventMsg.Data.Data.ID)
					if cio != nil {
						cio.State = eventMsg.Data.Data.State
						id := uint64(murmur.Sum32(cio.ID))
						if acc, found := accessories[id]; found {
							acc.Update(cio)
						}
					}
				}
				// Receive get_home message
				if msg.Msg == "get_home" {
					err = json.Unmarshal([]byte(message), &home)
					if err != nil {
						log.Error("error:", err)
						continue
					}
					
					if len(home.Data.Home) == 0 {
						log.Warn("get_home message has no home data")
						continue
					}
					
					// If server is already started, update existing accessories with current state
					if hapServerStarted {
						log.Info("HAP server already started, updating accessory states")
						// Update existing accessories with current state from Calaos
						for i := range home.Data.Home {
							for j := range home.Data.Home[i].IOs {
								cio := home.Data.Home[i].IOs[j]
								id := uint64(murmur.Sum32(cio.ID))
								if acc, found := accessories[id]; found {
									// Update the existing accessory with current state
									acc.Update(&cio)
								}
							}
						}
						continue
					}
					
					// Create a new accessory of type Bridge
					// bridge := NewCalaosGateway(config.BridgeName)

					info := accessory.Info{
						Name:         config.BridgeName,
						Manufacturer: "Calaos",
						Model:        "calaos-homekit",
						Firmware:     "3.0.0",
					}
					bridge := accessory.NewBridge(info)

					// Get a copy of all Calaos IOs
					calaosIOs = home.Data.Home[0].IOs

					// Associate Bridge and info to a new Ip transport
					accessories = make(map[uint64]CalaosAccessory)
					setupCalaosHome()
					
					if len(accessories) == 0 {
						log.Warn("No accessories found to expose in HomeKit")
						continue
					}
					
					list := []*accessory.A{}
					for _, acc := range accessories {
						list = append(list, acc.AccessoryGet())
					}

					// Store the data in the "/Calaos Gateway" directory.
					// Use absolute path to avoid issues with working directory
					storePath := "/root/Calaos Gateway"
					store := hap.NewFsStore(storePath)

					server, err := hap.NewServer(store, bridge.A, list...)
					if err != nil {
						log.Errorf("Failed to create HAP server: %v", err)
						continue
					}
					
					log.Info("Starting HAP server")
					// Set the PIN code
					server.Pin = config.PinCode

					// Mark server as started before launching
					hapServerStarted = true
					
					// Run the server.
					go func() {
						log.Info("HAP server listening for connections")
						if err := server.ListenAndServe(ctx); err != nil {
							log.Errorf("HAP server error: %v", err)
							hapServerStarted = false
						}
					}()
				}
			}
		}
	}()
}

func main() {
	log.Println("Starting Calaos-Homekit")
	flag.StringVar(&configFilename, "config", "./config.json", "Get the config to use. default value is ./config.json")
	flag.Parse()

	// Setup a listener for interrupts and SIGTERM signals to stop the server.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	log.Println("Opening Configuration filename : " + configFilename)
	file, err := os.Open(configFilename)
	if err != nil {
		log.Error("error:", err)
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Error("error:", err)
		file.Close()
		os.Exit(1)
	}
	log.Println("Configuration : ")
	log.Println(config.WebSocketServer)

	uriType := "ws"
	if config.WebSocketServer.Port == 443 {
		uriType = "wss"
	}
	calaosURI := uriType + "://" + config.WebSocketServer.Host + ":" + strconv.Itoa(config.WebSocketServer.Port) + "/api"

	loggedin = false

	log.Println("Opening :", calaosURI)

	websocketClient = Dial(calaosURI, func() { connectedCb(ctx) })

	// Wait for Ctrl + c to qui app and close websocket properly
	for {
		select {
		case <-c:
			// Stop delivering signals.
			defer signal.Stop(c)
			// Cancel the context to stop the server.
			defer cancel()

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
