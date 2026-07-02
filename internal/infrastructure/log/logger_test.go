package log

import (
	"testing"
)

func TestInitLoggerDevelopment(t *testing.T) {
	logger, err := InitLogger("development")
	if err != nil {
		t.Fatalf("InitLogger(development) failed: %v", err)
	}
	if logger == nil {
		t.Fatal("InitLogger returned nil logger")
	}
	logger.Sync() //nolint:errcheck
}

func TestInitLoggerProduction(t *testing.T) {
	logger, err := InitLogger("production")
	if err != nil {
		t.Fatalf("InitLogger(production) failed: %v", err)
	}
	if logger == nil {
		t.Fatal("InitLogger returned nil logger")
	}
	logger.Sync() //nolint:errcheck
}

func TestInitLoggerCustomEnv(t *testing.T) {
	logger, err := InitLogger("staging")
	if err != nil {
		t.Fatalf("InitLogger(staging) failed: %v", err)
	}
	if logger == nil {
		t.Fatal("InitLogger returned nil logger")
	}
	// Staging should behave like production (JSON output)
	logger.Sync() //nolint:errcheck
}
