#!/bin/bash

# TheneKunda Daemon Uninstallation Script

set -e

SERVICE_NAME="thenekunda.service"
SYSTEMD_DIR="/etc/systemd/system"

echo "Uninstalling TheneKunda systemd service..."

# Check if running with sufficient privileges
if [ "$EUID" -ne 0 ]; then
    echo "This script requires sudo privileges to uninstall the systemd service."
    echo "Please run with: sudo ./uninstall-daemon.sh"
    exit 1
fi

# Stop the service if it's running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "Stopping service..."
    systemctl stop $SERVICE_NAME
fi

# Disable the service
if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
    echo "Disabling service..."
    systemctl disable $SERVICE_NAME
fi

# Remove service file
if [ -f "$SYSTEMD_DIR/$SERVICE_NAME" ]; then
    echo "Removing service file..."
    rm "$SYSTEMD_DIR/$SERVICE_NAME"
fi

# Reload systemd daemon
echo "Reloading systemd daemon..."
systemctl daemon-reload

echo ""
echo "Uninstallation complete!"
echo ""
