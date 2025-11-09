#!/bin/bash

# TheneKunda Daemon Installation Script

set -e

SERVICE_NAME="thenekunda.service"
SERVICE_FILE="$(pwd)/$SERVICE_NAME"
SYSTEMD_DIR="/etc/systemd/system"

echo "Installing TheneKunda as a systemd service..."

# Check if running with sufficient privileges
if [ "$EUID" -ne 0 ]; then
    echo "This script requires sudo privileges to install the systemd service."
    echo "Please run with: sudo ./install-daemon.sh"
    exit 1
fi

# Check if service file exists
if [ ! -f "$SERVICE_FILE" ]; then
    echo "Error: $SERVICE_FILE not found!"
    exit 1
fi

# Copy service file to systemd directory
echo "Copying service file to $SYSTEMD_DIR..."
cp "$SERVICE_FILE" "$SYSTEMD_DIR/$SERVICE_NAME"

# Reload systemd daemon
echo "Reloading systemd daemon..."
systemctl daemon-reload

# Enable the service to start on boot
echo "Enabling service to start on boot..."
systemctl enable $SERVICE_NAME

echo ""
echo "Installation complete!"
echo ""
echo "Available commands:"
echo "  Start service:   sudo systemctl start $SERVICE_NAME"
echo "  Stop service:    sudo systemctl stop $SERVICE_NAME"
echo "  Restart service: sudo systemctl restart $SERVICE_NAME"
echo "  Check status:    sudo systemctl status $SERVICE_NAME"
echo "  View logs:       sudo journalctl -u $SERVICE_NAME -f"
echo "  Disable service: sudo systemctl disable $SERVICE_NAME"
echo ""
