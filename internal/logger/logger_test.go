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
