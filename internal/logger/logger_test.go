package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	tmpLog := "test_leitstand.log"
	defer os.Remove(tmpLog)

	err := Init(tmpLog)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Info("Test message")
	Infof("Formatted message: %d", 42)
	Warnf("Warning message")
	Errorf("Error message")
	Debugf("Debug message")
}

func TestLogRotation(t *testing.T) {
	tmpLog := "test_rotation.log"
	defer os.Remove(tmpLog)
	defer os.Remove(tmpLog + ".1")

	err := Init(tmpLog)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// Force rotation by checking rotateLocked behavior
	mu.Lock()
	rotateLocked()
	mu.Unlock()

	Info("Post-rotation message")
}
