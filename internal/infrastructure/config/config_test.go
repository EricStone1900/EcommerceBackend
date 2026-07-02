package config

import (
	"fmt"
	"testing"
)

func TestDSNFormat(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "productapp",
		Password: "secret",
		DBName:   "productsystem",
		SSLMode:  "disable",
	}

	got := cfg.DSN()
	want := "host=localhost port=5432 user=productapp password=secret dbname=productsystem sslmode=disable"
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDSNEmptyPassword(t *testing.T) {
	cfg := DatabaseConfig{
		Host:    "localhost",
		Port:    5432,
		User:    "u",
		DBName:  "d",
		SSLMode: "disable",
	}

	got := cfg.DSN()
	if !contains(got, "password=") {
		t.Errorf("DSN should include empty password=, got: %s", got)
	}
}

func TestValidateJWTSecretDefault(t *testing.T) {
	cfg := Config{
		JWT: JWTConfig{
			Secret: "your_jwt_secret_please_change",
		},
		Database: DatabaseConfig{
			Password: "real_password",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("validate() should fail with default JWT secret")
	}
}

func TestValidateDBPasswordDefault(t *testing.T) {
	cfg := Config{
		JWT: JWTConfig{
			Secret: "my_real_secret",
		},
		Database: DatabaseConfig{
			Password: "your_password_here",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("validate() should fail with default DB password")
	}
}

func TestValidateSuccess(t *testing.T) {
	cfg := Config{
		JWT: JWTConfig{
			Secret: "my_real_secret",
		},
		Database: DatabaseConfig{
			Password: "my_real_password",
		},
	}

	err := cfg.validate()
	if err != nil {
		t.Errorf("validate() should pass with real values, got: %v", err)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadConfig should fail with nonexistent file path")
	}
}

func ExampleDatabaseConfig_DSN() {
	cfg := DatabaseConfig{
		Host:    "localhost",
		Port:    5432,
		User:    "productapp",
		DBName:  "productsystem",
		SSLMode: "disable",
	}
	fmt.Println(cfg.DSN())
	// Output: host=localhost port=5432 user=productapp password= dbname=productsystem sslmode=disable
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
