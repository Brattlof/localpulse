package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type EndpointConfig struct {
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
}

type Config struct {
	mu sync.RWMutex

	Endpoints      []EndpointConfig `json:"endpoints"`
	DefaultPorts   []int            `json:"default_ports"`
	CheckInterval  int              `json:"check_interval_seconds"`
	LoadTestRPS    int              `json:"load_test_rps"`
	Timeout        int              `json:"timeout_seconds"`
	MaxConcurrency int              `json:"max_concurrency"`
	WindowSeconds  int              `json:"window_seconds"`
}

func DefaultConfig() *Config {
	return &Config{
		Endpoints:      []EndpointConfig{},
		DefaultPorts:   []int{3000, 3001, 8080, 8000, 5000, 9000, 4000, 4200, 5173},
		CheckInterval:  1,
		LoadTestRPS:    10,
		Timeout:        5,
		MaxConcurrency: 100,
		WindowSeconds:  30,
	}
}

func (c *Config) AddEndpoint(url, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ep := range c.Endpoints {
		if ep.URL == url {
			return
		}
	}

	c.Endpoints = append(c.Endpoints, EndpointConfig{
		URL:  url,
		Name: name,
	})
}

func (c *Config) RemoveEndpoint(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, ep := range c.Endpoints {
		if ep.URL == url {
			c.Endpoints = append(c.Endpoints[:i], c.Endpoints[i+1:]...)
			return
		}
	}
}

func (c *Config) GetEndpoints() []EndpointConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]EndpointConfig, len(c.Endpoints))
	copy(result, c.Endpoints)
	return result
}

func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".localpulse.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	if len(cfg.DefaultPorts) == 0 {
		cfg.DefaultPorts = DefaultConfig().DefaultPorts
	}
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 100
	}
	if cfg.WindowSeconds <= 0 {
		cfg.WindowSeconds = 30
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
