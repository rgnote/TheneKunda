package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rgnote/TheneKunda/internal/alerts"
	"github.com/rgnote/TheneKunda/internal/config"
	"github.com/rgnote/TheneKunda/internal/logger"
	"github.com/rgnote/TheneKunda/internal/services"
)

const banner = `
 _____ _                   _  __               _
|_   _| |__   ___ _ __   __| |/ /   _ _ __   __| | __ _
  | | | '_ \ / _ \ '_ \ / _' | |  | | | '_ \ / _' |/ _' |
  | | | | | |  __/ | | |  __/ |__| |_| | | | | (_| | (_| |
  |_| |_| |_|\___|_| |_|\___|_____\__,_|_| |_|\__,_|\__,_|

  Home Network Honeypot - v1.0.0
  Defensive Security Tool for Threat Detection
`

func main() {
	configFile := flag.String("config", "", "Path to configuration file")
	generateConfig := flag.Bool("generate-config", false, "Generate example configuration file")
	flag.Parse()

	// Generate example config if requested
	if *generateConfig {
		if err := config.SaveExample("config.example.yaml"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Example configuration saved to config.example.yaml")
		return
	}

	fmt.Println(banner)

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.General.LogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("TheneKunda honeypot starting...")

	// Initialize alert manager
	alertMgr := alerts.New(cfg.Alerts, log)
	alertMgr.Start(log.GetAlertChannel())

	// Track active services
	var activeServices []interface{ Stop() error }

	// Start SSH service
	if cfg.Services["ssh"].Enabled {
		sshService, err := services.NewSSHService(cfg.Services["ssh"], log)
		if err != nil {
			log.Error("Failed to create SSH service: %v", err)
		} else if err := sshService.Start(); err != nil {
			log.Error("Failed to start SSH service: %v", err)
		} else {
			activeServices = append(activeServices, sshService)
		}
	}

	// Start Telnet service
	if cfg.Services["telnet"].Enabled {
		telnetService := services.NewTelnetService(cfg.Services["telnet"], log)
		if err := telnetService.Start(); err != nil {
			log.Error("Failed to start Telnet service: %v", err)
		} else {
			activeServices = append(activeServices, telnetService)
		}
	}

	// Start HTTP service
	if cfg.Services["http"].Enabled {
		httpService := services.NewHTTPService(cfg.Services["http"], log)
		if err := httpService.Start(); err != nil {
			log.Error("Failed to start HTTP service: %v", err)
		} else {
			activeServices = append(activeServices, httpService)
		}
	}

	if len(activeServices) == 0 {
		log.Error("No services enabled. Exiting.")
		return
	}

	log.Info("Honeypot is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Shutdown
	log.Info("Shutting down...")
	for _, service := range activeServices {
		if err := service.Stop(); err != nil {
			log.Error("Error stopping service: %v", err)
		}
	}

	log.Info("Honeypot stopped.")
}
