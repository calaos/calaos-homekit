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

// Calaos message types
const (
	CalaosMsgTypeLogin   = "login"
	CalaosMsgTypeEvent   = "event"
	CalaosMsgTypeGetHome = "get_home"
	CalaosMsgTypeSetState = "set_state"
)

// Calaos message IDs
const (
	CalaosMsgIDLogin   = "1"
	CalaosMsgIDGetHome = "2"
	CalaosMsgIDUserCmd = "user_cmd"
)

// Calaos boolean string values
const (
	CalaosVisibleFalse = "false"
	CalaosSuccessTrue = "true"
)

// Calaos GUI types
const (
	CalaosGuiTypeTemp         = "temp"
	CalaosGuiTypeAnalogIn     = "analog_in"
	CalaosGuiTypeLightDimmer  = "light_dimmer"
	CalaosGuiTypeLight        = "light"
	CalaosGuiTypeShutterSmart = "shutter_smart"
)

// Calaos IO styles
const (
	CalaosIOStyleHumidity = "humidity"
)

// WebSocket URI types
const (
	URITypeWS  = "ws"
	URITypeWSS = "wss"
)

// WebSocket ports
const (
	PortWSS = 443
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
			if cio.Visible != CalaosVisibleFalse {
				switch cio.GuiType {
				case CalaosGuiTypeTemp:
					acc = NewTemperatureSensor(cio, id)

				case CalaosGuiTypeAnalogIn:
					if cio.IoStyle == CalaosIOStyleHumidity {
						acc = NewHumiditySensor(cio, id)
					}

				case CalaosGuiTypeLightDimmer:
					acc = NewLightDimmer(cio, id)

				case CalaosGuiTypeLight:
					if cio.IoStyle == "" {
						acc = NewLightDimmer(cio, id)
					}

				//TODO:
				// case "shutter":
				// 	acc = NewWindowCovering(cio, id)

				case CalaosGuiTypeShutterSmart:
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
	msg.MsgID = CalaosMsgIDUserCmd
	msg.Msg = CalaosMsgTypeSetState
	msg.Data.Id = cio.ID
	msg.Data.Value = cio.State

	str, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("Failed to marshal CalaosUpdate message: %v", err)
		return
	}

	if err := websocketClient.WriteMessage(websocket.TextMessage, []byte(str)); err != nil {
		log.Error("Write message error")
		return
	}
}

// sendLoginMessage sends the initial login message to the Calaos WebSocket server
func sendLoginMessage() error {
	loginMsg := CalaosJsonMsgLoginRequest{
		Msg:   CalaosMsgTypeLogin,
		MsgID: CalaosMsgIDLogin,
	}
	loginMsg.Data.CNUser = config.WebSocketServer.User
	loginMsg.Data.CNPass = config.WebSocketServer.Password

	msgBytes, err := json.Marshal(loginMsg)
	if err != nil {
		return err
	}
	return websocketClient.WriteMessage(websocket.TextMessage, msgBytes)
}

// handleLoginMessage processes login response messages
func handleLoginMessage(message []byte) error {
	var loginMsg CalaosJsonMsgLogin
	if err := json.Unmarshal(message, &loginMsg); err != nil {
		return err
	}

	if loginMsg.Data.Success == CalaosSuccessTrue {
		loggedin = true
		log.Info("Logged in")
		// Send get_home message to get all IO states
		getHomeMsg := CalaosJsonGetHomeRequest{
			Msg:   CalaosMsgTypeGetHome,
			MsgID: CalaosMsgIDGetHome,
		}
		getHomeBytes, err := json.Marshal(getHomeMsg)
		if err != nil {
			return err
		}
		return websocketClient.WriteMessage(websocket.TextMessage, getHomeBytes)
	}
	loggedin = false
	return nil
}

// handleEventMessage processes event messages and updates accessory states
func handleEventMessage(message []byte) error {
	var eventMsg CalaosJsonMsgEvent
	if err := json.Unmarshal(message, &eventMsg); err != nil {
		return err
	}

	cio := getIOFromId(eventMsg.Data.Data.ID)
	if cio != nil {
		cio.State = eventMsg.Data.Data.State
		id := uint64(murmur.Sum32(cio.ID))
		if acc, found := accessories[id]; found {
			acc.Update(cio)
		}
	}
	return nil
}

// updateAccessoryStates updates existing accessories with current state from Calaos
func updateAccessoryStates() {
	for i := range home.Data.Home {
		for j := range home.Data.Home[i].IOs {
			cio := home.Data.Home[i].IOs[j]
			id := uint64(murmur.Sum32(cio.ID))
			if acc, found := accessories[id]; found {
				acc.Update(&cio)
			}
		}
	}
}

// startHAPServer initializes and starts the HAP server with all accessories
func startHAPServer(ctx context.Context) error {
	info := accessory.Info{
		Name:         config.BridgeName,
		Manufacturer: "Calaos",
		Model:        "calaos-homekit",
		Firmware:     "3.0.0",
	}
	bridge := accessory.NewBridge(info)

	// Associate Bridge and info to a new Ip transport
	accessories = make(map[uint64]CalaosAccessory)
	setupCalaosHome()

	if len(accessories) == 0 {
		return nil
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
		return err
	}

	log.Info("Starting HAP server")
	server.Pin = config.PinCode
	hapServerStarted = true

	// Run the server.
	go func() {
		log.Info("HAP server listening for connections")
		if err := server.ListenAndServe(ctx); err != nil {
			log.Errorf("HAP server error: %v", err)
			hapServerStarted = false
		}
	}()
	return nil
}

// handleGetHomeMessage processes get_home messages and either updates or initializes accessories
func handleGetHomeMessage(message []byte, ctx context.Context) error {
	if err := json.Unmarshal(message, &home); err != nil {
		return err
	}

	if len(home.Data.Home) == 0 {
		log.Warn("get_home message has no home data")
		return nil
	}

	// If server is already started, update existing accessories with current state
	if hapServerStarted {
		log.Info("HAP server already started, updating accessory states")
		updateAccessoryStates()
		return nil
	}

	// Start HAP server for the first time
	if err := startHAPServer(ctx); err != nil {
		return err
	}

	if len(accessories) == 0 {
		log.Warn("No accessories found to expose in HomeKit")
	}
	return nil
}

func connectedCb(ctx context.Context) {
	// Send login message through Calaos websocket API
	if err := sendLoginMessage(); err != nil {
		log.Errorf("Failed to send login message: %v", err)
		return
	}

	go func() {
		defer websocketClient.Close()
		// Infinite loop
		for {
			_, message, err := websocketClient.ReadMessage()
			if err != nil {
				log.Error("read:", err)
				return
			}

			// Try to decode JSON message
			var msg CalaosJsonMsg
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			// Login message
			if msg.Msg == CalaosMsgTypeLogin {
				if err := handleLoginMessage(message); err != nil {
					log.Errorf("Failed to handle login message: %v", err)
					continue
				}
			}

			// If we received and we are logged in
			if loggedin {
				// Msg event received
				if msg.Msg == CalaosMsgTypeEvent {
					if err := handleEventMessage(message); err != nil {
						log.Errorf("Failed to handle event message: %v", err)
						continue
					}
				}
				// Receive get_home message
				if msg.Msg == CalaosMsgTypeGetHome {
					if err := handleGetHomeMessage(message, ctx); err != nil {
						log.Errorf("Failed to handle get_home message: %v", err)
						continue
					}
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

	uriType := URITypeWS
	if config.WebSocketServer.Port == PortWSS {
		uriType = URITypeWSS
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
