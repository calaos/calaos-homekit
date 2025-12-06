package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func setupTestHome() CalaosJsonMsgHome {
	return CalaosJsonMsgHome{
		Data: struct {
			Home    []CalaosHome  `json:"home"`
			Cameras []interface{} `json:"cameras"`
			Audio   []interface{} `json:"audio"`
		}{
			Home: []CalaosHome{
				{
					Name: "Test Room",
					IOs: []CalaosIO{
						{
							ID:      "test-io-1",
							Name:    "Test Light",
							GuiType: CalaosGuiTypeLightDimmer,
							Visible: "true",
							State:   "50",
						},
						{
							ID:      "test-io-2",
							Name:    "Test Temperature",
							GuiType: CalaosGuiTypeTemp,
							Visible: "true",
							State:   "22.5",
						},
						{
							ID:      "test-io-3",
							Name:    "Hidden IO",
							GuiType: CalaosGuiTypeLight,
							Visible: "false",
							State:   "0",
						},
					},
				},
				{
					Name: "Second Room",
					IOs: []CalaosIO{
						{
							ID:      "test-io-4",
							Name:    "Second Room Light",
							GuiType: CalaosGuiTypeLightDimmer,
							Visible: "true",
							State:   "75",
						},
					},
				},
			},
		},
	}
}

func setupTestConfig() Configuration {
	return Configuration{
		WebSocketServer: WebSocketConfig{
			Host:     "localhost",
			Port:     5454,
			User:     "testuser",
			Password: "testpass",
		},
		PinCode:    "12345678",
		BridgeName: "Test Bridge",
	}
}

// Test JSON message marshaling
func TestCalaosJsonMsgLoginRequest_Marshal(t *testing.T) {
	msg := CalaosJsonMsgLoginRequest{
		Msg:   CalaosMsgTypeLogin,
		MsgID: CalaosMsgIDLogin,
	}
	msg.Data.CNUser = "testuser"
	msg.Data.CNPass = "testpass"

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeLogin, result["msg"])
	assert.Equal(t, CalaosMsgIDLogin, result["msg_id"])
	
	dataMap := result["data"].(map[string]interface{})
	assert.Equal(t, "testuser", dataMap["cn_user"])
	assert.Equal(t, "testpass", dataMap["cn_pass"])
}

func TestCalaosJsonSetState_Marshal(t *testing.T) {
	msg := CalaosJsonSetState{
		Msg:   CalaosMsgTypeSetState,
		MsgID: CalaosMsgIDUserCmd,
	}
	msg.Data.Id = "test-io-1"
	msg.Data.Value = "75"

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result CalaosJsonSetState
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeSetState, result.Msg)
	assert.Equal(t, CalaosMsgIDUserCmd, result.MsgID)
	assert.Equal(t, "test-io-1", result.Data.Id)
	assert.Equal(t, "75", result.Data.Value)
}

func TestCalaosJsonGetHomeRequest_Marshal(t *testing.T) {
	msg := CalaosJsonGetHomeRequest{
		Msg:   CalaosMsgTypeGetHome,
		MsgID: CalaosMsgIDGetHome,
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result CalaosJsonGetHomeRequest
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeGetHome, result.Msg)
	assert.Equal(t, CalaosMsgIDGetHome, result.MsgID)
}

// Test helper functions
func TestGetIOFromId(t *testing.T) {
	testHome := setupTestHome()
	home = testHome

	tests := []struct {
		name     string
		id       string
		expected *CalaosIO
		found    bool
	}{
		{
			name:     "IO found in first room",
			id:       "test-io-1",
			expected: &testHome.Data.Home[0].IOs[0],
			found:    true,
		},
		{
			name:     "IO found in second room",
			id:       "test-io-4",
			expected: &testHome.Data.Home[1].IOs[0],
			found:    true,
		},
		{
			name:     "IO not found",
			id:       "non-existent",
			expected: nil,
			found:    false,
		},
		{
			name:     "Empty ID",
			id:       "",
			expected: nil,
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIOFromId(tt.id)
			if tt.found {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestGetNameFromId(t *testing.T) {
	testHome := setupTestHome()
	home = testHome

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "Name found",
			id:       "test-io-1",
			expected: "Test Light",
		},
		{
			name:     "Name found in second room",
			id:       "test-io-4",
			expected: "Second Room Light",
		},
		{
			name:     "Name not found",
			id:       "non-existent",
			expected: "",
		},
		{
			name:     "Empty ID",
			id:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNameFromId(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test message handling (requires mocking)
func TestHandleLoginMessage_Success(t *testing.T) {
	// This test would require mocking websocketClient
	// For now, we test the JSON unmarshaling part
	successMsg := `{
		"msg": "login",
		"msg_id": "1",
		"data": {
			"success": "true"
		}
	}`

	var loginMsg CalaosJsonMsgLogin
	err := json.Unmarshal([]byte(successMsg), &loginMsg)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeLogin, loginMsg.Msg)
	assert.Equal(t, CalaosSuccessTrue, loginMsg.Data.Success)
}

func TestHandleLoginMessage_Failure(t *testing.T) {
	failureMsg := `{
		"msg": "login",
		"msg_id": "1",
		"data": {
			"success": "false"
		}
	}`

	var loginMsg CalaosJsonMsgLogin
	err := json.Unmarshal([]byte(failureMsg), &loginMsg)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeLogin, loginMsg.Msg)
	assert.NotEqual(t, CalaosSuccessTrue, loginMsg.Data.Success)
}

func TestHandleEventMessage_Valid(t *testing.T) {
	// Setup test home
	testHome := setupTestHome()
	home = testHome

	eventMsgJSON := `{
		"msg": "event",
		"data": {
			"event_raw": "test",
			"data": {
				"id": "test-io-1",
				"state": "75"
			},
			"type_str": "test",
			"type": "test"
		}
	}`

	var eventMsg CalaosJsonMsgEvent
	err := json.Unmarshal([]byte(eventMsgJSON), &eventMsg)
	require.NoError(t, err)

	assert.Equal(t, CalaosMsgTypeEvent, eventMsg.Msg)
	assert.Equal(t, "test-io-1", eventMsg.Data.Data.ID)
	assert.Equal(t, "75", eventMsg.Data.Data.State)

	// Verify we can find the IO
	cio := getIOFromId(eventMsg.Data.Data.ID)
	require.NotNil(t, cio)
	assert.Equal(t, "test-io-1", cio.ID)
}

func TestHandleEventMessage_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var eventMsg CalaosJsonMsgEvent
	err := json.Unmarshal([]byte(invalidJSON), &eventMsg)
	assert.Error(t, err)
}

// Test constants
func TestConstants(t *testing.T) {
	assert.Equal(t, "login", CalaosMsgTypeLogin)
	assert.Equal(t, "event", CalaosMsgTypeEvent)
	assert.Equal(t, "get_home", CalaosMsgTypeGetHome)
	assert.Equal(t, "set_state", CalaosMsgTypeSetState)
	
	assert.Equal(t, "1", CalaosMsgIDLogin)
	assert.Equal(t, "2", CalaosMsgIDGetHome)
	assert.Equal(t, "user_cmd", CalaosMsgIDUserCmd)
	
	assert.Equal(t, "false", CalaosVisibleFalse)
	assert.Equal(t, "true", CalaosSuccessTrue)
	
	assert.Equal(t, "temp", CalaosGuiTypeTemp)
	assert.Equal(t, "light_dimmer", CalaosGuiTypeLightDimmer)
	assert.Equal(t, "shutter_smart", CalaosGuiTypeShutterSmart)
	
	assert.Equal(t, "humidity", CalaosIOStyleHumidity)
	
	assert.Equal(t, "ws", URITypeWS)
	assert.Equal(t, "wss", URITypeWSS)
	assert.Equal(t, 443, PortWSS)
}

