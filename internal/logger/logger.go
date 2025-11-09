package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Event represents a honeypot event
type Event struct {
	Timestamp   string                 `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Service     string                 `json:"service"`
	SourceIP    string                 `json:"source_ip"`
	SourcePort  int                    `json:"source_port,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Password    string                 `json:"password,omitempty"`
	Command     string                 `json:"command,omitempty"`
	Message     string                 `json:"message,omitempty"`
	ExtraData   map[string]interface{} `json:"extra_data,omitempty"`
}

// Logger handles all logging for the honeypot
type Logger struct {
	logDir      string
	attackLog   *os.File
	consoleLog  *log.Logger
	alertChan   chan Event
}

// New creates a new logger instance
func New(logDir string) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open attack log file
	attackLogPath := filepath.Join(logDir, "attacks.jsonl")
	attackLog, err := os.OpenFile(attackLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open attack log: %v", err)
	}

	// Setup console logger
	consoleLog := log.New(os.Stdout, "", log.LstdFlags)

	l := &Logger{
		logDir:     logDir,
		attackLog:  attackLog,
		consoleLog: consoleLog,
		alertChan:  make(chan Event, 100),
	}

	return l, nil
}

// LogConnection logs a connection attempt
func (l *Logger) LogConnection(service, ip string, port int, username, password string) Event {
	event := Event{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		EventType:  "connection",
		Service:    service,
		SourceIP:   ip,
		SourcePort: port,
		Username:   username,
		Password:   password,
	}

	l.writeEvent(event)

	msg := fmt.Sprintf("[%s] Connection from %s:%d", service, ip, port)
	if username != "" {
		msg += fmt.Sprintf(" - Username: %s", username)
	}
	if password != "" {
		msg += fmt.Sprintf(" - Password: %s", password)
	}
	l.consoleLog.Println(msg)

	// Send to alert channel
	select {
	case l.alertChan <- event:
	default:
	}

	return event
}

// LogCommand logs a command execution attempt
func (l *Logger) LogCommand(service, ip, command string) Event {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: "command",
		Service:   service,
		SourceIP:  ip,
		Command:   command,
	}

	l.writeEvent(event)

	msg := fmt.Sprintf("[%s] Command from %s: %s", service, ip, command)
	l.consoleLog.Println(msg)

	// Send to alert channel
	select {
	case l.alertChan <- event:
	default:
	}

	return event
}

// LogEvent logs a generic event
func (l *Logger) LogEvent(service, eventType, ip, message string, extraData map[string]interface{}) Event {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Service:   service,
		SourceIP:  ip,
		Message:   message,
		ExtraData: extraData,
	}

	l.writeEvent(event)

	msg := fmt.Sprintf("[%s] %s from %s", service, message, ip)
	l.consoleLog.Println(msg)

	// Send to alert channel
	select {
	case l.alertChan <- event:
	default:
	}

	return event
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.consoleLog.Printf("[INFO] "+format, v...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	l.consoleLog.Printf("[WARNING] "+format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.consoleLog.Printf("[ERROR] "+format, v...)
}

// writeEvent writes an event to the attack log file
func (l *Logger) writeEvent(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		l.Error("Failed to marshal event: %v", err)
		return
	}

	if _, err := l.attackLog.Write(append(data, '\n')); err != nil {
		l.Error("Failed to write event: %v", err)
	}
}

// GetAlertChannel returns the channel for alerts
func (l *Logger) GetAlertChannel() <-chan Event {
	return l.alertChan
}

// Close closes the logger
func (l *Logger) Close() error {
	close(l.alertChan)
	return l.attackLog.Close()
}
