package jwt

import (
	"testing"
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func createTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			SecretKey:            "test-secret-key",
			AccessTokenDuration:  time.Minute * 15,
			RefreshTokenDuration: time.Hour * 24,
		},
	}
}

func TestNewJWTService(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	if service.config == nil {
		t.Error("JWT service config should not be nil")
	}
	if service.config.SecretKey != "test-secret-key" {
		t.Errorf("Expected secret key 'test-secret-key', got '%s'", service.config.SecretKey)
	}
}

func TestGenerateToken(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	email := "test@example.com"
	expiry := time.Minute * 10

	token, err := service.GenerateToken(email, expiry)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.config.SecretKey), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token should be valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims from token")
	}

	if claims["email"] != email {
		t.Errorf("Expected email '%s', got '%s'", email, claims["email"])
	}

	// Check expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("Failed to extract expiration from token")
	}

	expectedExp := time.Now().Add(expiry).Unix()
	if int64(exp) < expectedExp-5 || int64(exp) > expectedExp+5 {
		t.Errorf("Token expiration should be around %d, got %d", expectedExp, int64(exp))
	}
}

func TestGenerateTokenPair(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(email)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.Token == "" {
		t.Error("Access token should not be empty")
	}
	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	if tokenPair.Token == tokenPair.RefreshToken {
		t.Error("Access token and refresh token should be different")
	}

	// Validate both tokens
	accessEmail, err := service.ValidateToken(tokenPair.Token)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}
	if accessEmail != email {
		t.Errorf("Expected email '%s', got '%s'", email, accessEmail)
	}

	refreshEmail, err := service.ValidateToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}
	if refreshEmail != email {
		t.Errorf("Expected email '%s', got '%s'", email, refreshEmail)
	}
}

func TestValidateToken(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	email := "test@example.com"
	token, err := service.GenerateToken(email, time.Minute*10)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test valid token
	validatedEmail, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate valid token: %v", err)
	}
	if validatedEmail != email {
		t.Errorf("Expected email '%s', got '%s'", email, validatedEmail)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	// Test invalid token
	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Should return error for invalid token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	email := "test@example.com"
	// Generate token with very short expiration
	token, err := service.GenerateToken(email, time.Millisecond*1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(time.Millisecond * 10)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Should return error for expired token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	email := "test@example.com"
	token, err := service.GenerateToken(email, time.Minute*10)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create service with different secret
	wrongCfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey: "different-secret-key",
		},
	}
	wrongService := NewJWTService(wrongCfg)

	_, err = wrongService.ValidateToken(token)
	if err == nil {
		t.Error("Should return error for token signed with different secret")
	}
}

func TestGetAccessTokenDuration(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	expectedDuration := time.Minute * 15
	actualDuration := service.GetAccessTokenDuration()

	if actualDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, actualDuration)
	}
}

func TestGenerateToken_EmptyEmail(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	token, err := service.GenerateToken("", time.Minute*10)
	if err != nil {
		t.Fatalf("Should not fail with empty email: %v", err)
	}

	if token == "" {
		t.Error("Should generate token even with empty email")
	}

	// Validate the token
	email, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token with empty email: %v", err)
	}
	if email != "" {
		t.Errorf("Expected empty email, got '%s'", email)
	}
}

func TestGenerateTokenPair_EmptyEmail(t *testing.T) {
	cfg := createTestConfig()
	service := NewJWTService(cfg)

	tokenPair, err := service.GenerateTokenPair("")
	if err != nil {
		t.Fatalf("Should not fail with empty email: %v", err)
	}

	if tokenPair.Token == "" || tokenPair.RefreshToken == "" {
		t.Error("Should generate both tokens even with empty email")
	}
}
