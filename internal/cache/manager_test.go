package cache

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	ttl := 5 * time.Minute
	mgr := NewManager(ttl)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	if mgr.cache == nil {
		t.Error("Cache map should be initialized")
	}

	if mgr.ttl != ttl {
		t.Errorf("Expected TTL %v, got %v", ttl, mgr.ttl)
	}
}

func TestSetAndGet(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	key := "test_key"
	value := "test_value"

	// Set value
	mgr.Set(key, value)

	// Get value
	result, found := mgr.Get(key)
	if !found {
		t.Fatal("Expected to find cached value")
	}

	if result != value {
		t.Errorf("Expected value '%s', got '%s'", value, result)
	}
}

func TestGet_NonExistent(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	result, found := mgr.Get("nonexistent")
	if found {
		t.Error("Should not find non-existent key")
	}

	if result != nil {
		t.Error("Result should be nil for non-existent key")
	}
}

func TestGet_Expired(t *testing.T) {
	mgr := NewManager(100 * time.Millisecond)

	key := "test_key"
	value := "test_value"

	mgr.Set(key, value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	result, found := mgr.Get(key)
	if found {
		t.Error("Should not find expired entry")
	}

	if result != nil {
		t.Error("Result should be nil for expired entry")
	}
}

func TestDelete(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	key := "test_key"
	value := "test_value"

	mgr.Set(key, value)

	// Verify it exists
	_, found := mgr.Get(key)
	if !found {
		t.Fatal("Value should exist before delete")
	}

	// Delete
	mgr.Delete(key)

	// Verify it's gone
	_, found = mgr.Get(key)
	if found {
		t.Error("Value should not exist after delete")
	}
}

func TestClear(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	// Add multiple entries
	mgr.Set("key1", "value1")
	mgr.Set("key2", "value2")
	mgr.Set("key3", "value3")

	// Clear all
	mgr.Clear()

	// Verify all are gone
	_, found1 := mgr.Get("key1")
	_, found2 := mgr.Get("key2")
	_, found3 := mgr.Get("key3")

	if found1 || found2 || found3 {
		t.Error("All entries should be cleared")
	}
}

func TestSetOverwrite(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	key := "test_key"
	value1 := "value1"
	value2 := "value2"

	mgr.Set(key, value1)

	result, _ := mgr.Get(key)
	if result != value1 {
		t.Errorf("Expected '%s', got '%s'", value1, result)
	}

	// Overwrite
	mgr.Set(key, value2)

	result, _ = mgr.Get(key)
	if result != value2 {
		t.Errorf("Expected '%s', got '%s'", value2, result)
	}
}

func TestConcurrentAccess(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	// Test concurrent writes and reads
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mgr.Set("concurrent_key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mgr.Get("concurrent_key")
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// No panic = success
}

func TestDifferentTypes(t *testing.T) {
	mgr := NewManager(5 * time.Minute)

	// Test with different types
	mgr.Set("string", "value")
	mgr.Set("int", 42)
	mgr.Set("bool", true)
	mgr.Set("struct", struct{ Name string }{Name: "test"})

	strVal, _ := mgr.Get("string")
	if strVal.(string) != "value" {
		t.Error("String value mismatch")
	}

	intVal, _ := mgr.Get("int")
	if intVal.(int) != 42 {
		t.Error("Int value mismatch")
	}

	boolVal, _ := mgr.Get("bool")
	if boolVal.(bool) != true {
		t.Error("Bool value mismatch")
	}

	structVal, _ := mgr.Get("struct")
	if structVal.(struct{ Name string }).Name != "test" {
		t.Error("Struct value mismatch")
	}
}
