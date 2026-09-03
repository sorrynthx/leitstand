package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	file    *os.File
	logger  *log.Logger
	logPath = "leitstand.log"
)

// Init initializes file-based logging.
func Init(customPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if customPath != "" {
		logPath = customPath
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	file = f
	logger = log.New(file, "", 0)

	logMessageLocked("INFO", "==================================================")
	logMessageLocked("INFO", "⚡ Leitstand Log Initialized at %s", time.Now().Format("2006-01-02 15:04:05"))
	logMessageLocked("INFO", "==================================================")
	return nil
}

// Close flushes and closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}

const (
	MaxLogSizeBytes int64 = 5 * 1024 * 1024 // 5 MB
)

func rotateLocked() {
	if file == nil {
		return
	}
	fi, err := file.Stat()
	if err != nil || fi.Size() < MaxLogSizeBytes {
		return
	}

	_ = file.Close()
	file = nil
	logger = nil

	backupPath := logPath + ".1"
	_ = os.Remove(backupPath)
	_ = os.Rename(logPath, backupPath)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		file = f
		logger = log.New(file, "", 0)
		timestamp := time.Now().Format("2006/01/02 15:04:05.000")
		logger.Printf("[%s] [INFO] 🔄 Log rotated: archived previous log to %s", timestamp, backupPath)
	}
}

func logMessageLocked(level, format string, args ...interface{}) {
	rotateLocked()

	if logger == nil {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			file = f
			logger = log.New(file, "", 0)
		} else {
			return
		}
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	logger.Printf("[%s] [%s] %s", timestamp, level, msg)
}

func logMessage(level, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	logMessageLocked(level, format, args...)
}

func Info(msg string) {
	logMessage("INFO", "%s", msg)
}

func Infof(format string, args ...interface{}) {
	logMessage("INFO", format, args...)
}

func Warnf(format string, args ...interface{}) {
	logMessage("WARN", format, args...)
}

func Errorf(format string, args ...interface{}) {
	logMessage("ERROR", format, args...)
}

func Debugf(format string, args ...interface{}) {
	logMessage("DEBUG", format, args...)
}
