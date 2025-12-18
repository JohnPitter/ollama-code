package modes

import (
	"testing"
)

func TestModeReadOnly(t *testing.T) {
	mode := ModeReadOnly

	if mode.AllowsWrites() {
		t.Error("ReadOnly mode should not allow writes")
	}

	if mode.RequiresConfirmation() {
		t.Error("ReadOnly mode should not require confirmation")
	}

	desc := mode.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestModeInteractive(t *testing.T) {
	mode := ModeInteractive

	if !mode.AllowsWrites() {
		t.Error("Interactive mode should allow writes")
	}

	if !mode.RequiresConfirmation() {
		t.Error("Interactive mode should require confirmation")
	}

	desc := mode.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestModeAutonomous(t *testing.T) {
	mode := ModeAutonomous

	if !mode.AllowsWrites() {
		t.Error("Autonomous mode should allow writes")
	}

	if mode.RequiresConfirmation() {
		t.Error("Autonomous mode should not require confirmation")
	}

	desc := mode.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

// Note: There is no ParseMode function exported.
// Mode validation is done in config.Validate()

func TestModeString(t *testing.T) {
	tests := []struct {
		mode     OperationMode
		expected string
	}{
		{ModeReadOnly, "readonly"},
		{ModeInteractive, "interactive"},
		{ModeAutonomous, "autonomous"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := string(tt.mode); got != tt.expected {
				t.Errorf("Mode string = %v, want %v", got, tt.expected)
			}
		})
	}
}
