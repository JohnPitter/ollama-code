package session

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	if mgr.sessionDir == "" {
		t.Error("Session directory should be set")
	}
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	name := "Test Session"
	workDir := "/test/dir"
	mode := "interactive"

	session, err := mgr.New(name, workDir, mode)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session == nil {
		t.Fatal("Session should not be nil")
	}

	if session.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, session.Name)
	}

	if session.WorkDir != workDir {
		t.Errorf("Expected workdir '%s', got '%s'", workDir, session.WorkDir)
	}

	if session.Mode != mode {
		t.Errorf("Expected mode '%s', got '%s'", mode, session.Mode)
	}

	if !session.Active {
		t.Error("New session should be active")
	}

	if session.ID == "" {
		t.Error("Session ID should be generated")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	session, _ := mgr.New("Test", "/test", "interactive")

	err := mgr.Save()
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Verify current session is saved
	if mgr.currentSession != session {
		t.Error("Current session should be set")
	}
}

func TestEnd(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	session, _ := mgr.New("Test", "/test", "interactive")

	if !session.Active {
		t.Fatal("Session should start active")
	}

	err := mgr.End()
	if err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	if session.Active {
		t.Error("Session should be inactive after end")
	}

	if mgr.currentSession != nil {
		t.Error("Current session should be nil after end")
	}
}

func TestEnd_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	err := mgr.End()
	if err != nil {
		t.Error("Ending non-existent session should not error")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create multiple sessions
	mgr.New("Session 1", "/test1", "interactive")
	mgr.End()

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	mgr.New("Session 2", "/test2", "readonly")
	mgr.End()

	sessions, err := mgr.List(10)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) < 2 {
		t.Errorf("Expected at least 2 sessions, got %d", len(sessions))
	}

	// Sessions should be ordered by last activity (newest first)
	if len(sessions) >= 2 {
		if sessions[0].Name != "Session 2" {
			t.Error("Sessions should be ordered by last activity")
		}
	}
}

func TestList_Limit(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create 3 sessions
	for i := 1; i <= 3; i++ {
		mgr.New("Session", "/test", "interactive")
		mgr.End()
		time.Sleep(10 * time.Millisecond)
	}

	// List with limit of 2
	sessions, err := mgr.List(2)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions (limit), got %d", len(sessions))
	}
}

func TestGetCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// No current session
	current := mgr.GetCurrent()
	if current != nil {
		t.Error("Should have no current session initially")
	}

	// Create session
	session, _ := mgr.New("Test", "/test", "interactive")

	// Get current
	current = mgr.GetCurrent()
	if current == nil {
		t.Fatal("Should have current session")
	}

	if current.ID != session.ID {
		t.Error("Current session ID mismatch")
	}
}

func TestSessionMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	session, _ := mgr.New("Test", "/test", "interactive")

	// Verify metadata is initialized
	if session.Metadata == nil {
		t.Error("Metadata should be initialized")
	}

	// Should be able to add metadata
	session.Metadata["key"] = "value"

	err := mgr.Save()
	if err != nil {
		t.Fatalf("Failed to save with metadata: %v", err)
	}
}

func TestSessionMessages(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	session, _ := mgr.New("Test", "/test", "interactive")

	// Messages should be initialized
	if session.Messages == nil {
		t.Error("Messages should be initialized")
	}

	if len(session.Messages) != 0 {
		t.Error("New session should have no messages")
	}
}

func TestSessionTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	before := time.Now()
	session, _ := mgr.New("Test", "/test", "interactive")
	after := time.Now()

	// Start time should be between before and after
	if session.StartTime.Before(before) || session.StartTime.After(after) {
		t.Error("Start time should be set to current time")
	}

	// Last activity should be similar to start time
	if session.LastActivity.Before(before) || session.LastActivity.After(after) {
		t.Error("Last activity should be set to current time")
	}
}
