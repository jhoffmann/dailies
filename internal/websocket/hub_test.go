package websocket

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("Hub clients map is nil")
	}
	if hub.broadcast == nil {
		t.Error("Hub broadcast channel is nil")
	}
	if hub.register == nil {
		t.Error("Hub register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("Hub unregister channel is nil")
	}
}

func TestBroadcast(t *testing.T) {
	hub := NewHub()

	// Start the hub in a goroutine
	go hub.Run()

	// Give the hub a moment to start
	time.Sleep(10 * time.Millisecond)

	// Test broadcasting a message (should not panic even with no clients)
	hub.Broadcast("test", "test message", nil)

	// Test should complete without hanging
	time.Sleep(10 * time.Millisecond)
}

func TestGetHub(t *testing.T) {
	hub1 := GetHub()
	hub2 := GetHub()

	if hub1 != hub2 {
		t.Error("GetHub() should return the same instance (singleton)")
	}

	if hub1 == nil {
		t.Error("GetHub() returned nil")
	}
}

func TestBroadcastHelpers(t *testing.T) {
	// These functions should not panic
	BroadcastTaskReset("test task", "daily", 1)
	BroadcastTaskUpdated("test-id", "test task", true)
	BroadcastTaskListRefresh()

	// Give time for any goroutines to process
	time.Sleep(10 * time.Millisecond)
}
