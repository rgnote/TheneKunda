package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/rgnote/TheneKunda/internal/config"
	"github.com/rgnote/TheneKunda/internal/logger"
	"golang.org/x/crypto/ssh"
)

// SSHService implements an SSH honeypot
type SSHService struct {
	config   config.ServiceConfig
	logger   *logger.Logger
	listener net.Listener
	sshConf  *ssh.ServerConfig
	stopChan chan struct{}
}

// NewSSHService creates a new SSH honeypot service
func NewSSHService(cfg config.ServiceConfig, log *logger.Logger) (*SSHService, error) {
	s := &SSHService{
		config:   cfg,
		logger:   log,
		stopChan: make(chan struct{}),
	}

	// Setup SSH server config
	s.sshConf = &ssh.ServerConfig{
		PasswordCallback: s.passwordCallback,
		PublicKeyCallback: s.publicKeyCallback,
		ServerVersion:    cfg.Banner,
	}

	// Generate or load host key
	privateKey, err := s.loadOrGenerateHostKey(cfg.HostKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load host key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %v", err)
	}

	s.sshConf.AddHostKey(signer)

	return s, nil
}

// Start starts the SSH honeypot service
func (s *SSHService) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	s.listener = listener
	s.logger.Info("SSH honeypot listening on %s", addr)

	go s.acceptConnections()

	return nil
}

// Stop stops the SSH honeypot service
func (s *SSHService) Stop() error {
	close(s.stopChan)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *SSHService) acceptConnections() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return
			default:
				s.logger.Error("Failed to accept connection: %v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *SSHService) handleConnection(netConn net.Conn) {
	defer netConn.Close()

	ip, port := getIPPort(netConn.RemoteAddr().String())

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConf)
	if err != nil {
		// Connection attempt logged in password callback
		return
	}
	defer sshConn.Close()

	s.logger.Info("SSH connection established from %s:%d", ip, port)

	// Discard all global requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		go s.handleChannel(newChannel, ip)
	}
}

func (s *SSHService) handleChannel(newChannel ssh.NewChannel, ip string) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		s.logger.Error("Failed to accept channel: %v", err)
		return
	}
	defer channel.Close()

	// Handle requests
	for req := range requests {
		switch req.Type {
		case "shell", "exec":
			if req.Type == "exec" {
				var payload struct{ Command string }
				if err := ssh.Unmarshal(req.Payload, &payload); err == nil {
					s.logger.LogCommand("SSH", ip, payload.Command)
				}
			}

			if req.WantReply {
				req.Reply(true, nil)
			}

			// Send fake shell prompt
			channel.Write([]byte("$ "))

			// Read commands
			buf := make([]byte, 1024)
			for {
				n, err := channel.Read(buf)
				if err != nil {
					if err != io.EOF {
						s.logger.Error("Error reading from channel: %v", err)
					}
					return
				}

				command := strings.TrimSpace(string(buf[:n]))
				if command != "" && command != "\r" && command != "\n" {
					s.logger.LogCommand("SSH", ip, command)

					// Send fake response
					channel.Write([]byte("command not found\r\n$ "))
				}
			}

		case "pty-req":
			if req.WantReply {
				req.Reply(true, nil)
			}

		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (s *SSHService) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	ip, port := getIPPort(conn.RemoteAddr().String())

	s.logger.LogConnection("SSH", ip, port, conn.User(), string(password))

	// Always reject the password but log it
	return nil, fmt.Errorf("password rejected")
}

func (s *SSHService) publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	ip, _ := getIPPort(conn.RemoteAddr().String())

	keyType := key.Type()
	fingerprint := ssh.FingerprintSHA256(key)

	extraData := map[string]interface{}{
		"key_type":    keyType,
		"fingerprint": fingerprint,
	}

	s.logger.LogEvent("SSH", "pubkey_auth", ip, "Public key authentication attempt", extraData)

	// Always reject but log
	return nil, fmt.Errorf("public key rejected")
}

func (s *SSHService) loadOrGenerateHostKey(keyPath string) (*rsa.PrivateKey, error) {
	// If keyPath is provided, try to load it
	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err == nil {
			block, _ := pem.Decode(keyData)
			if block != nil {
				return x509.ParsePKCS1PrivateKey(block.Bytes)
			}
		}
	}

	// Generate new key
	s.logger.Info("Generating new SSH host key...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Save key to logs directory
	keyDir := filepath.Join("logs", "ssh_host_key")
	os.MkdirAll(filepath.Dir(keyDir), 0755)

	keyFile, err := os.OpenFile(keyDir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err == nil {
		defer keyFile.Close()
		pem.Encode(keyFile, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
		s.logger.Info("SSH host key saved to %s", keyDir)
	}

	return privateKey, nil
}
