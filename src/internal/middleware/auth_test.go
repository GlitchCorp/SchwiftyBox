package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthTest(t *testing.T) (*gin.Engine, *jwt.Service) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:            "test-secret-key",
			AccessTokenDuration:  24 * time.Hour, // 24 hours
			RefreshTokenDuration: 24 * time.Hour, // 24 hours
		},
	}

	jwtService := jwt.NewJWTService(cfg)
	return engine, jwtService
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	engine, jwtService := setupAuthTest(t)

	// Add middleware to engine
	engine.Use(AuthMiddleware(jwtService))

	// Add test endpoint
	engine.GET("/test", func(c *gin.Context) {
		userEmail, exists := c.Get("user_email")
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", userEmail)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate valid token using refresh token duration
	token, err := jwtService.GenerateToken("test@example.com", 24*time.Hour)
	assert.NoError(t, err)

	// Validate token manually to debug
	email, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	engine, jwtService := setupAuthTest(t)

	// Add middleware to engine
	engine.Use(AuthMiddleware(jwtService))

	// Add test endpoint
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	engine, jwtService := setupAuthTest(t)

	// Add middleware to engine
	engine.Use(AuthMiddleware(jwtService))

	// Add test endpoint
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid header format
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	engine, jwtService := setupAuthTest(t)

	// Add middleware to engine
	engine.Use(AuthMiddleware(jwtService))

	// Add test endpoint
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	engine, jwtService := setupAuthTest(t)

	// Add middleware to engine
	engine.Use(AuthMiddleware(jwtService))

	// Add test endpoint
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with empty token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
