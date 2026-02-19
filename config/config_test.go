package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.DefaultPorts) == 0 {
		t.Error("DefaultConfig().DefaultPorts is empty")
	}
	if cfg.CheckInterval <= 0 {
		t.Error("DefaultConfig().CheckInterval should be positive")
	}
	if cfg.Timeout <= 0 {
		t.Error("DefaultConfig().Timeout should be positive")
	}
	if cfg.MaxConcurrency <= 0 {
		t.Error("DefaultConfig().MaxConcurrency should be positive")
	}
}

func TestConfig_AddEndpoint(t *testing.T) {
	cfg := DefaultConfig()

	cfg.AddEndpoint("http://localhost:3000", "API")
	if len(cfg.Endpoints) != 1 {
		t.Errorf("Endpoints length = %d, want 1", len(cfg.Endpoints))
	}

	cfg.AddEndpoint("http://localhost:3000", "API Again")
	if len(cfg.Endpoints) != 1 {
		t.Errorf("Adding duplicate should not increase count, got %d", len(cfg.Endpoints))
	}

	cfg.AddEndpoint("http://localhost:8080", "Backend")
	if len(cfg.Endpoints) != 2 {
		t.Errorf("Endpoints length = %d, want 2", len(cfg.Endpoints))
	}
}

func TestConfig_RemoveEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddEndpoint("http://localhost:3000", "API")
	cfg.AddEndpoint("http://localhost:8080", "Backend")

	cfg.RemoveEndpoint("http://localhost:3000")
	if len(cfg.Endpoints) != 1 {
		t.Errorf("Endpoints length = %d, want 1", len(cfg.Endpoints))
	}
	if cfg.Endpoints[0].URL != "http://localhost:8080" {
		t.Errorf("Remaining endpoint = %s, want http://localhost:8080", cfg.Endpoints[0].URL)
	}

	cfg.RemoveEndpoint("http://nonexistent:9999")
	if len(cfg.Endpoints) != 1 {
		t.Errorf("Removing nonexistent should not change count, got %d", len(cfg.Endpoints))
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".localpulse.json")

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	})

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)

	cfg := DefaultConfig()
	cfg.AddEndpoint("http://localhost:3000", "API")
	cfg.AddEndpoint("http://localhost:8080", "Backend")

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file not created at %s", configFile)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Endpoints) != 2 {
		t.Errorf("Loaded endpoints = %d, want 2", len(loaded.Endpoints))
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	})

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() with nonexistent file should not error, got %v", err)
	}
	if cfg == nil {
		t.Error("Load() should return default config for nonexistent file")
	}
}

func TestConfig_GetEndpoints(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddEndpoint("http://localhost:3000", "API")

	endpoints := cfg.GetEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("GetEndpoints() length = %d, want 1", len(endpoints))
	}

	endpoints[0].URL = "modified"

	original := cfg.GetEndpoints()
	if original[0].URL == "modified" {
		t.Error("GetEndpoints() should return a copy, not a reference")
	}
}
