package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLightDimmer(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-light-1",
		Name:    "Test Light",
		GuiType: CalaosGuiTypeLightDimmer,
		IoType:  "light_dimmer",
		State:   "50",
		Visible: "true",
	}

	acc := NewLightDimmer(cio, 12345)
	require.NotNil(t, acc)
	assert.NotNil(t, acc.Lightbulb)
	assert.NotNil(t, acc.Brightness)
	assert.NotNil(t, acc.Name)
	assert.Equal(t, uint64(12345), acc.Lightbulb.Id)
}

func TestLightDimmer_Update_LightDimmerType(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-light-1",
		Name:    "Test Light",
		GuiType: CalaosGuiTypeLightDimmer,
		State:   "50",
	}

	acc := NewLightDimmer(cio, 12345)

	tests := []struct {
		name           string
		state          string
		expectedBright int
		expectedOn     bool
		shouldError    bool
	}{
		{
			name:           "Valid brightness 50",
			state:          "50",
			expectedBright: 50,
			expectedOn:     true,
			shouldError:    false,
		},
		{
			name:           "Valid brightness 0 (should turn off)",
			state:          "0",
			expectedBright: 0,
			expectedOn:     false,
			shouldError:    false,
		},
		{
			name:           "Valid brightness 100",
			state:          "100",
			expectedBright: 100,
			expectedOn:     true,
			shouldError:    false,
		},
		{
			name:           "Valid brightness 1 (should turn on)",
			state:          "1",
			expectedBright: 1,
			expectedOn:     true,
			shouldError:    false,
		},
		{
			name:        "Invalid state (non-numeric)",
			state:       "abc",
			shouldError: true,
		},
		{
			name:        "Empty state",
			state:       "",
			shouldError: true,
		},
		{
			name:        "Out of range negative",
			state:       "-10",
			expectedBright: -10,
			expectedOn:     true, // Any non-zero value turns on
			shouldError: false, // Atoi will parse it
		},
		{
			name:        "Out of range > 100",
			state:       "150",
			expectedBright: 100, // HomeKit clamps brightness to 0-100
			expectedOn:     true, // Any non-zero value turns on
			shouldError: false, // Atoi will parse it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state
			updateCio.GuiType = CalaosGuiTypeLightDimmer

			err := acc.Update(&updateCio)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedBright >= 0 {
					assert.Equal(t, tt.expectedBright, acc.Brightness.Val)
					assert.Equal(t, tt.expectedOn, acc.Lightbulb.Lightbulb.On.Val)
				}
			}
		})
	}
}

func TestLightDimmer_Update_BooleanType(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-light-2",
		Name:    "Test Light",
		GuiType: CalaosGuiTypeLight, // Not light_dimmer, so uses boolean parsing
		State:   "true",
	}

	acc := NewLightDimmer(cio, 12346)

	tests := []struct {
		name           string
		state          string
		expectedBright int
		expectedOn     bool
		shouldError    bool
	}{
		{
			name:           "State true (should set brightness to 100)",
			state:          "true",
			expectedBright: 100,
			expectedOn:     true,
			shouldError:    false,
		},
		{
			name:           "State false (should set brightness to 0)",
			state:          "false",
			expectedBright: 0,
			expectedOn:     false,
			shouldError:    false,
		},
		{
			name:        "Invalid boolean string",
			state:       "maybe",
			shouldError: true,
		},
		{
			name:        "Empty state",
			state:       "",
			shouldError: true,
		},
		{
			name:        "Numeric string (not boolean)",
			state:       "50",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state
			updateCio.GuiType = CalaosGuiTypeLight

			err := acc.Update(&updateCio)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBright, acc.Brightness.Val)
				assert.Equal(t, tt.expectedOn, acc.Lightbulb.Lightbulb.On.Val)
			}
		})
	}
}

func TestLightDimmer_AccessoryGet(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-light-3",
		Name:    "Test Light",
		GuiType: CalaosGuiTypeLightDimmer,
		State:   "50",
	}

	acc := NewLightDimmer(cio, 12347)
	accessory := acc.AccessoryGet()

	require.NotNil(t, accessory)
	assert.Equal(t, acc.Lightbulb.A, accessory)
}

