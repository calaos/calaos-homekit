package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHumiditySensor(t *testing.T) {
	cio := CalaosIO{
		ID:       "test-humidity-1",
		Name:     "Test Humidity",
		GuiType:  CalaosGuiTypeAnalogIn,
		IoStyle:  CalaosIOStyleHumidity,
		IoType:   "analog_in",
		State:    "65.5",
		Visible:  "true",
	}

	acc := NewHumiditySensor(cio, 12345)
	require.NotNil(t, acc)
	assert.NotNil(t, acc.HumiditySensor)
	assert.Equal(t, uint64(12345), acc.HumiditySensor.Id)
}

func TestHumidity_Update(t *testing.T) {
	cio := CalaosIO{
		ID:       "test-humidity-1",
		Name:     "Test Humidity",
		GuiType:  CalaosGuiTypeAnalogIn,
		IoStyle:  CalaosIOStyleHumidity,
		State:    "65.5",
	}

	acc := NewHumiditySensor(cio, 12345)

	tests := []struct {
		name     string
		state    string
		expected float64
	}{
		{
			name:     "Valid humidity 65.5",
			state:    "65.5",
			expected: 65.5,
		},
		{
			name:     "Valid humidity 0",
			state:    "0",
			expected: 0.0,
		},
		{
			name:     "Valid humidity 100",
			state:    "100",
			expected: 100.0,
		},
		{
			name:     "Valid humidity with decimal",
			state:    "45.7",
			expected: 45.70000076293945, // float32 precision
		},
		{
			name:     "Invalid humidity (non-numeric) - should set to 0.0",
			state:    "abc",
			expected: 0.0,
		},
		{
			name:     "Empty state - should set to 0.0",
			state:    "",
			expected: 0.0,
		},
		{
			name:     "Negative value - should parse but clamped to 0",
			state:    "-10",
			expected: 0.0, // HomeKit clamps to 0-100
		},
		{
			name:     "Value above 100 - clamped to 100",
			state:    "150",
			expected: 100.0, // HomeKit clamps to 0-100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state

			// Humidity.Update always returns nil, even on parse errors
			err := acc.Update(&updateCio)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, acc.HumiditySensor.CurrentRelativeHumidity.Val)
		})
	}
}

func TestHumidity_AccessoryGet(t *testing.T) {
	cio := CalaosIO{
		ID:       "test-humidity-2",
		Name:     "Test Humidity",
		GuiType:  CalaosGuiTypeAnalogIn,
		IoStyle:  CalaosIOStyleHumidity,
		State:    "65.5",
	}

	acc := NewHumiditySensor(cio, 12346)
	accessory := acc.AccessoryGet()

	// Note: acc.A is nil in current implementation (commented out)
	// This test verifies the method doesn't panic
	_ = accessory
}

