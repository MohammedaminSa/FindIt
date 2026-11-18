package config

import (
	"os"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear any existing env variables to test fallbacks
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("SERVER_PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.DBHost == "" {
		t.Errorf("Expected DBHost to have default value, got empty")
	}

	if cfg.ServerPort == "" {
		t.Errorf("Expected ServerPort to have default value, got empty")
	}

	if cfg.UploadDir != "uploads" {
		t.Errorf("Expected UploadDir to default to 'uploads', got '%s'", cfg.UploadDir)
	}

	if cfg.MaxUploadSize != 5*1024*1024 {
		t.Errorf("Expected MaxUploadSize to default to 5MB, got %d", cfg.MaxUploadSize)
	}

	expectedDSNPrefix := "postgres://"
	if len(cfg.DatabaseURL()) < len(expectedDSNPrefix) || cfg.DatabaseURL()[:len(expectedDSNPrefix)] != expectedDSNPrefix {
		t.Errorf("DatabaseURL() should start with %s, got %s", expectedDSNPrefix, cfg.DatabaseURL())
	}
}
