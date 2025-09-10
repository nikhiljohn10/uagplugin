package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	CriticalLevel
	FatalLevel
)

var (
	currentLevel    = InfoLevel
	debugMode       = false
	logFile         *os.File
	logMu           sync.Mutex
	alertWebhookURL string
)

// SetAlertWebhook sets the webhook URL for critical alerts (empty disables alerting)
func SetAlertWebhook(url string) {
	alertWebhookURL = url
}

// SetLogFile sets the log file path for file output. Call this once at startup.
func SetLogFile(path string) error {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		logFile.Close()
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

func SetLevel(level Level) {
	currentLevel = level
}

func SetDebugMode(enabled bool) {
	debugMode = enabled
	if enabled {
		currentLevel = DebugLevel
	}
}

func IsDebugMode() bool {
	return debugMode
}

func logf(level Level, prefix string, format string, args ...any) {
	if level < currentLevel && !(level == DebugLevel && debugMode) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s %s\n", prefix, msg)
	fmt.Fprint(os.Stdout, line)
	logMu.Lock()
	if logFile != nil {
		logFile.WriteString(line)
	}
	logMu.Unlock()
	// Send alert if critical and webhook is set
	if level == CriticalLevel && alertWebhookURL != "" {
		go func(payload string) {
			http.Post(alertWebhookURL, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"message":%q}`, payload))))
		}(line)
	}
	if level == FatalLevel {
		if logFile != nil {
			logFile.Sync()
		}
		os.Exit(1)
	}
}

func Debug(format string, args ...any)    { logf(DebugLevel, "[DEBUG]", format, args...) }
func Info(format string, args ...any)     { logf(InfoLevel, "[INFO]", format, args...) }
func Warn(format string, args ...any)     { logf(WarnLevel, "[WARN]", format, args...) }
func Error(format string, args ...any)    { logf(ErrorLevel, "[ERROR]", format, args...) }
func Critical(format string, args ...any) { logf(CriticalLevel, "[CRITICAL]", format, args...) }
func Fatal(format string, args ...any)    { logf(FatalLevel, "[FATAL]", format, args...) }
