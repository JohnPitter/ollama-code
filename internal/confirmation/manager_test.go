package confirmation

import (
	"bufio"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
	if mgr.reader == nil {
		t.Error("Reader should not be nil")
	}
}

func TestConfirm_Yes(t *testing.T) {
	inputs := []string{"s\n", "sim\n", "y\n", "yes\n"}
	for _, input := range inputs {
		mgr := NewManager()
		mgr.reader = bufio.NewReader(strings.NewReader(input))
		confirmed, err := mgr.Confirm("Test", "")
		if err != nil || !confirmed {
			t.Errorf("Failed with '%s'", input)
		}
	}
}

func TestConfirm_No(t *testing.T) {
	inputs := []string{"n\n", "no\n", "invalid\n"}
	for _, input := range inputs {
		mgr := NewManager()
		mgr.reader = bufio.NewReader(strings.NewReader(input))
		confirmed, _ := mgr.Confirm("Test", "")
		if confirmed {
			t.Errorf("Should reject '%s'", input)
		}
	}
}

func TestConfirmWithPreview(t *testing.T) {
	mgr := NewManager()
	mgr.reader = bufio.NewReader(strings.NewReader("s\n"))
	confirmed, err := mgr.ConfirmWithPreview("Test", "Preview")
	if err != nil || !confirmed {
		t.Error("Should confirm")
	}
}

func TestConfirmDangerous_Accept(t *testing.T) {
	mgr := NewManager()
	mgr.reader = bufio.NewReader(strings.NewReader("CONFIRMO\n"))
	confirmed, err := mgr.ConfirmDangerousAction("Test", "Warning")
	if err != nil || !confirmed {
		t.Error("Should confirm with CONFIRMO")
	}
}

func TestConfirmDangerous_Reject(t *testing.T) {
	inputs := []string{"confirmo\n", "yes\n", "\n"}
	for _, input := range inputs {
		mgr := NewManager()
		mgr.reader = bufio.NewReader(strings.NewReader(input))
		confirmed, _ := mgr.ConfirmDangerousAction("Test", "")
		if confirmed {
			t.Errorf("Should reject '%s'", input)
		}
	}
}
