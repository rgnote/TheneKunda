# TheneKunda

**A purpose-built honeypot for home networks**

TheneKunda is a defensive security tool designed to detect and log unauthorized access attempts on home networks. It emulates common network services (SSH, Telnet, HTTP) to attract and record malicious activity, providing valuable threat intelligence.

## Features

- **Multiple Service Honeypots**
  - SSH honeypot (emulates OpenSSH)
  - Telnet honeypot (emulates basic telnet service)
  - HTTP honeypot (emulates Apache web server)

- **Comprehensive Logging**
  - Structured JSON logs for all events
  - Connection attempts with credentials
  - Command execution attempts
  - HTTP requests with headers and payloads

- **Alert System**
  - Real-time console alerts
  - Webhook integration (Slack, Discord, custom)
  - Configurable alert levels

- **Easy Deployment**
  - Docker support with docker-compose
  - Minimal configuration required
  - Low resource footprint

## Quick Start

### Using Docker (Recommended)

```bash
# Generate example configuration
docker run --rm thenekundа-honeypot -generate-config > config.example.yaml

# Copy and customize config
cp config.example.yaml config.yaml

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f
```

### Build from Source

**Prerequisites:**
- Go 1.21 or higher

```bash
# Clone the repository
git clone https://github.com/rgnote/TheneKunda.git
cd TheneKunda

# Download dependencies
go mod download

# Build
make build

# Generate example config
./thenekundа -generate-config

# Run
./thenekundа -config config.yaml
```

## Configuration

Create a `config.yaml` file to customize the honeypot:

```yaml
general:
  log_dir: logs

services:
  ssh:
    enabled: true
    port: 2222
    banner: "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"

  telnet:
    enabled: true
    port: 2323
    banner: "Ubuntu 20.04 LTS"

  http:
    enabled: true
    port: 8080
    server_banner: "Apache/2.4.41 (Ubuntu)"

alerts:
  webhook:
    enabled: false
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  console:
    enabled: true
```

### Port Configuration

**Important:** For security reasons, the default ports are set to non-standard values:
- SSH: 2222 (standard: 22)
- Telnet: 2323 (standard: 23)
- HTTP: 8080 (standard: 80)

To use standard ports, you'll need root/admin privileges:

```bash
# Linux (using setcap)
sudo setcap 'cap_net_bind_service=+ep' ./thenekundа

# Or run as root (not recommended for production)
sudo ./thenekundа -config config.yaml
```

For Docker, map the container ports to standard ports:

```yaml
ports:
  - "22:2222"   # Map host port 22 to container port 2222
  - "23:2323"   # Map host port 23 to container port 2323
  - "80:8080"   # Map host port 80 to container port 8080
```

## Deployment Strategies

### Home Network Deployment

1. **Dedicated Device**: Run on a Raspberry Pi or old computer
2. **Isolated Network**: Place on a separate VLAN if possible
3. **Port Forwarding**: Forward standard ports (22, 23, 80) from your router to the honeypot

### Security Considerations

- **Isolation**: Keep the honeypot isolated from your main network
- **Monitoring**: Regularly review logs for suspicious activity
- **Updates**: Keep the honeypot software updated
- **No Sensitive Data**: Never store sensitive data on the honeypot system
- **Legal**: Ensure honeypot deployment complies with local laws

## Log Analysis

Logs are stored in two formats:

### 1. Console Logs (`logs/honeypot.log`)
Human-readable format for quick review.

### 2. Attack Logs (`logs/attacks.jsonl`)
Structured JSON format for analysis:

```json
{
  "timestamp": "2025-11-06T12:34:56Z",
  "event_type": "connection",
  "service": "SSH",
  "source_ip": "192.168.1.100",
  "source_port": 54321,
  "username": "admin",
  "password": "password123"
}
```

### Analyzing Logs with jq

```bash
# Count attacks by IP
cat logs/attacks.jsonl | jq -r '.source_ip' | sort | uniq -c | sort -rn

# Find most common usernames
cat logs/attacks.jsonl | jq -r 'select(.username != null) | .username' | sort | uniq -c | sort -rn

# Find most common passwords
cat logs/attacks.jsonl | jq -r 'select(.password != null) | .password' | sort | uniq -c | sort -rn

# Show all commands attempted
cat logs/attacks.jsonl | jq -r 'select(.command != null) | .command'

# Filter by service
cat logs/attacks.jsonl | jq 'select(.service == "SSH")'
```

## Webhook Integration

### Slack

1. Create a Slack webhook: https://api.slack.com/messaging/webhooks
2. Add the URL to your config:

```yaml
alerts:
  webhook:
    enabled: true
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

### Discord

1. Create a Discord webhook in your server settings
2. Add `/slack` to the end of the URL
3. Use it in the config

### Custom Webhooks

TheneKunda sends events as JSON POST requests:

```json
{
  "timestamp": "2025-11-06T12:34:56Z",
  "event_type": "connection",
  "service": "SSH",
  "source_ip": "192.168.1.100",
  "username": "root",
  "password": "toor"
}
```

## Development

### Building

```bash
make build          # Build binary
make test           # Run tests
make fmt            # Format code
make vet            # Run go vet
make docker-build   # Build Docker image
```

### Project Structure

```
TheneKunda/
├── cmd/
│   └── thenekundа/      # Main application
│       └── main.go
├── internal/
│   ├── alerts/         # Alert system
│   ├── config/         # Configuration management
│   ├── logger/         # Logging system
│   └── services/       # Honeypot services
│       ├── ssh.go
│       ├── telnet.go
│       ├── http.go
│       └── utils.go
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## Troubleshooting

### Permission Denied on Low Ports

If you get "permission denied" errors when binding to ports < 1024:

```bash
# Option 1: Use setcap (Linux)
sudo setcap 'cap_net_bind_service=+ep' ./thenekundа

# Option 2: Run as root (not recommended)
sudo ./thenekundа -config config.yaml

# Option 3: Use higher ports and port forwarding
# Configure services to use ports > 1024 and use iptables:
sudo iptables -t nat -A PREROUTING -p tcp --dport 22 -j REDIRECT --to-port 2222
```

### No Attacks Detected

- Ensure the honeypot is accessible from the internet
- Check firewall rules on your router and host
- Verify port forwarding is configured correctly
- Some attacks may take days or weeks to appear

### High CPU Usage

- Check if you're under active attack (review logs)
- Consider rate limiting at the firewall level
- Restart the honeypot if necessary

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.

## License

This project is provided as-is for educational and defensive security purposes.

## Disclaimer

This software is intended for defensive security and educational purposes only. Users are responsible for ensuring their use complies with applicable laws and regulations. The authors are not responsible for any misuse of this software.

## Acknowledgments

- Built with Go and love for cybersecurity
- Inspired by the need for accessible home network security tools

---

**Stay safe, stay secure! 🍯🛡️**
