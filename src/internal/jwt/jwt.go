package jwt

import (
	"log"
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
)

// Module provides JWT service dependency injection
var Module = fx.Module("jwt",
	fx.Provide(NewJWTService),
)

// Service handles JWT operations
type Service struct {
	config *config.JWTConfig
}

// TokenResponse represents the response containing tokens
type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.Config) *Service {
	return &Service{
		config: &cfg.JWT,
	}
}

// GenerateToken generates a JWT token with the specified email and expiry
func (s *Service) GenerateToken(email string, expiry time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(expiry).Unix(),
	})
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateAccessToken generates an access token
func (s *Service) GenerateAccessToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      email,
		"token_type": "access",
		"exp":        time.Now().Add(s.config.AccessTokenDuration).Unix(),
	})
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      email,
		"token_type": "refresh",
		"exp":        time.Now().Add(s.config.RefreshTokenDuration).Unix(),
	})
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateTokenPair generates both access and refresh tokens
func (s *Service) GenerateTokenPair(email string) (*TokenResponse, error) {
	log.Printf("GenerateTokenPair called for email: %s", email)

	accessToken, err := s.GenerateAccessToken(email)
	if err != nil {
		log.Printf("Failed to generate access token: %v", err)
		return nil, err
	}
	log.Printf("Access token generated successfully")

	refreshToken, err := s.GenerateRefreshToken(email)
	if err != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return nil, err
	}
	log.Printf("Refresh token generated successfully")

	return &TokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates a JWT token and returns the email claim
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	// Check if this is an access token (for backward compatibility, also accept tokens without token_type)
	tokenType, ok := claims["token_type"].(string)
	if ok && tokenType != "access" {
		return "", jwt.ErrInvalidKey
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return email, nil
}

// ValidateRefreshToken validates a refresh token specifically
func (s *Service) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	// Check if this is a refresh token (for backward compatibility, also accept tokens without token_type)
	tokenType, ok := claims["token_type"].(string)
	if ok && tokenType != "refresh" {
		return "", jwt.ErrInvalidKey
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return email, nil
}

// GetAccessTokenDuration returns the access token duration
func (s *Service) GetAccessTokenDuration() time.Duration {
	return s.config.AccessTokenDuration
}
