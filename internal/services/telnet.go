package services

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/rgnote/TheneKunda/internal/config"
	"github.com/rgnote/TheneKunda/internal/logger"
)

// TelnetService implements a Telnet honeypot
type TelnetService struct {
	config   config.ServiceConfig
	logger   *logger.Logger
	listener net.Listener
	stopChan chan struct{}
}

// NewTelnetService creates a new Telnet honeypot service
func NewTelnetService(cfg config.ServiceConfig, log *logger.Logger) *TelnetService {
	return &TelnetService{
		config:   cfg,
		logger:   log,
		stopChan: make(chan struct{}),
	}
}

// Start starts the Telnet honeypot service
func (t *TelnetService) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", t.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	t.listener = listener
	t.logger.Info("Telnet honeypot listening on %s", addr)

	go t.acceptConnections()

	return nil
}

// Stop stops the Telnet honeypot service
func (t *TelnetService) Stop() error {
	close(t.stopChan)
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

func (t *TelnetService) acceptConnections() {
	for {
		select {
		case <-t.stopChan:
			return
		default:
		}

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.stopChan:
				return
			default:
				t.logger.Error("Failed to accept connection: %v", err)
				continue
			}
		}

		go t.handleConnection(conn)
	}
}

func (t *TelnetService) handleConnection(conn net.Conn) {
	defer conn.Close()

	ip, port := getIPPort(conn.RemoteAddr().String())

	// Send banner
	if t.config.Banner != "" {
		conn.Write([]byte(t.config.Banner + "\r\n\r\n"))
	}

	// Send login prompt
	conn.Write([]byte("login: "))

	scanner := bufio.NewScanner(conn)
	var username, password string
	state := "username"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch state {
		case "username":
			username = line
			conn.Write([]byte("Password: "))
			state = "password"

		case "password":
			password = line

			// Log the credentials
			t.logger.LogConnection("Telnet", ip, port, username, password)

			// Send fake error and ask again
			conn.Write([]byte("\r\nLogin incorrect\r\nlogin: "))
			username = ""
			password = ""
			state = "username"

		case "command":
			if line != "" {
				t.logger.LogCommand("Telnet", ip, line)

				// Send fake response
				conn.Write([]byte("bash: " + line + ": command not found\r\n$ "))
			} else {
				conn.Write([]byte("$ "))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.logger.Error("Error reading from connection: %v", err)
	}
}
