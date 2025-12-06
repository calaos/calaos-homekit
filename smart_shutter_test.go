package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSmartShutter(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-1",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		IoType:  "shutter_smart",
		State:   "up 0",
		Visible: "true",
	}

	acc := NewSmartShutter(cio, 12345)
	require.NotNil(t, acc)
	assert.NotNil(t, acc.WindowCovering)
	assert.NotNil(t, acc.HoldPosition)
	assert.NotNil(t, acc.Name)
	assert.Equal(t, uint64(12345), acc.WindowCovering.Id)
}

func TestSmartShutter_Update_UpCommand(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-1",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "up 0",
	}

	acc := NewSmartShutter(cio, 12345)

	tests := []struct {
		name              string
		state             string
		expectedPosition  int
		expectedState     int
		expectedHold      bool
		shouldError       bool
	}{
		{
			name:             "Up command with position 0 (Calaos open = HomeKit 100)",
			state:            "up 0",
			expectedPosition: 100, // 100 - 0 = 100
			expectedState:    OPENING,
			expectedHold:     false,
			shouldError:      false,
		},
		{
			name:             "Up command with position 50 (Calaos 50 = HomeKit 50)",
			state:            "up 50",
			expectedPosition: 50, // 100 - 50 = 50
			expectedState:    OPENING,
			expectedHold:     false,
			shouldError:      false,
		},
		{
			name:             "Up command with position 100 (Calaos closed = HomeKit 0)",
			state:            "up 100",
			expectedPosition: 0, // 100 - 100 = 0
			expectedState:    OPENING,
			expectedHold:     false,
			shouldError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state
			updateCio.GuiType = CalaosGuiTypeShutterSmart

			err := acc.Update(&updateCio)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPosition, acc.WindowCovering.WindowCovering.CurrentPosition.Val)
				assert.Equal(t, tt.expectedPosition, acc.WindowCovering.WindowCovering.TargetPosition.Val)
				assert.Equal(t, tt.expectedState, acc.WindowCovering.WindowCovering.PositionState.Val)
				if holdVal, ok := acc.HoldPosition.Val.(bool); ok {
					assert.Equal(t, tt.expectedHold, holdVal)
				} else {
					// HoldPosition may be nil/unset for up/down commands
					assert.False(t, tt.expectedHold, "HoldPosition should be false when nil")
				}
			}
		})
	}
}

func TestSmartShutter_Update_DownCommand(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-2",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "down 0",
	}

	acc := NewSmartShutter(cio, 12346)

	tests := []struct {
		name              string
		state             string
		expectedPosition  int
		expectedState     int
		shouldError       bool
	}{
		{
			name:             "Down command with position 0",
			state:            "down 0",
			expectedPosition: 100,
			expectedState:    CLOSING,
			shouldError:      false,
		},
		{
			name:             "Down command with position 50",
			state:            "down 50",
			expectedPosition: 50,
			expectedState:    CLOSING,
			shouldError:      false,
		},
		{
			name:             "Down command with position 100",
			state:            "down 100",
			expectedPosition: 0,
			expectedState:    CLOSING,
			shouldError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state
			updateCio.GuiType = CalaosGuiTypeShutterSmart

			err := acc.Update(&updateCio)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPosition, acc.WindowCovering.WindowCovering.CurrentPosition.Val)
				assert.Equal(t, tt.expectedPosition, acc.WindowCovering.WindowCovering.TargetPosition.Val)
				assert.Equal(t, tt.expectedState, acc.WindowCovering.WindowCovering.PositionState.Val)
			}
		})
	}
}

func TestSmartShutter_Update_StopCommand(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-3",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "stop 50",
	}

	acc := NewSmartShutter(cio, 12347)

	updateCio := cio
	updateCio.State = "stop 50"
	updateCio.GuiType = CalaosGuiTypeShutterSmart

	err := acc.Update(&updateCio)
	assert.NoError(t, err)

	assert.Equal(t, 50, acc.WindowCovering.WindowCovering.CurrentPosition.Val)
	assert.Equal(t, 50, acc.WindowCovering.WindowCovering.TargetPosition.Val)
	assert.Equal(t, STOPPED, acc.WindowCovering.WindowCovering.PositionState.Val)
	assert.Equal(t, true, acc.HoldPosition.Val)
}

func TestSmartShutter_Update_InvalidStates(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-4",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "up 0",
	}

	acc := NewSmartShutter(cio, 12348)

	tests := []struct {
		name        string
		state       string
		shouldError bool
	}{
		{
			name:        "Invalid state format (no value)",
			state:       "up",
			shouldError: true,
		},
		{
			name:        "Invalid state format (non-numeric value)",
			state:       "up abc",
			shouldError: true,
		},
		{
			name:        "Empty state",
			state:       "",
			shouldError: true, // Will cause index out of bounds in strings.Fields
		},
		{
			name:        "Unknown command",
			state:       "unknown 50",
			shouldError: false, // Returns nil, doesn't update
		},
		{
			name:        "Calibration command (not implemented)",
			state:       "calibration 0",
			shouldError: false, // Returns nil, doesn't update
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state
			updateCio.GuiType = CalaosGuiTypeShutterSmart

			if tt.state == "" {
				// Empty state causes panic due to index out of bounds
				assert.Panics(t, func() {
					_ = acc.Update(&updateCio)
				})
			} else {
				err := acc.Update(&updateCio)
				if tt.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestSmartShutter_Update_NonShutterType(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-5",
		Name:    "Test Shutter",
		GuiType: "other_type", // Not shutter_smart
		State:   "up 0",
	}

	acc := NewSmartShutter(cio, 12349)

	updateCio := cio
	updateCio.GuiType = "other_type"

	err := acc.Update(&updateCio)
	assert.NoError(t, err) // Returns nil for non-shutter types
}

func TestSmartShutter_AccessoryGet(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-shutter-6",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "up 0",
	}

	acc := NewSmartShutter(cio, 12350)
	accessory := acc.AccessoryGet()

	require.NotNil(t, accessory)
	assert.Equal(t, acc.WindowCovering.A, accessory)
}

func TestSmartShutter_PositionInversion(t *testing.T) {
	// Test that Calaos position values are correctly inverted to HomeKit values
	cio := CalaosIO{
		ID:      "test-shutter-7",
		Name:    "Test Shutter",
		GuiType: CalaosGuiTypeShutterSmart,
		State:   "up 0",
	}

	acc := NewSmartShutter(cio, 12351)

	testCases := []struct {
		calaosPos int // Calaos position (0 = open, 100 = closed)
		homekitPos int // Expected HomeKit position (100 = open, 0 = closed)
	}{
		{0, 100},   // Calaos open = HomeKit fully open
		{50, 50},   // Calaos middle = HomeKit middle
		{100, 0},   // Calaos closed = HomeKit fully closed
		{25, 75},   // Calaos 25% closed = HomeKit 75% open
		{75, 25},   // Calaos 75% closed = HomeKit 25% open
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Calaos_%d_to_HomeKit_%d", tc.calaosPos, tc.homekitPos), func(t *testing.T) {
			updateCio := cio
			updateCio.State = fmt.Sprintf("up %d", tc.calaosPos)
			updateCio.GuiType = CalaosGuiTypeShutterSmart

			err := acc.Update(&updateCio)
			assert.NoError(t, err)
			assert.Equal(t, tc.homekitPos, acc.WindowCovering.WindowCovering.CurrentPosition.Val)
		})
	}
}

