package services

import (
	"fmt"
	"strings"
)

// getIPPort extracts IP and port from an address string
func getIPPort(addr string) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) >= 2 {
		ip := strings.Join(parts[:len(parts)-1], ":")
		ip = strings.Trim(ip, "[]") // Remove brackets for IPv6

		var port int
		fmt.Sscanf(parts[len(parts)-1], "%d", &port)

		return ip, port
	}
	return addr, 0
}
