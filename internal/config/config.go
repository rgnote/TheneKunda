package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the honeypot configuration
type Config struct {
	General  GeneralConfig            `yaml:"general"`
	Services map[string]ServiceConfig `yaml:"services"`
	Alerts   AlertsConfig             `yaml:"alerts"`
}

// GeneralConfig contains general settings
type GeneralConfig struct {
	LogDir string `yaml:"log_dir"`
}

// ServiceConfig contains service-specific settings
type ServiceConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Port         int    `yaml:"port"`
	Banner       string `yaml:"banner,omitempty"`
	ServerBanner string `yaml:"server_banner,omitempty"`
	HostKey      string `yaml:"host_key,omitempty"`
}

// AlertsConfig contains alert settings
type AlertsConfig struct {
	Webhook WebhookConfig `yaml:"webhook"`
	Console ConsoleConfig `yaml:"console"`
}

// WebhookConfig contains webhook alert settings
type WebhookConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// ConsoleConfig contains console alert settings
type ConsoleConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			LogDir: "logs",
		},
		Services: map[string]ServiceConfig{
			"ssh": {
				Enabled: true,
				Port:    2223,
				Banner:  "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
			},
			"telnet": {
				Enabled: true,
				Port:    2323,
				Banner:  "Ubuntu 20.04 LTS",
			},
			"http": {
				Enabled:      true,
				Port:         8089,
				ServerBanner: "Apache/2.4.41 (Ubuntu)",
			},
		},
		Alerts: AlertsConfig{
			Webhook: WebhookConfig{
				Enabled: false,
				URL:     "",
			},
			Console: ConsoleConfig{
				Enabled: true,
			},
		},
	}
}

// Load loads configuration from a file
func Load(filename string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return cfg, nil
}

// SaveExample saves an example configuration file
func SaveExample(filename string) error {
	cfg := DefaultConfig()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
