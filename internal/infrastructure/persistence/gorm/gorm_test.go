package gorm

import (
	"testing"
	"time"

	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/config"
)

func TestDSNFormat(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	got := cfg.DSN()
	want := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDSNFormatWithAllFields(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:            "db.example.com",
		Port:            12345,
		User:            "admin",
		Password:        "s3cret!",
		DBName:          "mydb",
		SSLMode:         "require",
		MaxOpenConns:    50,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}

	got := cfg.DSN()
	want := "host=db.example.com port=12345 user=admin password=s3cret! dbname=mydb sslmode=require"
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}
