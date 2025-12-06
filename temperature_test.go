package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemperatureSensor(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-temp-1",
		Name:    "Test Temperature",
		GuiType: CalaosGuiTypeTemp,
		IoType:  "temp",
		State:   "22.5",
		Visible: "true",
	}

	acc := NewTemperatureSensor(cio, 12345)
	require.NotNil(t, acc)
	assert.NotNil(t, acc.Thermometer)
	assert.Equal(t, uint64(12345), acc.Thermometer.Id)
	
	// Verify temperature limits are set
	assert.Equal(t, float64(-50), acc.Thermometer.TempSensor.CurrentTemperature.MinVal)
	assert.Equal(t, float64(50), acc.Thermometer.TempSensor.CurrentTemperature.MaxVal)
	assert.Equal(t, float64(0.1), acc.Thermometer.TempSensor.CurrentTemperature.StepVal)
}

func TestTemp_Update(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-temp-1",
		Name:    "Test Temperature",
		GuiType: CalaosGuiTypeTemp,
		State:   "22.5",
	}

	acc := NewTemperatureSensor(cio, 12345)

	tests := []struct {
		name        string
		state       string
		expected    float64
		shouldError bool
	}{
		{
			name:        "Valid temperature 22.5",
			state:       "22.5",
			expected:    22.5,
			shouldError: false,
		},
		{
			name:        "Valid temperature 0",
			state:       "0",
			expected:    0.0,
			shouldError: false,
		},
		{
			name:        "Valid negative temperature",
			state:       "-10.5",
			expected:    -10.5,
			shouldError: false,
		},
		{
			name:        "Valid temperature at minimum limit",
			state:       "-50",
			expected:    -50.0,
			shouldError: false,
		},
		{
			name:        "Valid temperature at maximum limit",
			state:       "50",
			expected:    50.0,
			shouldError: false,
		},
		{
			name:        "Valid temperature with decimal",
			state:       "21.7",
			expected:    21.700000762939453, // float32 precision
			shouldError: false,
		},
		{
			name:        "Invalid temperature (non-numeric)",
			state:       "abc",
			shouldError: true,
		},
		{
			name:        "Empty state",
			state:       "",
			shouldError: true,
		},
		{
			name:        "Temperature below minimum",
			state:       "-51",
			expected:    -50.0, // Clamped to minimum
			shouldError: false, // ParseFloat will succeed, but value is clamped
		},
		{
			name:        "Temperature above maximum",
			state:       "51",
			expected:    50.0, // Clamped to maximum
			shouldError: false, // ParseFloat will succeed, but value is clamped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCio := cio
			updateCio.State = tt.state

			err := acc.Update(&updateCio)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify the value was set (may be clamped by HomeKit limits)
				assert.Equal(t, tt.expected, acc.Thermometer.TempSensor.CurrentTemperature.Val)
			}
		})
	}
}

func TestTemp_AccessoryGet(t *testing.T) {
	cio := CalaosIO{
		ID:      "test-temp-2",
		Name:    "Test Temperature",
		GuiType: CalaosGuiTypeTemp,
		State:   "22.5",
	}

	acc := NewTemperatureSensor(cio, 12346)
	accessory := acc.AccessoryGet()

	require.NotNil(t, accessory)
	assert.Equal(t, acc.Thermometer.A, accessory)
}

