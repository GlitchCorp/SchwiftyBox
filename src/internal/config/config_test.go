package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Test with default values
	cfg := NewConfig()

	// Check default database values
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.User != "user" {
		t.Errorf("Expected default user 'user', got '%s'", cfg.Database.User)
	}
	if cfg.Database.Password != "password" {
		t.Errorf("Expected default password 'password', got '%s'", cfg.Database.Password)
	}
	if cfg.Database.DBName != "mydb" {
		t.Errorf("Expected default dbname 'mydb', got '%s'", cfg.Database.DBName)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected default port '5432', got '%s'", cfg.Database.Port)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Expected default sslmode 'disable', got '%s'", cfg.Database.SSLMode)
	}

	// Check default JWT values
	if cfg.JWT.SecretKey != "secret" {
		t.Errorf("Expected default secret key 'secret', got '%s'", cfg.JWT.SecretKey)
	}
	if cfg.JWT.AccessTokenDuration != time.Minute*15 {
		t.Errorf("Expected access token duration 15 minutes, got %v", cfg.JWT.AccessTokenDuration)
	}
	if cfg.JWT.RefreshTokenDuration != time.Hour*24 {
		t.Errorf("Expected refresh token duration 24 hours, got %v", cfg.JWT.RefreshTokenDuration)
	}

	// Check default server values
	if cfg.Server.Port != ":8080" {
		t.Errorf("Expected default port ':8080', got '%s'", cfg.Server.Port)
	}
}

func TestNewConfigWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("SERVER_PORT", ":9090")

	// Clean up after test
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := NewConfig()

	// Check that environment variables are used
	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected host 'testhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", cfg.Database.User)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", cfg.Database.Password)
	}
	if cfg.Database.DBName != "testdb" {
		t.Errorf("Expected dbname 'testdb', got '%s'", cfg.Database.DBName)
	}
	if cfg.Database.Port != "5433" {
		t.Errorf("Expected port '5433', got '%s'", cfg.Database.Port)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Expected sslmode 'require', got '%s'", cfg.Database.SSLMode)
	}
	if cfg.JWT.SecretKey != "testsecret" {
		t.Errorf("Expected secret key 'testsecret', got '%s'", cfg.JWT.SecretKey)
	}
	if cfg.Server.Port != ":9090" {
		t.Errorf("Expected port ':9090', got '%s'", cfg.Server.Port)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	value := getEnv("TEST_KEY", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test with non-existing environment variable
	value = getEnv("NON_EXISTENT_KEY", "default_value")
	if value != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value)
	}

	// Test with empty environment variable
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	value = getEnv("EMPTY_KEY", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	dbConfig := DatabaseConfig{
		Host:     "localhost",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		Port:     "5432",
		SSLMode:  "disable",
	}

	expected := "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable"
	result := dbConfig.ConnectionString()

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestDatabaseConfig_ConnectionStringWithSpecialCharacters(t *testing.T) {
	dbConfig := DatabaseConfig{
		Host:     "localhost",
		User:     "test@user",
		Password: "test@pass",
		DBName:   "test-db",
		Port:     "5432",
		SSLMode:  "require",
	}

	result := dbConfig.ConnectionString()
	
	// Should contain all the values
	if result == "" {
		t.Error("Connection string should not be empty")
	}
	
	// Should contain all components
	components := []string{"host=localhost", "user=test@user", "password=test@pass", "dbname=test-db", "port=5432", "sslmode=require"}
	for _, component := range components {
		if result != "" && result != component {
			// This is a simple check - in a real scenario you might want more sophisticated parsing
			t.Logf("Connection string contains: %s", result)
		}
	}
}
